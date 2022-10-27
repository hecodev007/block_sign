package services

import (
	"bifrostsync/common/conf"
	"bifrostsync/common/log"
	"bifrostsync/common/rabbitmq"
	"bifrostsync/models/bo"
	"bifrostsync/models/po"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type PushTask struct {
	PushID   int64  `json:"push_id"`   // 通知结果id
	TxData   []byte `json:"tx_data"`   // 推送数据
	PushNum  int    `json:"push_num"`  // 推送次数
	PushTime int64  `json:"push_time"` // 推送时间
	PushUrl  string `json:"push_url"`  // url
	UserID   int64  `json:"user_id"`   // 商户id
}

type PushData struct {
	Uid  int64  `json:"uid"`
	Url  string `json:"url"`
	Data string `json:"data"`
}

type PushServer struct {
	client       *http.Client
	rmq          *rabbitmq.Server
	stop         chan struct{}
	done         chan struct{}
	watch        *WatchControl
	pushTaskList chan *PushTask
	waitPushTask map[int64]*PushTask
	publishers   map[string]*rabbitmq.Publisher //签名生产者
	cfg          conf.PushConfig
}

func NewPushServer(cfg conf.PushConfig, watch *WatchControl) *PushServer {
	s := &PushServer{
		client:       http.DefaultClient,
		watch:        watch,
		stop:         make(chan struct{}),
		done:         make(chan struct{}),
		pushTaskList: make(chan *PushTask, 10000),
		waitPushTask: make(map[int64]*PushTask),
		publishers:   make(map[string]*rabbitmq.Publisher),
		cfg:          cfg,
	}
	if cfg.Enable && cfg.Type == "rabbitmq" {
		if err := s.connectMQ(); err != nil {
			panic(err.Error())
		}
	}
	return s
}

func (s *PushServer) Stop() {
	close(s.stop)
	<-s.done
	s.disconnectMQ()
}

// 运行
func (s *PushServer) Start() {
	log.Infof("push server start ")
	var rePushTime = []int64{10, 30, 60, 180, 720, 1800, 7200}
	timer := time.NewTicker(10 * time.Second) // timer
	run := true

	go func() {
		for run {
			select {
			case <-s.stop:
				log.Infof("push server stop")
				run = false
				break
			case <-timer.C:
				{
					currTime := time.Now().Unix()
					for _, v := range s.waitPushTask {
						if currTime >= v.PushTime {
							// 添加到推送队列
							s.pushTaskList <- v //TODO:如果满了（10000条），有堵住死锁的风险
							// run := true
							//for run {
							//select {
							//	case s.pushTaskList <- v:
							//		delete(s.waitPushTask, v.PushID)
							//		break
							//	default:
							//		run=fase
							//		break
							//}}
							delete(s.waitPushTask, v.PushID)
						}
					}
				}
				break
			case task := <-s.pushTaskList:
				if s.cfg.Enable {
					notifyDB, err := po.SelectNotifyResult(task.PushID)
					if err != nil {
						log.Errorf("SelectNotifyResult err : %v", err)
						break
					}
					notifyDB.Result = 0
					notifyDB.Num = task.PushNum + 1
					notifyDB.Timestamp = time.Now()
					if s.cfg.Type == "rabbitmq" {
						msg, err := s.mqPush(task)
						notifyDB.Content = msg
						if err == nil {
							notifyDB.Result = 1
						} else {
							log.Info("mqPush err:", err.Error())
						}
					} else {
						msg, err := s.httpPush(task)
						notifyDB.Content = msg
						if err == nil {
							notifyDB.Result = 1
						} else {
							log.Info("httpPush err:", err.Error())
						}
					}
					po.UpdateNotifyResult(notifyDB)
					if notifyDB.Result == 1 {
						log.Infof("push tx success %v", string(task.TxData))
						break
					}
					//重推超过3次的话,不再重推
					if notifyDB.Num > 3 {
						break
					}
					log.Infof(" push task : %v ,err : %v , need repush", task, err)
					// repush, 添加到重试队列
					task.PushNum = notifyDB.Num
					task.PushTime = notifyDB.Timestamp.Unix() + rePushTime[notifyDB.Num-1]
					s.waitPushTask[task.PushID] = task
				}
				break
			}
		}
		log.Infof("push server shutdown ")
		s.done <- struct{}{}
	}()
}

// 添加推送任务
func (s *PushServer) AddPushTask(height int64, txid string, watchlist map[string]bool, pushdata []byte) {
	users := make(map[int64]string)
	for k := range watchlist {
		was, err := s.watch.GetWatchAddress(k)
		if err != nil {
			log.Debugf("don't find user info")
			break
		}
		for _, wui := range was {
			if users[wui.UserID] == "" {
				//写入一个新的通知
				result := &po.NotifyResult{
					Userid: wui.UserID,
					//Height:    height,
					Result: 0,
					Num:    0,
					Txid:   txid,
					//Type:      0,
					Timestamp: time.Now(),
				}
				pushId, err := po.InsertNotifyResult(result)
				if err != nil {
					log.Errorf("%v", err)
					continue
				}
				task := &PushTask{
					PushID:   pushId,
					PushNum:  result.Num,
					TxData:   pushdata,
					PushTime: result.Timestamp.Unix(),
					PushUrl:  wui.NotifyUrl,
					UserID:   wui.UserID,
				}
				s.pushTaskList <- task
				users[wui.UserID] = txid
			}
		}
	}
}

// 添加区块确认数推送任务
func (s *PushServer) AddPushUserTask(height int64, pushdata []byte) {
	userIds, err := po.SelectWatchHeight(height)
	if err != nil {
		log.Warnf("don't have user watch height")
	}
	for userId := range userIds {
		if !s.watch.IsWatchUserExist(userId) {
			continue
		}
		result := &po.NotifyResult{
			Userid: userId,
			//Height:    height,
			Result: 0,
			Num:    0,
			Txid:   "",
			//Type:      1,
			Timestamp: time.Now(),
		}
		pushId, err := po.InsertNotifyResult(result)
		if err != nil {
			log.Warnf("%v", err)
			continue
		}
		url, err := s.watch.GetWatchUserNotifyUrl(userId)
		if err != nil {
			log.Warnf("%v", err)
		}
		task := &PushTask{
			PushID:   pushId,
			PushNum:  result.Num,
			TxData:   pushdata,
			PushTime: result.Timestamp.Unix(),
			PushUrl:  url,
			UserID:   userId,
		}
		s.pushTaskList <- task
	}
}
func (s *PushServer) httpPush(pt *PushTask) (string, error) {
	var (
		req *http.Request
		err error
	)
	if s.cfg.Agent {
		ds, err := json.Marshal(&PushData{
			Uid:  pt.UserID,
			Url:  pt.PushUrl,
			Data: string(pt.TxData),
		})
		if err != nil {
			log.Warnf("err : %v", err)
			return fmt.Sprintf("err : %v", err), err
		}
		req, err = http.NewRequest("POST", s.cfg.Url, bytes.NewBuffer(ds))
		if err != nil {
			log.Warnf("url %s ,err : %v", s.cfg.Url, err)
			return fmt.Sprintf("url %s ,err : %v", s.cfg.Url, err), err
		}
		req.SetBasicAuth(s.cfg.User, s.cfg.Password)
	} else {
		req, err = http.NewRequest("POST", pt.PushUrl, bytes.NewBuffer(pt.TxData))
		if err != nil {
			log.Warnf("url %s ,err : %v", pt.PushUrl, err)
			return fmt.Sprintf("url %s ,err : %v", pt.PushUrl, err), err
		}
	}
	log.Infof("NewRequest %v ", req)
	resp, err := s.client.Do(req)
	if err != nil {
		log.Warnf("post %v ,err : %v", req.URL, err)
		return fmt.Sprintf("post %v ,err : %v", req.URL, err), err
	}
	if resp.StatusCode != 200 {
		return fmt.Sprintf("bad status code %d", resp.StatusCode), fmt.Errorf("bad status code %d", resp.StatusCode)
	}
	log.Infof(" push tx %s ", string(pt.TxData))
	respData, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Warnf("parse body error: %v", err)
		return fmt.Sprintf("parse body error: %v", err), err
	}
	var pushResult bo.PushResult
	err = json.Unmarshal(respData, &pushResult)
	if err != nil {
		return fmt.Sprintf("err : %v", err), err
	}
	if pushResult.Code == 0 {
		return pushResult.Message, nil
	}
	return fmt.Sprintf("code : %d ,message:%s", pushResult.Code, pushResult.Message), fmt.Errorf("code : %d ", pushResult.Code)
}
func (s *PushServer) mqPush(pt *PushTask) (string, error) {
	for _, pb := range s.publishers {
		if err := pb.SendTx(pt.TxData); err != nil {
			log.Errorf("send tx err : %v", err)
			return err.Error(), err
		}
	}
	return "", nil
}

