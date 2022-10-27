package ckb

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/btcsuite/btcutil/bech32"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daockb"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/common/log"
	"strconv"

	"errors"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
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
var ValuePrecision float64 = 100000000.0 // 8位小数

var LastBlockNumber int64 = 0

func GetValue(value float64) float64 {
	_value, _ := strconv.ParseFloat(fmt.Sprintf("%.8f", value/ValuePrecision), 64)
	return _value
}

func GetValueStr(value float64) string {
	return fmt.Sprintf("%.8f", value/ValuePrecision)
}

const (
	PREFIX_MAINNET          = "ckb"
	PREFIX_TESTNET          = "ckt"
	SECP_BLAKE160_CODE_HASH = "9bd7e06f3ecf4be0f2fcd2188b23f1b9fcc88e5d4b65a8637b17723bbda3cce8" //如果版本变更注意切换
	MULTISIG_CODE_HASH      = "5c5069eb0857efc65e1bca0c07df34c31663b3622fd3876c876320fc9634e2a8" //如果版本变更注意切换
	TYPE_SHORT              = "01"                                                               //short version for locks with popular code_hash
	TYPE_FULL_DATA          = "02"                                                               //full version with hash_type = "Data"
	TYPE_FULL_TYPE          = "04"                                                               //full version with hash_type = "Type"
	CODE_HASH_IDX_BLAKE160  = "00"
	CODE_HASH_IDX_MULTISIG  = "01"
)

type HashType string

const (
	DATA HashType = "data" //byte 00
	TYPE HashType = "type" // byte 01
)

