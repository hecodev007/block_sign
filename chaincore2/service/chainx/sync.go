package chainx

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daochainx"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/common/log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/astaxie/beego"
)

type Task struct {
	Txids    chan string
	TxResult chan int

	Vins       []chan interface{}
	VinResult  []chan float64
	Vouts      []chan interface{}
	VoutResult []chan float64
}

type Event struct {
	Index string        `json:"index"`
	Data  []interface{} `json:"data"`
}

type Events struct {
	Method string `json:"method"`
	Ev     Event  `json:"event"`
}

// 通道列表
var JobsTaskList []*Task
var c = make(chan os.Signal, 10)
var ValuePrecision float64 = 100000000.0

func GetValue(value float64) float64 {
	_value, _ := strconv.ParseFloat(fmt.Sprintf("%.8f", value/ValuePrecision), 64)
	return _value
}

func GetValueStr(value float64) string {
	return fmt.Sprintf("%.8f", value/ValuePrecision)
}

// 初始化通道
func InitSync() {
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGTRAP, syscall.SIGHUP, syscall.SIGQUIT)

	nums := runtime.NumCPU()
	log.Debug("init chan num", nums)
	for i := 0; i < nums; i++ {
		JobsTaskList = append(JobsTaskList, new(Task))

		JobsTaskList[i].Txids = make(chan string, 100000)
		JobsTaskList[i].TxResult = make(chan int)

		for j := 0; j < nums; j++ {
			JobsTaskList[i].Vins = append(JobsTaskList[i].Vins, make(chan interface{}, 100000))
			JobsTaskList[i].VinResult = append(JobsTaskList[i].VinResult, make(chan float64))

			JobsTaskList[i].Vouts = append(JobsTaskList[i].Vouts, make(chan interface{}, 100000))
			JobsTaskList[i].VoutResult = append(JobsTaskList[i].VoutResult, make(chan float64))
		}
	}

	// 是否回滚
	if beego.AppConfig.DefaultBool("rollback", false) {
		var rollheight int64 = beego.AppConfig.DefaultInt64("rollheight", 0)
		if rollheight == 0 {
			// 获取db区块高度
			dbblockcount, err2 := dao.GetMaxBlockIndex()
			if dbblockcount == -1 || err2 != nil {
				return
			}
			rollheight = dbblockcount
		}
		Rollback(rollheight)
	}
}

// 清空通道数据
func Clear() {
	nums := runtime.NumCPU()
	log.Debug("clearn chan num", nums)
	for i := 0; i < nums; i++ {
		close(JobsTaskList[i].Txids)
		close(JobsTaskList[i].TxResult)

		for j := 0; j < nums; j++ {
			close(JobsTaskList[i].Vins[j])
			close(JobsTaskList[i].VinResult[j])
			close(JobsTaskList[i].Vouts[j])
			close(JobsTaskList[i].VoutResult[j])
		}
	}
}

func StartSync() {
	go func() {
		run := true
		for true {
			select {
			case s := <-c:
				log.Debug("exit", s)
				run = false
				break
			default:
				break
			}
			if !run {
				break
			}
			result := SyncData()
			if result == false {
				break
			}
		}
		Clear()
		os.Exit(0)
	}()
}

// 同步链数据
func SyncData() bool {
	//定义recover方法，在后面程序出现异常的时候就会捕获
	//defer func() {
	//	//	if r := recover(); r != nil {
	//	//		// 这里可以对异常进行一些处理和捕获
	//	//		log.Debug("Recovered:", r)
	//	//	}
	//	//
	//	//	StartSync()
	//	//}()

	starttime := common.GetMillTime()
	endtime := starttime + 3*1000

	// 获取节点区块高度
	respdata, err := common.Request("chain_getheader", nil)
	if err != nil {
		beego.Error(err)
		time.Sleep(time.Millisecond * 500)
		return true
	} else {
		//log.Debug(string(respdata))
	}
	var datas map[string]interface{}
	err = json.Unmarshal(respdata, &datas)
	if err != nil || datas["error"] != nil {
		log.Debug(err, datas["error"])
		time.Sleep(time.Millisecond * 500)
		return true
	}
	if datas["result"] == nil {
		time.Sleep(time.Millisecond * 500)
		return true
	}
	result := datas["result"].(map[string]interface{})
	blockcount := common.StrBaseToInt64(result["number"].(string), 16)

	// 获取db区块高度
	dbblockcount, err2 := dao.GetMaxBlockIndex()
	if err2 != nil {
		beego.Error(err2)
		time.Sleep(time.Millisecond * 500)
		return true
	}

	if dbblockcount >= blockcount {
		time.Sleep(time.Millisecond * 500)
		return true
	}
	log.Debug(blockcount, dbblockcount)

	tmpcount := dbblockcount
	oncecount, _ := beego.AppConfig.Int("oncecount")
	for i := 0; i < oncecount; i++ {
		// 获取区块数据
		tmpheight := tmpcount + 1

		var getBlockHashResult models.GetBlockHashResult
		err = common.RequestObject("chain_getBlockHash", []interface{}{tmpheight}, &getBlockHashResult)
		if err != nil || getBlockHashResult.Error != "" {
			beego.Error(err, getBlockHashResult.Error)
			return true
		} else {
			//log.Debug(getBlockHashResult.Result)
		}

		hash := getBlockHashResult.Result
		respdata, err := common.Request("chain_getBlock", []interface{}{hash})
		if err != nil {
			beego.Error(err, getBlockHashResult.Error)
			return true
		} else {
			//log.Debug(string(respdata))
		}

		// 解析区块到数据
		log.Debug("start parse block to db index ", tmpheight)
		err = parse_data_todb(respdata, hash, tmpheight)
		log.Debug("end parse block to db index ", tmpheight)
		if err != nil {
			beego.Error(err)
			break
		}

		if tmpheight >= (blockcount - 1) {
			break
		}

		tmpcount++

		currtime := common.GetMillTime()
		if currtime >= endtime {
			break
		}
	}

	currtime := common.GetMillTime()
	if (currtime + 10) < endtime {
		time.Sleep(time.Millisecond * 100)
	}

	return true
}