//连接rabbitmq
func (s *PushServer) connectMQ() error {
	rmq, err := rabbitmq.NewServer(s.cfg.MqUrl, s.cfg.Reconns)
	if err != nil {
		return err
	}
	s.rmq = rmq
	//创建签名生产和结果生产
	for _, name := range s.cfg.Publishers {
		if consumer, err := rabbitmq.NewConsumer(s.rmq,
			rabbitmq.ExParams{
				Name:    "data",
				Kind:    "direct",
				Durable: true},
			rabbitmq.QueueParams{
				Name:    name,
				Durable: true,
				Binds:   []rabbitmq.BindParams{{Key: name}},
			},
			rabbitmq.ConsumerParams{
				Tag: "chain-server",
			},
		); err == nil {
			consumer.Close()
		} else {
			return err
		}
		publisher, err := rabbitmq.NewPublisher(s.rmq,
			rabbitmq.ExParams{
				Name:    "data",
				Kind:    "direct",
				Durable: true,
			},
			rabbitmq.PublishParams{
				RoutingKey: name,
			})
		if err != nil {
			return err
		}
		s.publishers[name] = publisher
	}
	return nil
}
func (s *PushServer) disconnectMQ() {
	for _, pb := range s.publishers {
		if err := pb.Close(); err != nil {
			log.Errorf("producer shutdown err : %v", err)
		}
	}
	if err := s.rmq.Shutdown("disconnect with finish"); err != nil {
		log.Errorf("disconnectMQ err : %v", err)
	}
}
