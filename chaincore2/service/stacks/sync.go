package stacks

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daostacks"
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

// 通道列表
var JobsTaskList []*Task
var c = make(chan os.Signal, 10)
var ValuePrecision float64 = 1000000.0

func GetValue(value float64) float64 {
	_value, _ := strconv.ParseFloat(fmt.Sprintf("%.6f", value/ValuePrecision), 64)
	return _value
}

func GetValueStr(value float64) string {
	return fmt.Sprintf("%.6f", value/ValuePrecision)
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
	respdata, err := common.Request("getblockcount", nil)
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
	blockcount := int64(datas["result"].(float64))

	// 获取db区块高度
	dbblockcount, err2 := dao.GetMaxBlockIndex()
	if err2 != nil {
		beego.Error(err2)
		time.Sleep(time.Millisecond * 500)
		return true
	}

	if dbblockcount >= (blockcount - beego.AppConfig.DefaultInt64("delayheight", 6)) {
		time.Sleep(time.Millisecond * 500)
		return true
	}
	log.Debug(blockcount, dbblockcount)

	tmpcount := dbblockcount
	oncecount, _ := beego.AppConfig.Int("oncecount")

	// 区块交易信息
	enablegoroutine := beego.AppConfig.DefaultBool("enablegoroutine", true)
	for i := 0; i < oncecount; i++ {
		// 获取区块数据
		tmpval := tmpcount + 1

		var getBlockHashResult models.GetBlockHashResult
		err = common.RequestObject("getblockhash", []interface{}{tmpval}, &getBlockHashResult)
		if err != nil || getBlockHashResult.Error != "" {
			beego.Error(err, getBlockHashResult.Error)
			return true
		} else {
			//log.Debug(getBlockHashResult.Result)
		}

		respdata, err := common.RequestStr("getblock", []interface{}{getBlockHashResult.Result, 1})
		if err != nil {
			beego.Error(err, getBlockHashResult.Error)
			return true
		} else {
			//log.Debug(respdata)
		}

		// 解析区块到数据
		log.Debug("start parse block to db index ", tmpval)
		err = parse_data_todb(respdata, enablegoroutine)
		log.Debug("end parse block to db index ", tmpval)
		if err != nil {
			beego.Error(err)
			break
		}

		if tmpval >= (blockcount - 1 - beego.AppConfig.DefaultInt64("delayheight", 6)) {
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

// 解析指定区块高度到db
func SyncBlockData(tmpval int64) {
	var getBlockHashResult models.GetBlockHashResult
	err := common.RequestObject("getblockhash", []interface{}{tmpval}, &getBlockHashResult)
	if err != nil || getBlockHashResult.Error != "" {
		beego.Error(err, getBlockHashResult.Error)
		return
	} else {
		//log.Debug(getBlockHashResult.Result)
	}

	respdata, err := common.RequestStr("getblock", []interface{}{getBlockHashResult.Result, 1})
	if err != nil {
		beego.Error(err, getBlockHashResult.Error)
		return
	} else {
		//log.Debug(respdata)
	}

	// 解析区块到数据
	log.Debug("start parse block to db index ", tmpval)
	err = parse_data_todb(respdata, false)
	log.Debug("end parse block to db index ", tmpval)
	if err != nil {
		beego.Error(err)
	}
}

// 解析指定区块高度到db
func SyncBlockDataHash(blockhash string) {
	respdata, err := common.RequestStr("getblock", []interface{}{blockhash, 1})
	if err != nil {
		beego.Error(err)
		return
	} else {
		//log.Debug(respdata)
	}

	// 解析区块到数据
	log.Debug("start parse block to db index ", blockhash)
	err = parse_data_todb(respdata, false)
	log.Debug("end parse block to db index ", blockhash)
	if err != nil {
		beego.Error(err)
	}
}

func worker_tx(id int, jobs <-chan string, results chan<- int, hash string, highindex int64, blockInfo *dao.BlockInfo) {
	//log.Debug("start", len(jobs))
	count := len(jobs)
	offset := 0
	for i := 0; i < count; i++ {
		select {
		case txid := <-jobs:
			log.Debug(txid, id)
			offset += 1

			// 获取原始交易信息
			log.Debug(txid)
			respdata, err := common.Request("getrawtransaction", []interface{}{txid})
			if err != nil {
				beego.Error(err)
				continue
			} else {
				//log.Debug(string(respdata))
			}

			var datas map[string]interface{}
			err = json.Unmarshal(respdata, &datas)
			if err != nil || datas["error"] != nil {
				log.Debug(err, datas["error"])
				continue
			}

			// 解析原始交易信息
			respdata, err = common.Request("decoderawtransaction", []interface{}{datas["result"].(string)})
			if err != nil {
				beego.Error(err)
				continue
			} else {
				//log.Debug(string(respdata))
			}

			err = json.Unmarshal(respdata, &datas)
			if err != nil {
				log.Debug(err)
				continue
			}

			tx := datas["result"].(map[string]interface{})
			err = Parse_block_tx_todb(id, hash, highindex, tx, blockInfo)
			if err != nil {
				log.Debug(err)
				continue
			}
			log.Debug(txid, id, "finish")
			if offset >= count {
				break
			}

		default:
			offset += 1
			if count == 0 || offset >= count || offset >= 10 {
				break
			}
		}
	}
	//log.Debug("finish 2")
	results <- 1
}

// 解析区块到数据库 result
func parse_data_todb(blockdata string, enablegoroutine bool) error {
	var datas map[string]interface{}
	err := json.Unmarshal([]byte(blockdata), &datas)
	if err != nil {
		log.Debug(err)
		return err
	}

	if datas["result"] == nil {
		return nil
	}

	// 区块详情
	result := datas["result"].(map[string]interface{})

	highindex, hash := int64(result["height"].(float64)), result["hash"].(string)
	blockInfo := Parse_block(result, true)
	if blockInfo == nil {
		log.Debug("block existern !")
		return errors.New("block existern !")
	}

	{
		url := fmt.Sprintf("%s%d", beego.AppConfig.String("stackseurl"), highindex)
		req := httplib.Get(url).SetTimeout(time.Second*3, time.Second*10)
		respdata, err := req.Bytes()
		if respdata == nil || err != nil {
			return err
		}

		var result []interface{}
		err = json.Unmarshal(respdata, &result)
		if err != nil {
			log.Debug(err)
			return err
		}

		for i := 0; i < len(result); i++ {
			tx := result[i].(map[string]interface{})
			Parse_block_tx_todb(0, hash, highindex, tx, blockInfo)
		}
	}

	// 写入区块信息
	num, err := blockInfo.InsertBlockInfo()
	if num <= 0 || err != nil {
		return err
	}

	// 更新区块确认数
	if blockInfo.Confirmations < beego.AppConfig.DefaultInt64("confirmations", 6) {
		go update_confirmations(blockInfo.FrontBlockHash)
	}

	return nil
}

// 解析区块详情到数据库
func Parse_block(result map[string]interface{}, checkfind bool) *dao.BlockInfo {
	if result == nil {
		return nil
	}

	hash := result["hash"].(string)

	block := dao.NewBlockInfo()
	if checkfind {
		num := block.GetBlockCountByHash(hash)
		if num > 0 {
			return nil
		}
	}

	//log.Debug(result)
	block.Height = int64(result["height"].(float64))
	block.Hash = hash
	block.Confirmations = int64(result["confirmations"].(float64))
	block.Timestamp = int64(result["time"].(float64))
	if result["previousblockhash"] != nil {
		block.FrontBlockHash = result["previousblockhash"].(string)
	}
	if result["nextblockhash"] != nil {
		block.NextBlockHash = result["nextblockhash"].(string)
	}
	block.Transactions = len(result["tx"].([]interface{}))

	return block
}

// 解析交易信息到db
func Parse_block_tx_todb(id int, hash string, height int64, tx map[string]interface{}, blockInfo *dao.BlockInfo) error {
	if tx == nil {
		return nil
	}

	if tx["token_units"] == nil || tx["token_units"].(string) != "STACKS" {
		return nil
	}
	if tx["opcode"] == nil || tx["opcode"].(string) != "TOKEN_TRANSFER" {
		return nil
	}

	// 查询交易是否存在
	blocktx := dao.NewBlockTX()
	if blocktx.SelectCount(tx["txid"].(string)) > 0 {
		return nil
	}

	blocktx.Height = height
	blocktx.Hash = hash
	blocktx.Txid = tx["txid"].(string)
	blocktx.Sysfee = 0
	blocktx.From = tx["address"].(string)
	blocktx.To = tx["recipient_address"].(string)
	blocktx.Amount = common.StrToInt64(tx["token_fee"].(string))
	blocktx.Memo = tx["scratch_area"].(string)

	//log.Debug(tx)
	var tmpWatchList map[string]bool = make(map[string]bool)

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
			From:   ConvertAdress("btc", blocktx.From),
			To:     ConvertAdress("btc", blocktx.To),
			Amount: GetValueStr(float64(blocktx.Amount)),
			Memo:   blocktx.Memo,
		})

		pusdata, err := json.Marshal(&pushBlockTx)
		if err == nil {
			AddPushTask(height, blocktx.Txid, tmpWatchList, pusdata)
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

func update_confirmations(frontHash string) {
	// 更新确认数
	confirmations := beego.AppConfig.DefaultInt64("confirmations", 6)
	previousblockhash := frontHash
	for i := int64(0); i < confirmations; i++ {
		respdata, err := common.Request("getblock", []interface{}{previousblockhash})
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
		prevBlockInfo := Parse_block(result, false)
		if prevBlockInfo == nil {
			log.Debug("block existern !")
			continue
		}

		// 重新读取交易
		for true {
			url := fmt.Sprintf("%s%d", beego.AppConfig.String("stackseurl"), prevBlockInfo.Height)
			req := httplib.Get(url).SetTimeout(time.Second*3, time.Second*10)
			respdata, err := req.Bytes()
			if respdata == nil || err != nil {
				break
			}

			var result []interface{}
			err = json.Unmarshal(respdata, &result)
			if err != nil {
				log.Debug(err)
				break
			}

			for i := 0; i < len(result); i++ {
				tx := result[i].(map[string]interface{})
				Parse_block_tx_todb(0, prevBlockInfo.Hash, prevBlockInfo.Height, tx, prevBlockInfo)
			}

			break
		}

		// update db
		//log.Debug(prevBlockInfo.Height, prevBlockInfo.Confirmations, prevBlockInfo.NextBlockHash)
		dao.UpdateConfirmations(prevBlockInfo.Height, prevBlockInfo.Confirmations, prevBlockInfo.NextBlockHash)

		pushBlockTx := new(models.PushUtxoBlockInfo)
		pushBlockTx.Type = models.PushTypeConfir
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

func ConvertAdress(format string, addr string) string {
	url := beego.AppConfig.String("parseurl")
	req := httplib.Post(url).SetTimeout(time.Second*3, time.Second*10)
	req.JSONBody(map[string]interface{}{
		"CoinName": format,
		"Address":  addr,
	})
	result, err := req.Bytes()
	if err != nil {
		log.Debug(err)
		return ""
	} else {
		resp, _ := req.Response()
		if resp.StatusCode != 200 {
			log.Debug(resp.Status)
			return ""
		} else {
			var tmp map[string]interface{}
			json.Unmarshal(result, &tmp)
			if int(tmp["code"].(float64)) == 0 {
				return tmp["data"].(string)
			}
		}
	}
	return ""
}
