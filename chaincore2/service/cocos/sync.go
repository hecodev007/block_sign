package cocos

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daococos"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/common/log"
	"github.com/shopspring/decimal"
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
var ValuePrecision float64 = 100000.0
var PrecisionDecimal = decimal.NewFromFloat(ValuePrecision)
var CocosAssetId string = "1.3.0"
var CocosGASAssetId string = "1.3.1"
var MaxHeight int64 = 0 // 链最新高度

func GetValue(value float64) float64 {
	_value, _ := strconv.ParseFloat(fmt.Sprintf("%.5f", value/ValuePrecision), 64)
	return _value
}

func GetValueStr(value float64) string {
	return fmt.Sprintf("%.5f", value/ValuePrecision)
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

	if result, err := IsLocked(); err != nil || result {
		// 解锁钱包
		if !UnlockWallet(beego.AppConfig.String("walletpwd")) {
			time.Sleep(time.Millisecond * 500)
			return true
		}
	}

	// 获取节点区块高度
	data, err := common.Request("get_dynamic_global_properties", nil)
	if err != nil {
		log.Error(err)
		time.Sleep(time.Millisecond * 500)
		return true
	} else {
		//log.Debug(string(data))
	}

	var datas map[string]interface{}
	if err := json.Unmarshal(data, &datas); err != nil {
		log.Error(err)
		time.Sleep(time.Millisecond * 500)
		return true
	}
	if datas["result"] == nil {
		time.Sleep(time.Millisecond * 500)
		return true
	}
	result := datas["result"].(map[string]interface{})
	blockcount := int64(result["last_irreversible_block_num"].(float64))
	//MaxHeight = blockcount

	// 获取db区块高度
	dbblockcount, err2 := dao.GetMaxBlockIndex()
	if err2 != nil {
		log.Error(err2)
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
	for i := 0; i < oncecount; i++ {
		// 获取区块数据
		tmpval := tmpcount + 1
		MaxHeight = tmpval
		respdata, err := common.RequestStr("get_block", []interface{}{tmpval})
		if err != nil {
			log.Error(err)
			return true
		} else {
			//log.Debug(string(respdata))
		}

		// 解析区块到数据
		log.Debug("start parse block to db index ", tmpval)
		err = parse_data_todb(respdata, tmpval)
		log.Debug("end parse block to db index ", tmpval)
		if err != nil {
			log.Error(err)
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
	respdata, err := getblock_data(tmpval)
	if err != nil {
		log.Error(err)
		return
	} else {
		//log.Debug(respdata)
	}

	// 解析区块到数据
	log.Debug("start parse block to db index ", tmpval)
	err = parse_data_todb(respdata, tmpval)
	log.Debug("end parse block to db index ", tmpval)
	if err != nil {
		log.Error(err)
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

			var datas map[string]interface{}

			// 解析原始交易信息
			respdata, err := common.Request("get_transaction_by_id", []interface{}{txid})
			if err != nil {
				log.Error(err)
				continue
			} else {
				//log.Debug(string(respdata))
			}

			err = json.Unmarshal(respdata, &datas)
			if err != nil {
				log.Debug(err)
				continue
			}
			if datas["result"] == nil {
				log.Debug("get_transaction_by_id not found", txid, "reindex = 1 and txindex = 1 ?")
				continue
			}
			tx := datas["result"].(map[string]interface{})
			err = parse_block_tx_todb(id, hash, highindex, txid, tx, blockInfo)
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
func parse_data_todb(blockdata string, height int64) error {
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

	highindex, hash := height, result["block_id"].(string)
	blockInfo := parse_block(result, true, height)
	if blockInfo == nil {
		log.Debug("block existern !")
		return errors.New("block existern !")
	}

	// 区块交易信息
	enablegoroutine := beego.AppConfig.DefaultBool("enablegoroutine", false)
	cpus := runtime.NumCPU()
	txs := result["transactions"].([]interface{})
	for i := 0; i < len(txs); i++ {
		_tx := txs[i].([]interface{})
		txid := _tx[0].(string)

		// 投递到通道
		if enablegoroutine {
			index := i % cpus
			JobsTaskList[index].Txids <- txid
		} else {
			err = parse_block_tx_todb(0, hash, highindex, txid, _tx[1].(map[string]interface{}), blockInfo)
			if err != nil {
				log.Debug(err)
				return err
			}
		}
	}

	if enablegoroutine {
		// 开始执行任务
		for w := 0; w < cpus; w++ {
			go worker_tx(w, JobsTaskList[w].Txids, JobsTaskList[w].TxResult, hash, highindex, blockInfo)
		}

		for a := 0; a < cpus; a++ {
			<-JobsTaskList[a].TxResult
			//log.Debug(a, "finish", result)
		}
	}

	// 写入区块信息
	num, err := blockInfo.InsertBlockInfo()
	if num <= 0 || err != nil {
		return err
	}

	// 更新区块确认数
	//if blockInfo.Confirmations < beego.AppConfig.DefaultInt64("confirmations", 6) {
	//	go update_confirmations(blockInfo.FrontBlockHash, highindex)
	//}

	return nil
}

// 解析区块详情到数据库
func parse_block(result map[string]interface{}, checkfind bool, height int64) *dao.BlockInfo {
	if result == nil {
		return nil
	}

	hash := result["block_id"].(string)

	block := dao.NewBlockInfo()
	if checkfind {
		num := block.GetBlockCountByHash(hash)
		if num > 0 {
			return nil
		}
	}

	//log.Debug(result)
	block.Height = height
	block.Hash = hash
	block.Confirmations = beego.AppConfig.DefaultInt64("confirmations", 6) + 1
	t, _ := time.Parse("2006-01-02T15:04:05", result["timestamp"].(string))
	block.Timestamp = t.Unix()
	if result["previous"] != nil {
		block.FrontBlockHash = result["previous"].(string)
	}
	if result["nextblockhash"] != nil {
		block.NextBlockHash = result["nextblockhash"].(string)
	}
	block.Transactions = len(result["transactions"].([]interface{}))

	return block
}

// 解析交易信息到db
func parse_block_tx_todb(id int, hash string, height int64, txid string, tx map[string]interface{}, blockInfo *dao.BlockInfo) error {
	if tx == nil {
		return nil
	}
	operations := tx["operations"].([]interface{})
	_operations := operations[0].([]interface{})
	if int(_operations[0].(float64)) != 0 {
		return nil
	}

	txinfo := _operations[1].(map[string]interface{})
	from_id := txinfo["from"].(string)
	to_id := txinfo["to"].(string)

	amountobj := txinfo["amount"].(map[string]interface{})
	//if amountobj["asset_id"] != CocosAssetId {
	//	return nil
	//}
	//change by flynn
	// 2020-10-14
	// 添加判断其他代币的功能
	asset_id := amountobj["asset_id"].(string)
	if WatchContractList[asset_id] == nil {
		return nil
	}
	var decimalAmount decimal.Decimal
	switch amountobj["amount"].(type) {
	case string:
		decimalAmount, _ = decimal.NewFromString(amountobj["amount"].(string))
	case float64:
		decimalAmount = decimal.NewFromFloat(amountobj["amount"].(float64))
	}

	// 手续费
	var feeAmount decimal.Decimal
	if tx["operation_results"] != nil {
		operation_results := tx["operation_results"].([]interface{})
		for _, v := range operation_results {
			vv := v.([]interface{})
			if len(vv) >= 2 {
				tmp := vv[1].(map[string]interface{})
				if tmp["fees"] != nil {
					tmp2 := tmp["fees"].([]interface{})
					if len(tmp2) > 0 {
						fees := tmp2[0].(map[string]interface{})
						if fees["asset_id"] == CocosAssetId {
							feeAmount = decimal.NewFromFloat(fees["amount"].(float64))
							feeAmount = feeAmount.Div(PrecisionDecimal)
						}
					}
				}
			}
		}
	}

	//log.Debug(tx)
	var innerAccount bool = false
	var tmpWatchList map[string]bool = make(map[string]bool)

	// 处理精度
	//_amount, _ := decimalAmount.Div(PrecisionDecimal).Float64()
	//write by flynn 2020-10-14
	coin_set := WatchContractList[asset_id]
	_amount, _ := decimalAmount.Shift(-int32(coin_set.Decimal)).Float64()
	//------------------------------------
	blocktx := dao.NewBlockTX()
	blocktx.Height = height
	blocktx.Hash = hash
	blocktx.Txid = txid
	blocktx.Sysfee = 0
	if AccountMap[from_id] != "" {
		blocktx.From = AccountMap[from_id]
	} else {
		blocktx.From = GetAccountById(from_id)
	}
	if AccountMap[to_id] != "" {
		blocktx.To = AccountMap[to_id]
	} else {
		blocktx.To = GetAccountById(to_id)
	}
	if blocktx.From != "" || blocktx.To != "" {
		innerAccount = true
	}
	blocktx.Amount = _amount
	if txinfo["memo"] != nil {
		memoinfo := txinfo["memo"].([]interface{})
		if int(memoinfo[0].(float64)) == 0 {
			blocktx.Memo = memoinfo[1].(string)
		} else if innerAccount {
			blocktx.Memo = GetRawMemo(memoinfo[1].(map[string]interface{}))
		}
	}

	if WatchAddressList[from_id] != nil {
		tmpWatchList[from_id] = true
		log.Debug(from_id, blocktx.From)
	}
	if WatchAddressList[to_id] != nil {
		tmpWatchList[to_id] = true
		log.Debug(to_id, blocktx.To)
	}

	blocktx.Sysfee, _ = feeAmount.Float64()
	num, err := blocktx.Insert()
	if num <= 0 || err != nil {
		log.Error(err)
	} else {
		// push
		if len(tmpWatchList) > 0 {
			pushBlockTx := new(models.PushAccountBlockInfo)
			pushBlockTx.Type = models.PushTypeAccountTX
			pushBlockTx.Height = blockInfo.Height
			pushBlockTx.Hash = blockInfo.Hash
			pushBlockTx.CoinName = beego.AppConfig.String("coin")
			pushBlockTx.Confirmations = blockInfo.Confirmations
			pushBlockTx.Time = blockInfo.Timestamp

			var pushtx models.PushAccountTx
			pushtx.Txid = blocktx.Txid
			pushtx.Fee = blocktx.Sysfee
			pushtx.From = blocktx.From
			pushtx.To = blocktx.To
			pushtx.Amount = fmt.Sprintf("%f", blocktx.Amount)
			pushtx.Memo = blocktx.Memo
			if asset_id != "1.3.0" {
				pushtx.Contract = asset_id //添加asset_id
			}
			pushBlockTx.Txs = append(pushBlockTx.Txs, pushtx)
			pusdata, err := json.Marshal(&pushBlockTx)
			if err == nil {
				AddPushTask(blocktx.Height, blocktx.Txid, tmpWatchList, pusdata)
			}
		}
	}

	return nil
}

func update_confirmations(frontHash string, height int64) {
	// 更新确认数
	confirmations := beego.AppConfig.DefaultInt64("confirmations", 6)
	//previousblockhash := frontHash
	for i := int64(0); i < confirmations; i++ {
		frotheight := height - i - 1
		respdata, err := common.RequestStr("get_block", []interface{}{frotheight})
		if err != nil {
			log.Error(err.Error())
			return
		} else {
			//log.Debug(string(respdata))
		}

		var datas map[string]interface{}
		err = json.Unmarshal([]byte(respdata), &datas)
		if err != nil {
			log.Debug(err)
			continue
		}

		if datas["result"] == nil {
			continue
		}

		// 区块详情
		result := datas["result"].(map[string]interface{})
		prevBlockInfo := parse_block(result, false, frotheight)
		if prevBlockInfo == nil {
			log.Debug("block existern !")
			continue
		}

		// update db
		//log.Debug(prevBlockInfo.Height, prevBlockInfo.Confirmations, prevBlockInfo.NextBlockHash)
		dao.UpdateConfirmations(prevBlockInfo.Height, prevBlockInfo.Confirmations, prevBlockInfo.NextBlockHash)

		pushBlockTx := new(models.PushUtxoBlockInfo)
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

		//previousblockhash = prevBlockInfo.FrontBlockHash

		if prevBlockInfo.Confirmations >= confirmations || prevBlockInfo.FrontBlockHash == "" {
			break
		}
	}
}