// 解析区块到数据库 result
func parse_data_todb(blockdata []byte, hash string, height int64) error {
	if blockdata == nil {
		return nil
	}

	var datas map[string]interface{}
	err := json.Unmarshal(blockdata, &datas)
	if err != nil {
		log.Debug(err)
		return err
	}

	if datas["result"] == nil {
		return nil
	}

	// 区块详情
	result := datas["result"].(map[string]interface{})
	blockInfo := parse_block(result, true, hash, height, true)
	if blockInfo == nil {
		log.Debug("block existern !")
		return errors.New("block existern !")
	}

	{
		url := fmt.Sprintf("%s?height=%d", beego.AppConfig.String("parseurl"), height)
		req := httplib.Get(url).SetTimeout(time.Second*1, time.Second*5)
		respdata, err := req.Bytes()
		if respdata == nil || err != nil {
			return err
		}
		//log.Debug(string(respdata))

		var datas map[string]interface{}
		err = json.Unmarshal(respdata, &datas)
		if err != nil {
			log.Debug(err)
			return err
		}
		if datas["result"] == nil {
			return errors.New("get data fail!")
		}

		result := datas["result"].([]interface{})
		blockInfo.Transactions = len(result)

		for i := 0; i < len(result); i++ {
			tx := result[i].(map[string]interface{})
			err = parse_block_tx_todb(0, hash, height, tx, blockInfo)
			if err != nil {
				log.Debug(err)
				return err
			}
		}
	}

	// 写入区块信息
	num, err := blockInfo.InsertBlockInfo()
	if num <= 0 || err != nil {
		return err
	}

	// 更新区块确认数
	if blockInfo.Confirmations < beego.AppConfig.DefaultInt64("confirmations", 6) {
		go update_confirmations(blockInfo.FrontBlockHash, height)
	}

	return nil
}

// 解析区块详情到数据库
func parse_block(result map[string]interface{}, checkfind bool, hash string, cmpheight int64, gettime bool) *dao.BlockInfo {
	if result == nil {
		return nil
	}

	block := dao.NewBlockInfo()
	if checkfind {
		num := block.GetBlockCountByHash(hash)
		if num > 0 {
			return nil
		}
	}

	// 根据hash获取出块时间
	_time := int64(0)
	if gettime {
		url := fmt.Sprintf("%s?hash=%s", beego.AppConfig.String("timestampurl"), hash)
		req := httplib.Get(url).SetTimeout(time.Second*1, time.Second*3)
		data, err := req.Bytes()
		if data != nil && err == nil {
			var timestamp map[string]interface{}
			if err := json.Unmarshal(data, &timestamp); err == nil {

				if timestamp["result"] != nil {
					result := timestamp["result"].(map[string]interface{})
					_time = int64(result["timestamp"].(float64))
				}
			}
		}
	}

	_block := result["block"].(map[string]interface{})
	header := _block["header"].(map[string]interface{})

	//log.Debug(result)
	block.Height = common.StrBaseToInt64(header["number"].(string), 16)
	block.Hash = hash
	block.Confirmations = cmpheight - block.Height + 1
	block.Timestamp = _time
	if header["parentHash"] != nil {
		block.FrontBlockHash = header["parentHash"].(string)
	}
	if header["nextblockhash"] != nil {
		block.NextBlockHash = header["nextblockhash"].(string)
	}
	block.Transactions = 0

	return block
}