type CkbScript struct {
	CodeHash string
	Args     string
	HashType HashType
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
	var blockCountResult models.GetBlockHashResult
	err := common.RequestObject("get_tip_block_number", nil, &blockCountResult)
	if err != nil {
		beego.Error(err)
		time.Sleep(time.Millisecond * 500)
		return true
	} else {
		//log.Debug(blockCountResult)
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
		respdata, err := common.Request("get_block_by_number", []interface{}{height})
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
	respdata, err := common.Request("get_block_by_number", []interface{}{height})
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
	respdata, err := common.Request("get_block", []interface{}{blockhash})
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
			respdata, err := common.Request("get_transaction", []interface{}{txid})
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

			_tx := datas["result"].(map[string]interface{})
			transaction := _tx["transaction"].(map[string]interface{})
			err = Parse_block_tx_todb(id, hash, highindex, transaction, blockInfo)
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

	if datas["result"] == nil {
		return nil
	}

	// 区块详情
	result := datas["result"].(map[string]interface{})

	// 高度，块hash
	header := result["header"].(map[string]interface{})
	highindex, hash := common.StrBaseToInt64((header["number"].(string)), 16), header["hash"].(string)

	blockInfo := Parse_block(result, true, cmpheight)
	if blockInfo == nil {
		log.Debug("block existern !")
		return errors.New("block existern !")
	}

	if blockInfo.Transactions > 0 {
		// 获取交易
		cpus := runtime.NumCPU()
		txs := result["transactions"].([]interface{})
		for i := 0; i < len(txs); i++ {
			tx := txs[i].(map[string]interface{})
			txid := tx["hash"].(string)

			// 投递到通道
			if enablegoroutine {
				index := i % cpus
				JobsTaskList[index].Txids <- txid
			} else {
				log.Debug(txid)
				err = Parse_block_tx_todb(0, hash, highindex, tx, blockInfo)
				if err != nil {
					log.Debug(err)
					return err
				}
				log.Debug(txid, "finish")
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

	header := result["header"].(map[string]interface{})
	hash := header["hash"].(string)

	block := dao.NewBlockInfo()
	if checkfind {
		num := block.GetBlockCountByHash(hash)
		if num > 0 {
			return nil
		}
	}

	//log.Debug(result)
	block.Height = common.StrBaseToInt64((header["number"].(string)), 16)
	block.Hash = hash
	block.Confirmations = cmpheight - block.Height + 1
	block.Timestamp = common.StrBaseToInt64((header["timestamp"].(string)), 16) / 1000
	block.FrontBlockHash = header["parent_hash"].(string)
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

// 处理vout
func disposs_vout(hash string, highindex int64, txid string, vout map[string]interface{}, n int) (float64, string, *dao.BlockTXVout, bool) {
	var vout_amount float64 = 0
	var watchaddr string = ""
	if vout == nil {
		return vout_amount, watchaddr, nil, false
	}

	lock := vout["lock"].(map[string]interface{})
	amount, err := common.StrBaseToBigInt(vout["capacity"].(string), 16)
	if !err {
		return vout_amount, watchaddr, nil, false
	}
	sc := &CkbScript{
		CodeHash: lock["code_hash"].(string),
		Args:     lock["args"].(string),
		HashType: TYPE,
	}

	txvout := dao.NewBlockTXVout()
	txvout.Height = highindex
	txvout.Hash = hash
	txvout.Txid = txid
	txvout.Voutn = n
	txvout.Voutvalue = amount.Int64()
	vout_amount += float64(txvout.Voutvalue)
	if beego.AppConfig.DefaultString("network", "mainnet") == "mainnet" {
		txvout.Voutaddress = sc.GetCkbAddr(PREFIX_MAINNET)
	} else {
		txvout.Voutaddress = sc.GetCkbAddr(PREFIX_TESTNET)
	}
	txvout.CodeHash = sc.CodeHash
	txvout.Args = sc.Args

	txvout.Invaild = 0
	txvout.Status = common.Confirmed
	num, err2 := txvout.Insert()
	if num <= 0 || err2 != nil {
		log.Debug(err)
		return vout_amount, watchaddr, nil, false
	}

	// 添加或设置用户资产
	if txvout.Voutaddress != "" && txvout.Voutvalue > 0 {
		dao.FindSetBlockAmount(txvout.Voutaddress, txvout.Voutvalue)
	}
	watchaddr = txvout.Voutaddress
	return vout_amount, watchaddr, txvout, false
}

func disposs_vin(hash string, highindex int64, txid string, vin map[string]interface{}) (float64, string, *dao.BlockTXVin) {
	var vin_amount float64 = 0
	var watchaddr string = ""
	if vin == nil {
		return vin_amount, watchaddr, nil
	}

	txvin := dao.NewBlockTXVin()
	txvin.Height = highindex
	txvin.Hash = hash
	txvin.Txid = txid

	if vin["previous_output"] == nil {
		return vin_amount, watchaddr, nil
	}

	// prevhash previndex
	previous_output := vin["previous_output"].(map[string]interface{})
	if previous_output["tx_hash"] != nil {
		txvin.Vintxid = previous_output["tx_hash"].(string)
	}
	if previous_output["index"] != nil {
		txvin.VinVoutindex = common.StrBaseToInt64((previous_output["index"].(string)), 16)
	}

	num, err := txvin.Insert()
	if num <= 0 || err != nil {
		log.Debug(err)
		return vin_amount, watchaddr, nil
	}

	if previous_output["tx_hash"].(string) != "0x0000000000000000000000000000000000000000000000000000000000000000" && previous_output["index"].(string) != "0xffffffff" {
		//获取上一个输出交易
		txvout := dao.NewBlockTXVout()
		result, err := txvout.Select(txvin.Vintxid, txvin.VinVoutindex)
		if result && err == nil {
			vin_amount += float64(txvout.Voutvalue)
			txvout.UpdateStatus(common.Spent, txvin.Vintxid, txvin.VinVoutindex)

			// 扣除资产
			dao.UpdateBlockAmount(txvout.Voutaddress, -txvout.Voutvalue)

			watchaddr = txvout.Voutaddress
			txvin.Address = txvout.Voutaddress
			txvin.Amount = txvout.Voutvalue
		}
	}

	return vin_amount, watchaddr, txvin
}

// 解析交易信息到db
func Parse_block_tx_todb(id int, hash string, index int64, tx map[string]interface{}, blockInfo *dao.BlockInfo) error {
	if tx == nil {
		return nil
	}

	//log.Debug(tx)
	var tmpWatchList map[string]bool = make(map[string]bool)
	var blockvout_list []*dao.BlockTXVout
	var blockvin_list []*dao.BlockTXVin

	blocktx := dao.NewBlockTX()
	blocktx.Height = index
	blocktx.Hash = hash
	blocktx.Txid = tx["hash"].(string)
	blocktx.Sysfee = 0
	blocktx.Vincount = len(tx["inputs"].([]interface{}))
	blocktx.Voutcount = len(tx["outputs"].([]interface{}))
	blocktx.Coinbase = 0

	var vout_amount float64 = 0.0
	var coinbasetx bool = false
	if tx["outputs"] != nil {
		//log.Debug(tx["outputs"])
		vouts := tx["outputs"].([]interface{})
		for j := 0; j < len(vouts); j++ {
			vout := vouts[j].(map[string]interface{})
			tmpamount, watchaddr, blockvout, _ := disposs_vout(hash, index, blocktx.Txid, vout, j)
			vout_amount += tmpamount
			if blockvout != nil {
				blockvout_list = append(blockvout_list, blockvout)
			}

			// 关注列表
			if watchaddr != "" && WatchAddressList[watchaddr] != nil {
				log.Debug("watchaddr", watchaddr)
				tmpWatchList[watchaddr] = true
			}
		}
	}

	var vin_amount float64 = 0.0
	if tx["inputs"] != nil {
		//log.Debug(tx["inputs"])
		vins := tx["inputs"].([]interface{})
		for j := 0; j < len(vins); j++ {
			vin := vins[j].(map[string]interface{})
			tmpamount, watchaddr, blockvin := disposs_vin(hash, index, blocktx.Txid, vin)
			vin_amount += tmpamount
			if blockvin != nil {
				blockvin_list = append(blockvin_list, blockvin)
			}

			// 关注列表
			if watchaddr != "" && WatchAddressList[watchaddr] != nil {
				log.Debug("watchaddr", watchaddr)
				tmpWatchList[watchaddr] = true
			}
		}
	}

	// 手续费
	if tx["fee"] == nil {
		blocktx.Sysfee = (vin_amount - vout_amount) / ValuePrecision
		if blocktx.Sysfee < 0 {
			blocktx.Sysfee = 0
		}
	} else {
		blocktx.Sysfee = tx["fee"].(float64)
	}

	// push
	if len(tmpWatchList) > 0 {
		pushBlockTx := new(models.PushUtxoBlockInfo)
		pushBlockTx.Type = models.PushTypeTX
		pushBlockTx.Height = blockInfo.Height
		pushBlockTx.Hash = blockInfo.Hash
		pushBlockTx.CoinName = beego.AppConfig.String("coin")
		pushBlockTx.Confirmations = blockInfo.Confirmations
		pushBlockTx.Time = blockInfo.Timestamp
		var pushtx models.PushUtxoTx
		pushtx.Txid = blocktx.Txid
		pushtx.Fee = blocktx.Sysfee
		pushtx.Coinbase = coinbasetx
		for i := 0; i < len(blockvin_list); i++ {
			// checkout address
			if blockvin_list[i].Address == "" && blockvin_list[i].Amount == 0 {
				for true {
					// 获取原始交易信息
					respdata, err := common.Request("get_transaction", []interface{}{blockvin_list[i].Vintxid})
					if err != nil {
						beego.Error(err)
						break
					} else {
						//log.Debug(string(respdata))
					}

					var datas map[string]interface{}
					err = json.Unmarshal(respdata, &datas)
					if err != nil || datas["error"] != nil {
						log.Debug(err, datas["error"])
						break
					}

					if datas["result"] == nil {
						log.Debug("get_transaction not found", blockvin_list[i].Vintxid, "reindex = 1 and txindex = 1 ?")
						break
					}
					tx := datas["result"].(map[string]interface{})
					transaction := tx["transaction"].(map[string]interface{})
					tmpvouts := transaction["outputs"].([]interface{})
					if tmpvouts != nil && tmpvouts[blockvin_list[i].VinVoutindex] != nil {
						vout := tmpvouts[blockvin_list[i].VinVoutindex].(map[string]interface{})
						blockvin_list[i].Amount = common.StrBaseToInt64((vout["capacity"].(string)), 16)
						lock := vout["lock"].(map[string]interface{})
						sc := &CkbScript{
							CodeHash: lock["code_hash"].(string),
							Args:     lock["args"].(string),
							HashType: TYPE,
						}
						if beego.AppConfig.DefaultString("network", "mainnet") == "mainnet" {
							blockvin_list[i].Address = sc.GetCkbAddr(PREFIX_MAINNET)
						} else {
							blockvin_list[i].Address = sc.GetCkbAddr(PREFIX_TESTNET)
						}
						if blockvin_list[i].Address != "" {
							tmpWatchList[blockvin_list[i].Address] = true
							vin_amount += float64(blockvin_list[i].Amount)
						}
					}
					break
				}
			}
			value := GetValueStr(float64(blockvin_list[i].Amount))
			pushtx.Vin = append(pushtx.Vin, models.PushTxInput{Txid: blockvin_list[i].Vintxid, Vout: int(blockvin_list[i].VinVoutindex), Addresse: blockvin_list[i].Address, Value: value})
		}
		for i := 0; i < len(blockvout_list); i++ {
			value := GetValueStr(float64(blockvout_list[i].Voutvalue))
			pushtx.Vout = append(pushtx.Vout, models.PushTxOutput{Addresse: blockvout_list[i].Voutaddress, Value: value, N: blockvout_list[i].Voutn, CodeHash: &blockvout_list[i].CodeHash})
		}
		// 重新计算手续费
		blocktx.Sysfee = GetValue(float64(vin_amount - vout_amount))
		if blocktx.Sysfee < 0 {
			blocktx.Sysfee = 0
		}
		pushtx.Fee = blocktx.Sysfee
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
		respdata, err := common.Request("get_block", []interface{}{previousblockhash})
		if err != nil {
			beego.Error(err)
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

		// 区块详情
		prevBlockInfo := Parse_block(result, false, cmpheight)
		if prevBlockInfo == nil {
			log.Debug("block existern !")
			continue
		}

		// update db
		//log.Debug(prevBlockInfo.Height, prevBlockInfo.Confirmations, prevBlockInfo.NextBlockHash)
		dao.UpdateConfirmationsHash(prevBlockInfo.Hash, prevBlockInfo.Confirmations, previousblockhash)

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

func FromHexToAddress(prefix string, hexaddr string) string {
	if strings.HasPrefix(hexaddr, "0x") {
		hexaddr = strings.TrimPrefix(hexaddr, "0x")
	}

	blake160Addr, _ := hex.DecodeString(hexaddr)
	prefixFlag, _ := hex.DecodeString("0100")
	payload := append(prefixFlag, blake160Addr...)

	converted, err := bech32.ConvertBits(payload, 8, 5, true)
	if err != nil {
		log.Debug(err)
		return ""
	}
	addr, err := bech32.Encode(prefix, converted)
	if err != nil {
		log.Debug(err)
		return ""
	}

	return addr
}

func (sc *CkbScript) GetCkbAddr(prefix string) string {
	var payload []byte
	if strings.HasPrefix(sc.Args, "0x") {
		sc.Args = sc.Args[2:len(sc.Args)]
	}
	if strings.HasPrefix(sc.CodeHash, "0x") {
		sc.CodeHash = sc.CodeHash[2:len(sc.CodeHash)]
	}
	if sc.HashType == TYPE && len(sc.Args) == 40 {
		if sc.CodeHash == SECP_BLAKE160_CODE_HASH {
			payload, _ = hex.DecodeString(TYPE_SHORT + CODE_HASH_IDX_BLAKE160 + sc.Args)
		} else if sc.CodeHash == MULTISIG_CODE_HASH {
			payload, _ = hex.DecodeString(TYPE_SHORT + CODE_HASH_IDX_MULTISIG + sc.Args)
		} else {
			payload, _ = sc.generateFullAddress()
		}
	} else {
		payload, _ = sc.generateFullAddress()
	}
	converted, err := bech32.ConvertBits(payload, 8, 5, true)
	if err != nil {
		panic(err)
	}
	addr, err := bech32.Encode(prefix, converted)
	if err != nil {
		panic(err)
	}
	return addr
}

//全地址格式
func (sc *CkbScript) generateFullAddress() ([]byte, error) {
	var hashType string
	if sc.HashType == TYPE {
		hashType = TYPE_FULL_TYPE
	} else {
		hashType = TYPE_FULL_DATA
	}
	return hex.DecodeString(hashType + sc.CodeHash + sc.Args)
}
