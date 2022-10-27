package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/common/log"
	"github.com/group-coldwallet/trxsync/models"
	"github.com/group-coldwallet/trxsync/models/po"
	"io/ioutil"
	"net/http"
	"time"
)

type PushTask struct {
	PushID   int64  // 通知结果id
	TxData   []byte // 推送数据
	PushNum  int    // 推送次数
	PushTime int64  // 推送时间
	PushUrl  string // url
	UserID   int64  // 商户id
}

type PushData struct {
	Uid  int64  `json:"uid"`
	Url  string `json:"url"`
	Data string `json:"data"`
}

// 通道列表
var PushTaskList chan *PushTask = make(chan *PushTask, 10000)
var WaitPushTask map[int64]*PushTask = make(map[int64]*PushTask)
var done chan struct{} = make(chan struct{})
var stop chan struct{} = make(chan struct{})

// 添加推送任务
func (bs *BaseService) AddPushTask(height int64, txid string, watchlist map[string]bool, pushdata []byte) {
	log.Infof("收到一笔push交易：txid=[%s]", txid)
	var users map[int64]string = make(map[int64]string)
	for k, _ := range watchlist {
		uaf, err := bs.Watcher.GetWatchAddress(k)
		if err != nil {
			//log.Errorf("get watch address error: %v",err)
			continue
		}
		for i := 0; i < len(uaf); i++ {
			uid := uaf[i].UserID
			url := uaf[i].NotifyUrl
			if users[uid] == "" {
				result := &po.NotifyResult{
					UserID:    uid,
					Height:    height,
					Result:    0,
					Num:       0,
					Txid:      txid,
					Type:      0,
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
					PushUrl:  url,
					UserID:   uid,
				}
				PushTaskList <- task
				users[uid] = txid
			}
		}

	}
}

// 添加区块确认数推送任务
func (bs *BaseService) AddPushUserTask(height int64, pushdata []byte) {
	users, _ := po.SelectWatchHeight(height)
	for userId := range users {
		if bs.Watcher.IsWatchUserExist(userId) {
			result := &po.NotifyResult{
				UserID:    userId,
				Height:    height,
				Result:    0,
				Num:       0,
				Txid:      "",
				Type:      1,
				Timestamp: time.Now(),
			}
			pushId, err := po.InsertNotifyResult(result)
			if err != nil {
				log.Errorf("%v", err)
				continue
			}
			url, err := bs.Watcher.GetWatchUserNotifyUrl(userId)
			if err != nil {
				log.Errorf("%v", err)
			}
			task := &PushTask{
				PushID:   pushId,
				PushNum:  result.Num,
				TxData:   pushdata,
				PushTime: result.Timestamp.Unix(),
				PushUrl:  url,
				UserID:   userId,
			}
			PushTaskList <- task
		}
	}
}

// 运行
func (bs *BaseService) RunPush() {
	log.Info("run push")
	run := true
	var repushtime = []int64{10, 30, 60, 180, 720, 1800, 7200}
	timer := time.NewTicker(10 * time.Second) // timer
	go func() {
		for run {
			select {
			case s := <-stop:
				log.Debug("push exit", s)
				run = false
				break
			case <-timer.C:
				{
					currtime := time.Now().Unix()
					for _, v := range WaitPushTask {
						if currtime >= v.PushTime {
							// 添加到推送队列
							//log.Debug("timeout add PushTaskList", v.PushID)
							PushTaskList <- v
							delete(WaitPushTask, v.PushID)
						}
					}
				}
			case task := <-PushTaskList:
				{
					notifyDB, err := po.SelectNotifyResult(task.PushID)
					if err != nil {
						log.Errorf("SelectNotifyResult err : %v", err)
						break
					}
					notifyDB.Result = 0
					notifyDB.Num = task.PushNum + 1
					notifyDB.Timestamp = time.Now()
					msg, err := bs.httpPush(task)
					log.Infof("推送结果为：%s", msg)
					notifyDB.Content = msg
					if err == nil {
						notifyDB.Result = 1
					} else {
						log.Errorf("push error: %v,msg: %s", err, msg)
					}
					po.UpdateNotifyResult(notifyDB)
					if notifyDB.Result == 1 {
						log.Info("push tx success")
						break
					}
					//重推超过3次的话,不再重推
					if notifyDB.Num > 3 {
						break
					}
					log.Infof(" push task : %v ,err : %v , need repush", string(task.TxData), err)
					// repush, 添加到重试队列
					task.PushNum = notifyDB.Num
					task.PushTime = notifyDB.Timestamp.Unix() + repushtime[notifyDB.Num-1]
					WaitPushTask[task.PushID] = task
				}
			default:
				time.Sleep(200 * time.Millisecond)
				break
			}
		}
		log.Infof("push server shutdown ")
		done <- struct{}{}
	}()
}

func (bs *BaseService) httpPush(pt *PushTask) (string, error) {
	var (
		req *http.Request
		err error
	)
	ds, err := json.Marshal(&PushData{
		Uid:  pt.UserID,
		Url:  pt.PushUrl,
		Data: string(pt.TxData),
	})

	if err != nil {
		log.Errorf("err : %v", err)
		return fmt.Sprintf("err : %v", err), err
	}
	log.Infof("推送数据为： %v", string(pt.TxData))
	req, err = http.NewRequest("POST", bs.Cfg.Push.Url, bytes.NewBuffer(ds))
	if err != nil {
		log.Errorf("url %s ,err : %v", bs.Cfg.Push.Url, err)
		return fmt.Sprintf("url %s ,err : %v", bs.Cfg.Push.Url, err), err
	}
	if bs.Cfg.Push.Agent {
		req.SetBasicAuth(bs.Cfg.Push.User, bs.Cfg.Push.Password)
	}
	log.Infof("NewRequest %v ", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		//log.Info(11111)
		log.Errorf("post %v ,err : %v", req.URL, err)
		return fmt.Sprintf("post %v ,err : %v", req.URL, err), err
	}
	//log.Infof("resp: %v",resp)
	if resp.StatusCode != 200 {
		//log.Info(22222)
		log.Errorf("bad push: %v", resp)
		return fmt.Sprintf("bad status code %d", resp.StatusCode), fmt.Errorf("bad status code %d", resp.StatusCode)
	}
	log.Infof(" push tx %s ", string(pt.TxData))
	respData, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Errorf("parse body error: %v", err)
		return fmt.Sprintf("parse body error: %v", err), err
	}
	var pushResult models.PushResult
	err = json.Unmarshal(respData, &pushResult)
	if err != nil {
		return fmt.Sprintf("err : %v", err), err
	}
	if pushResult.Code == 0 {
		return pushResult.Message, nil
	}
	return fmt.Sprintf("code : %d ,message:%s", pushResult.Code, pushResult.Message), fmt.Errorf("code : %d ", pushResult.Code)

}

func (bs *BaseService) StopPush() {
	close(stop)
	<-done
}