// 解析交易信息到db
func parse_block_tx_todb(id int, hash string, height int64, tx map[string]interface{}, blockInfo *dao.BlockInfo) error {
	if tx == nil || tx["tx"] == nil {
		return nil
	}

	//log.Debug(tx)
	if tx["result"].(string) != "ExtrinsicSuccess" {
		return nil
	}

	txid := tx["txHash"].(string)
	_tx := tx["tx"].(map[string]interface{})
	if _tx["method"] == nil || _tx["signature"] == nil {
		return nil
	}

	method := _tx["method"].(map[string]interface{})
	if method["methodName"] != "xAssets::transfer" {
		return nil
	}
	args := method["args"].(map[string]interface{})
	if args["token"] != "PCX" {
		return nil
	}

	signature := _tx["signature"].(map[string]interface{})

	var tmpWatchList map[string]bool = make(map[string]bool)
	blocktx := dao.NewBlockTX()
	blocktx.Height = height
	blocktx.Hash = hash
	blocktx.Txid = txid
	blocktx.Sysfee = 0
	blocktx.From = signature["signer"].(string)
	blocktx.To = args["dest"].(string)
	blocktx.Amount = int64(args["value"].(float64))
	blocktx.Memo = args["memo"].(string)

	// 计算手续费
	var evs []Events
	_events := tx["events"].([]interface{})
	if len(_events) >= 6 {
		tmp, _ := json.Marshal(_events)
		json.Unmarshal(tmp, &evs)
		blocktx.Sysfee = evs[1].Ev.Data[1].(float64) + evs[3].Ev.Data[1].(float64)
	}

	if blocktx.From != "" && WatchAddressList[blocktx.From] != nil {
		tmpWatchList[blocktx.From] = true
		log.Debug("watchaddr", blocktx.From)
	}

	if blocktx.To != "" && WatchAddressList[blocktx.To] != nil {
		tmpWatchList[blocktx.To] = true
		log.Debug("watchaddr", blocktx.To)
	}

	// push
	if len(tmpWatchList) > 0 {
		pushBlockTx := new(models.PushAccountBlockInfo)
		pushBlockTx.Type = models.PushTypeAccountTX
		pushBlockTx.Height = blockInfo.Height
		pushBlockTx.Hash = blockInfo.Hash
		pushBlockTx.CoinName = beego.AppConfig.String("coin")
		pushBlockTx.Confirmations = blockInfo.Confirmations
		pushBlockTx.Time = blockInfo.Timestamp
		pushBlockTx.Txs = append(pushBlockTx.Txs, models.PushAccountTx{
			Txid:   blocktx.Txid,
			Fee:    GetValue(blocktx.Sysfee),
			From:   blocktx.From,
			To:     blocktx.To,
			Amount: GetValueStr(float64(blocktx.Amount)),
			Memo:   blocktx.Memo,
		})

		pusdata, err := json.Marshal(&pushBlockTx)
		if err == nil {
			AddPushTask(height, hash, tmpWatchList, pusdata)
		} else {
			log.Debug(err)
		}
	}

	num, err := blocktx.Insert()
	if num <= 0 || err != nil {
		beego.Error(err)
	}
	return nil
}

func update_confirmations(frontHash string, height int64) {
	// 更新确认数
	confirmations := beego.AppConfig.DefaultInt64("confirmations", 6)
	previousblockhash := frontHash
	for i := int64(0); i < confirmations; i++ {
		respdata, err := common.Request("chain_getBlock", []interface{}{previousblockhash})
		if err != nil {
			beego.Error(err.Error())
			return
		} else {
			//log.Debug(string(respdata))
		}

		var datas map[string]interface{}
		err = json.Unmarshal(respdata, &datas)
		if err != nil {
			log.Debug(err)
			continue
		}

		if datas["result"] == nil {
			continue
		}

		// 区块详情
		result := datas["result"].(map[string]interface{})
		prevBlockInfo := parse_block(result, false, previousblockhash, height, false)
		if prevBlockInfo == nil {
			log.Debug("block existern !")
			continue
		}

		// update db
		//log.Debug(prevBlockInfo.Height, prevBlockInfo.Confirmations, prevBlockInfo.NextBlockHash)
		dao.UpdateConfirmations(prevBlockInfo.Height, prevBlockInfo.Confirmations, prevBlockInfo.NextBlockHash)

		pushBlockTx := new(models.PushAccountBlockInfo)
		pushBlockTx.Type = models.PushTypeAccountConfir
		pushBlockTx.Height = prevBlockInfo.Height
		pushBlockTx.Hash = prevBlockInfo.Hash
		pushBlockTx.CoinName = beego.AppConfig.String("coin")
		pushBlockTx.Confirmations = prevBlockInfo.Confirmations
		pushBlockTx.Time = prevBlockInfo.Timestamp
		pusdata, err := json.Marshal(&pushBlockTx)
		if err == nil {
			AddPushUserTask(prevBlockInfo.Height, pusdata)
		}

		previousblockhash = prevBlockInfo.FrontBlockHash

		if prevBlockInfo.Confirmations >= confirmations || prevBlockInfo.FrontBlockHash == "" {
			break
		}
	}
}
