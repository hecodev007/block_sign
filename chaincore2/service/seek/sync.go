package seek

import (
	"encoding/hex"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daoseek"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/common/log"

	"errors"
	"github.com/shopspring/decimal"
	"math/big"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcomm "github.com/ethereum/go-ethereum/common"
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
var ValuePrecision float64 = 10000000000000000000.0 // 10 ^ 18
var _abi abi.ABI

var LastBlockNumber int64 = 0

// 初始化通道
func InitSync() {
	_abi, _ = abi.JSON(strings.NewReader(TokenABI))
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
	var blockCountResult models.GetBlockHashResult
	resp, err := common.Post("eth_blockNumber", []interface{}{})
	if err != nil {
		beego.Error(err)
		time.Sleep(time.Millisecond * 500)
		return true
	} else {
		//log.Debug(string(resp))
	}

	err = json.Unmarshal(resp, &blockCountResult)
	if err != nil {
		beego.Error(err)
		time.Sleep(time.Millisecond * 500)
		return true
	}

	blockcount := common.StrBaseToInt64(blockCountResult.Result, 16)
	LastBlockNumber = blockcount

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

		height := "0x" + common.Int64ToString(tmpval, 16)
		respdata, err := common.Post("eth_getBlockByNumber", []interface{}{height, false})
		if err != nil {
			beego.Error(err)
			return true
		} else {
			//log.Debug(respdata)
		}

		// 解析区块到数据
		log.Debug("start parse block to db index ", tmpval)
		err = parse_data_todb(respdata, enablegoroutine, tmpval)
		log.Debug("end parse block to db index ", tmpval)
		if err != nil {
			beego.Error(err)
			break
		}

		if tmpval >= (blockcount - 1 - beego.AppConfig.DefaultInt64("delayheight", 2)) {
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
	height := "0x" + common.Int64ToString(tmpval, 16)
	respdata, err := common.Post("eth_getBlockByNumber", []interface{}{height, false})
	if err != nil {
		beego.Error(err)
		return
	} else {
		//log.Debug(respdata)
	}

	// 解析区块到数据
	log.Debug("start parse block to db index ", tmpval)
	err = parse_data_todb(respdata, false, LastBlockNumber)
	log.Debug("end parse block to db index ", tmpval)
	if err != nil {
		beego.Error(err)
	}
}

// 解析指定区块高度到db
func SyncBlockDataHash(blockhash string) {
	respdata, err := common.Post("eth_getBlockByHash", []interface{}{blockhash, false})
	if err != nil {
		beego.Error(err)
		return
	} else {
		//log.Debug(respdata)
	}

	// 解析区块到数据
	log.Debug("start parse block to db index ", blockhash)
	err = parse_data_todb(respdata, false, LastBlockNumber)
	log.Debug("end parse block to db index ", blockhash)
	if err != nil {
		beego.Error(err)
	}
}

// 解析指定txid
func SyncblockTxid(txid string) {
	respdata, err := common.Post("eth_getTransactionByHash", []interface{}{txid})
	if err != nil {
		beego.Error(err)
		return
	} else {
		//log.Debug(respdata)
	}

	var datas map[string]interface{}
	err = json.Unmarshal(respdata, &datas)
	if err != nil || datas["error"] != nil {
		log.Debug(err, datas["error"])
		return
	}

	if datas["result"] == nil {
		return
	}

	tx := datas["result"].(map[string]interface{})
	blockhash := tx["blockHash"].(string)

	// 获取
	{
		blockdata, err := common.Post("eth_getBlockByHash", []interface{}{blockhash, false})
		if err != nil {
			beego.Error(err)
			return
		} else {
			//log.Debug(respdata)
		}

		var datas map[string]interface{}
		err = json.Unmarshal(blockdata, &datas)
		if err != nil {
			log.Debug(err)
			return
		}

		// 区块详情
		result := datas["result"].(map[string]interface{})

		blockInfo := Parse_block(result, false, LastBlockNumber)
		if blockInfo == nil {
			log.Debug("block existern !")
			return
		}
		err = Parse_block_tx_todb(0, blockInfo.Hash, blockInfo.Height, tx, blockInfo)
		if err != nil {
			log.Debug(err)
			return
		}
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

			respdata, err := common.Post("eth_getTransactionByHash", []interface{}{txid})
			if err != nil {
				beego.Error(err)
				continue
			} else {
				//log.Debug(respdata)
			}

			var datas map[string]interface{}
			err = json.Unmarshal(respdata, &datas)
			if err != nil || datas["error"] != nil {
				log.Debug(err, datas["error"])
				continue
			}

			if datas["result"] == nil {
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
func parse_data_todb(blockdata []byte, enablegoroutine bool, cmpheight int64) error {
	if blockdata == nil {
		return nil
	}

	var datas map[string]interface{}
	err := json.Unmarshal(blockdata, &datas)
	if err != nil {
		log.Debug(err)
		return err
	}

	// 区块详情
	result := datas["result"].(map[string]interface{})

	// 高度，块hash
	highindex, hash := common.StrBaseToInt64((result["number"].(string)), 16), result["hash"].(string)

	blockInfo := Parse_block(result, true, cmpheight)
	if blockInfo == nil {
		log.Debug("block existern !")
		return errors.New("block existern !")
	}

	cpus := runtime.NumCPU()
	txs := result["transactions"].([]interface{})
	for i := 0; i < len(txs); i++ {
		txid := txs[i].(interface{}).(string)

		// 投递到通道
		if enablegoroutine {
			index := i % cpus
			JobsTaskList[index].Txids <- txid
		} else {
			respdata, err := common.Post("eth_getTransactionByHash", []interface{}{txid})
			if err != nil {
				beego.Error(err)
				continue
			} else {
				//log.Debug(respdata)
			}

			var datas map[string]interface{}
			err = json.Unmarshal(respdata, &datas)
			if err != nil || datas["error"] != nil {
				log.Debug(err, datas["error"])
				continue
			}

			if datas["result"] == nil {
				continue
			}

			tx := datas["result"].(map[string]interface{})
			err = Parse_block_tx_todb(0, hash, highindex, tx, blockInfo)
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
	if blockInfo.Confirmations < beego.AppConfig.DefaultInt64("confirmations", 6) {
		go update_confirmations(blockInfo.FrontBlockHash, cmpheight)
	}

	return nil
}

// 解析区块详情到数据库
func Parse_block(result map[string]interface{}, checkfind bool, cmpheight int64) *dao.BlockInfo {
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
	block.Height = common.StrBaseToInt64((result["number"].(string)), 16)
	block.Hash = hash
	block.Confirmations = cmpheight - block.Height + 1
	block.Timestamp = common.StrBaseToInt64((result["timestamp"].(string)), 16)
	block.FrontBlockHash = result["parentHash"].(string)
	block.NextBlockHash = ""
	block.Transactions = len(result["transactions"].([]interface{}))

	return block
}

// Bytes2Hex returns the hexadecimal encoding of d.
func Bytes2Hex(d []byte) string {
	return hex.EncodeToString(d)
}

// Hex2Bytes returns the bytes represented by the hexadecimal string str.
func Hex2Bytes(str string) []byte {
	h, _ := hex.DecodeString(str)
	return h
}

// 解析交易信息到db
func Parse_block_tx_todb(id int, hash string, index int64, tx map[string]interface{}, blockInfo *dao.BlockInfo) error {
	if tx == nil {
		return nil
	}

	if tx["from"] == nil || tx["to"] == nil {
		return nil
	}

	watch := false // 该交易是否有关注地址
	txid := tx["hash"].(string)
	from := strings.ToLower(tx["from"].(string))
	to := strings.ToLower(tx["to"].(string))

	// 获取交易凭据
	respdata, err := common.Post("eth_getTransactionReceipt", []interface{}{txid})
	if err != nil {
		beego.Error(err)
		return nil
	} else {
		//log.Debug(string(respdata))
	}

	var datas map[string]interface{}
	err = json.Unmarshal(respdata, &datas)
	if err != nil {
		log.Debug(err)
		return nil
	}

	// 获取交易凭据
	receipt := datas["result"].(map[string]interface{})
	if common.StrBaseToInt((receipt["status"].(string)), 16) != 0x01 {
		log.Debug("status error", txid)
		return nil
	}

	//log.Debug(tx)
	var tmpWatchList map[string]bool = make(map[string]bool)

	blocktx := dao.NewBlockTX()
	blocktx.Height = index
	blocktx.Hash = hash
	blocktx.Txid = txid
	blocktx.GasPrice = common.StrBaseToInt64((tx["gasPrice"].(string)), 16)
	blocktx.GasLimit = common.StrBaseToInt64((tx["gas"].(string)), 16)

	amount, _ := common.StrBaseToBigInt(tx["value"].(string), 16)
	blocktx.Amount = amount.String()
	blocktx.From = from
	blocktx.To = to

	if WatchContractList[to] != nil {
		contractAddress := to
		packed := Hex2Bytes(tx["input"].(string))
		method, err := _abi.MethodById(packed)
		if err != nil || method == nil {
			log.Debug(err)
			return nil
		}
		log.Debug(method.Name, txid)

		values, err := method.Inputs.UnpackValues(packed[4:])
		if err != nil {
			log.Debug(err)
			return nil
		}
		switch method.Name {
		case "transfer":
			{
				if len(values) != 2 {
					break
				}

				switch values[0].(type) {
				case ethcomm.Address:
					address := values[0].(ethcomm.Address)
					_amount := values[1].(*big.Int)

					blocktx.To = strings.ToLower(address.String())
					blocktx.Contract = contractAddress
					blocktx.Amount = _amount.String()

					if WatchAddressList[blocktx.To] != nil {
						watch = true
						tmpWatchList[blocktx.To] = true
						log.Debug("watchaddr", blocktx.To)
					}

					//log.Debug(from, to, _amount.Int64())

				default:
					break
				}

				break
			}
		}
	} else {
		if blocktx.From != "" && WatchAddressList[blocktx.From] != nil {
			watch = true
			tmpWatchList[blocktx.From] = true
			log.Debug("watchaddr", blocktx.From)
		}

		if blocktx.To != "" && WatchAddressList[blocktx.To] != nil {
			watch = true
			tmpWatchList[blocktx.To] = true
			log.Debug("watchaddr", blocktx.To)
		}
	}

	if !watch {
		return nil
	}

	// 手续费
	fee := decimal.NewFromFloat(float64(common.StrBaseToInt64((receipt["gasUsed"].(string)), 16)))
	fee = fee.Mul(decimal.NewFromFloat(float64(blocktx.GasPrice)))
	_fee, _ := fee.Div(decimal.New(1, 18)).Float64()
	blocktx.Fee = _fee

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
		pushtx.Fee = blocktx.Fee
		pushtx.From = blocktx.From
		pushtx.To = blocktx.To
		pushtx.Memo = blocktx.Memo
		pushtx.Contract = blocktx.Contract
		if blocktx.Contract != "" {
			amount, _ := decimal.NewFromString(blocktx.Amount)
			_amount := amount.Div(decimal.New(1, int32(WatchContractList[blocktx.Contract].Decimal))).String()
			pushtx.Amount = _amount
		} else {
			amount, _ := decimal.NewFromString(blocktx.Amount)
			_amount := amount.Div(decimal.New(1, 18)).String()
			pushtx.Amount = _amount
		}
		pushBlockTx.Txs = append(pushBlockTx.Txs, pushtx)

		pusdata, err := json.Marshal(&pushBlockTx)
		if err == nil {
			AddPushTask(blocktx.Height, blocktx.Txid, tmpWatchList, pusdata)
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

func update_confirmations(frontHash string, cmpheight int64) {
	// 更新确认数
	confirmations := beego.AppConfig.DefaultInt64("confirmations", 6)
	previousblockhash := frontHash
	for i := int64(0); i < confirmations; i++ {
		respdata, err := common.Post("eth_getBlockByHash", []interface{}{previousblockhash, false})
		if err != nil {
			beego.Error(err)
			return
		} else {
			//log.Debug(string(respdata))
		}

		var datas map[string]interface{}
		err = json.Unmarshal(respdata, &datas)
		if err != nil {
			log.Debug(err, string(respdata))
			continue
		}

		if datas["result"] == nil {
			return
		}

		// 区块详情
		result := datas["result"].(map[string]interface{})

		// 区块详情
		prevBlockInfo := Parse_block(result, false, cmpheight)
		if prevBlockInfo == nil {
			log.Debug("block existern !")
			continue
		}

		// update db
		//log.Debug(prevBlockInfo.Height, prevBlockInfo.Confirmations, prevBlockInfo.NextBlockHash)
		dao.UpdateConfirmations(prevBlockInfo.Height, prevBlockInfo.Confirmations, previousblockhash)

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

func FromHexToAddress(hexaddr string) string {
	if strings.HasPrefix(hexaddr, "0x") {
		hexaddr = strings.TrimPrefix(hexaddr, "0x")
	}

	// 转换地址
	respdata, err := common.Post("fromhexaddress", []interface{}{hexaddr})
	if err != nil {
		beego.Error(err)
		return ""
	}

	var datas map[string]interface{}
	err = json.Unmarshal(respdata, &datas)
	if err != nil || datas["error"] != nil {
		log.Debug(err, datas["error"])
		return ""
	}

	return datas["result"].(string)
}
