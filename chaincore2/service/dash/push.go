package dash

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	dao "github.com/group-coldwallet/chaincore2/dao/daodash"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/common/log"
	"os"
	"os/signal"
	"syscall"
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

// 通道列表
var PushTaskList chan *PushTask = make(chan *PushTask, 10000)
var WaitPushTask map[int64]*PushTask = make(map[int64]*PushTask)

// 初始化推送通道
func InitPush() {

}

// 清空通道数据
func ClearPush() {
	log.Debug("clear chan push")
}

// 添加推送任务
func AddPushTask(height int64, txid string, watchlist map[string]bool, pushdata []byte) {
	log.Debug("AddPushTask ", height, txid)
	var users map[int64]string = make(map[int64]string)
	for k, _ := range watchlist {
		for i := 0; i < len(WatchAddressList[k]); i++ {
			uid := WatchAddressList[k][i].UserID
			url := WatchAddressList[k][i].NotifyUrl
			log.Debug(k, uid, url)
			if users[uid] == "" {
				result := dao.NewNotifyResult()
				result.UserID = uid
				result.Result = 0
				result.Timestamp = time.Now().Unix()
				result.Num = 0
				result.Txid = txid
				pushid, err := result.Insert()
				//log.Debug(pushid, err )
				if err != nil {
					log.Debug(err)
					continue
				}
				RealAddPush(pushid, pushdata, result.Num, result.Timestamp, url, uid)
				users[uid] = txid
				dao.InsertWatchHeight(uid, height)
			}
		}
	}
}

// 添加区块确认数推送任务
func AddPushUserTask(height int64, pushdata []byte) {
	users, _ := dao.SelectWatchHeight(height)
	for i := 0; i < len(users); i++ {
		if UserWatchList[users[i]] == nil {
			continue
		}
		result := dao.NewNotifyResult()
		result.UserID = users[i]
		result.Result = 0
		result.Timestamp = time.Now().Unix()
		result.Num = 0
		result.Txid = ""
		pushid, err := result.Insert()
		if err != nil {
			log.Debug(err)
			continue
		}
		RealAddPush(pushid, pushdata, result.Num, result.Timestamp, UserWatchList[users[i]].NotifyUrl, users[i])
	}
}

func RealAddPush(pushid int64, pushdata []byte, pushnum int, pushtime int64, url string, uid int64) {
	task := new(PushTask)
	task.PushID = pushid
	task.PushNum = pushnum
	task.TxData = pushdata
	task.PushTime = pushtime
	task.PushUrl = url
	task.UserID = uid
	PushTaskList <- task
}

// 运行
func RunPush() {
	log.Debug("run push")
	run := true
	var repushtime = []int64{10, 30, 60, 180, 720, 1800, 7200}
	var c = make(chan os.Signal, 10)
	timer := time.NewTicker(10 * time.Second) // timer
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGTRAP, syscall.SIGHUP, syscall.SIGQUIT)

	go func() {
		for run {
			select {
			case s := <-c:
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
			case txdata := <-PushTaskList:
				{
					//log.Debug(txdata.PushID, txdata.PushNum, txdata.PushTime, string(txdata.TxData), txdata.PushUrl)
					reqbody := map[string]interface{}{
						"uid":  txdata.UserID,
						"url":  txdata.PushUrl,
						"data": string(txdata.TxData),
					}
					agenturl := beego.AppConfig.String("agenturl")
					req := httplib.Post(agenturl).SetTimeout(time.Second*3, time.Second*10)
					req.SetBasicAuth(beego.AppConfig.String("agentuser"), beego.AppConfig.String("agentpass"))
					req.JSONBody(reqbody)
					result, err := req.Bytes()

					notifyDB := dao.NewNotifyResult()
					notifyDB.Id = txdata.PushID
					notifyDB.Result = 0
					notifyDB.Timestamp = time.Now().Unix()
					notifyDB.Content = ""
					notifyDB.Num = txdata.PushNum + 1

					if err != nil {
						notifyDB.Content = err.Error()
						notifyDB.Update()
					} else {
						resp, _ := req.Response()
						if resp.StatusCode != 200 {
							notifyDB.Content = resp.Status
							notifyDB.Update()
						} else {
							var pushResult models.PushResult
							err = json.Unmarshal(result, &pushResult)
							if err != nil {
								notifyDB.Content = err.Error()
								notifyDB.Update()
							}

							if pushResult.Code == 0 {
								notifyDB.Result = 1
							}
							notifyDB.Content = pushResult.Message
							notifyDB.Update()
							if notifyDB.Result == 1 {
								break // success
							}
						}
					}

					if notifyDB.Num > 3 {
						break
					}

					// repush, 添加到重试队列
					log.Debug("repush", txdata.PushID)
					txdata.PushNum = notifyDB.Num
					txdata.PushTime = notifyDB.Timestamp + repushtime[notifyDB.Num-1]
					WaitPushTask[txdata.PushID] = txdata

					break
				}
			default:
				time.Sleep(200 * time.Millisecond)
				break
			}
		}

		if !beego.AppConfig.DefaultBool("enablesync", true) {
			os.Exit(0)
		}
	}()
}
