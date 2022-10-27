package fibos

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daofibos"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/common/log"
	"math/rand"
	"strconv"
	"strings"

	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/token"
)

type Task struct {
	WockerID       int
	HeightList     chan int64 // 高度列表
	HeightListExit chan int
}

type HeightResultInfo struct {
	Height int64
	Result int //(1:成功，2：失败)
}

type Quantity struct {
	Quantity eos.Asset       `json:"quantity"`
	Contract eos.AccountName `json:"contract"`
}

// Transfer represents the `transfer` struct on `eosio.token` contract.
type Transfer struct {
	From     eos.AccountName `json:"from"`
	To       eos.AccountName `json:"to"`
	Quantity Quantity        `json:"quantity"`
	Memo     string          `json:"memo"`
}

// 通道列表
var JobsTaskList []*Task
var c = make(chan os.Signal, 10)
var ValuePrecision float64 = 10000.0 // 4位小数
var WockerNum int = 0
var LastBlockNumber int64 = 0
var MaxTaskNum int = 0                                   // 最大允许运行任务数量
var DispossHeightMap map[int64]int = make(map[int64]int) // key:height,value(0:已分配,1:成功，2：失败)
var HeightResults chan *HeightResultInfo = make(chan *HeightResultInfo, 10000)

func GetValue(value float64) float64 {
	_value, _ := strconv.ParseFloat(fmt.Sprintf("%.4f", value/ValuePrecision), 64)
	return _value
}

func GetValueStr(value float64) string {
	return fmt.Sprintf("%.4f", value/ValuePrecision)
}

func AssetString(a *eos.Asset) string {
	amt := a.Amount
	if amt < 0 {
		amt = -amt
	}
	strInt := fmt.Sprintf("%d", amt)
	if len(strInt) < int(a.Symbol.Precision+1) {
		// prepend `0` for the difference:
		strInt = strings.Repeat("0", int(a.Symbol.Precision+uint8(1))-len(strInt)) + strInt
	}

	var result string
	if a.Symbol.Precision == 0 {
		result = strInt
	} else {
		result = strInt[:len(strInt)-int(a.Symbol.Precision)] + "." + strInt[len(strInt)-int(a.Symbol.Precision):]
	}
	if a.Amount < 0 {
		result = "-" + result
	}

	return fmt.Sprintf("%s", result)
}

func worker_task(id int, jobs <-chan int64, exitlist chan<- int, results chan<- *HeightResultInfo) {
	stop := false
	log.Debug("wockerid start", id)
	api := eos.New(beego.AppConfig.String("nodeurl"))

	for !stop {
		select {
		case height := <-jobs:
			if height == -1 {
				stop = true
				break
			}

			// 解析高度
			_res := 1
			if err := parse_data_todb(height, api); err != nil {
				log.Error(err)
				_res = 2
			}

			// 处理结果
			result := &HeightResultInfo{
				Height: height,
				Result: _res,
			}
			results <- result
		}
	}
	log.Debug("wockerid exit", id)
	exitlist <- 1
}

// 初始化通道
func InitSync() {
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGTRAP, syscall.SIGHUP, syscall.SIGQUIT)

	nums := beego.AppConfig.DefaultInt("tasks", 32)
	MaxTaskNum = beego.AppConfig.DefaultInt("maxtasks", 32)
	WockerNum = nums
	log.Debug("init chan num", nums)
	for i := 0; i < nums; i++ {
		JobsTaskList = append(JobsTaskList, new(Task))

		JobsTaskList[i].HeightList = make(chan int64, 100000)
		JobsTaskList[i].HeightListExit = make(chan int)

		go worker_task(i, JobsTaskList[i].HeightList, JobsTaskList[i].HeightListExit, HeightResults)
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
	log.Debug("clearn chan num", WockerNum)
	for i := 0; i < WockerNum; i++ {
		close(JobsTaskList[i].HeightList)
		close(JobsTaskList[i].HeightListExit)
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
		for i := 0; i < WockerNum; i++ {
			JobsTaskList[i].HeightList <- int64(-1)
		}
		// wait task stop
		for i := 0; i < WockerNum; i++ {
			<-JobsTaskList[i].HeightListExit
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

	addtask_func := func(height int64) {
		// 查找无任务的队列
		alloc := false
		for i := 0; i < WockerNum; i++ {
			if len(JobsTaskList[i].HeightList) == 0 {
				JobsTaskList[i].HeightList <- height
				DispossHeightMap[height] = 0
				alloc = true
				break
			}
		}
		if !alloc {
			// 随机分配工作队列
			rand.Seed(time.Now().UnixNano())
			wid := rand.Intn(WockerNum)
			JobsTaskList[wid].HeightList <- height
			DispossHeightMap[height] = 0
		}
	}

	starttime := common.GetMillTime()
	endtime := starttime + beego.AppConfig.DefaultInt64("oncetime", 1*1000)

	// 高度处理结果
	haveresult := true
	for haveresult {
		select {
		case result := <-HeightResults:
			//log.Debug(result.Height, result.Result)
			DispossHeightMap[result.Height] = result.Result
			break
		default:
			haveresult = false
			break
		}
	}

	for true {
		// 删除比db高度小，并且已完成的高度
		for k_height, v := range DispossHeightMap {
			if v == 1 {
				delete(DispossHeightMap, k_height)
			} else if v == 2 {
				// 重新处理失败
				addtask_func(k_height)
			}
		}

		// 获取正在执行任务数量
		total := 0
		for i := 0; i < WockerNum; i++ {
			total += len(JobsTaskList[i].HeightList)
		}
		if total >= MaxTaskNum {
			break
		}

		api := eos.New(beego.AppConfig.String("nodeurl"))
		if api == nil {
			log.Error("node connect error")
			break
		}
		getInfoResult, err := api.GetInfo()
		if err != nil {
			log.Error(err)
			break
		}
		LastBlockNumber = int64(getInfoResult.LastIrreversibleBlockNum)

		// 获取db区块高度
		dbblockcount, err2 := dao.GetMaxBlockIndex()
		if err2 != nil {
			beego.Error(err2)
			break
		}

		log.Debug(LastBlockNumber, dbblockcount)

		// 分配任务
		allownum := MaxTaskNum - total
		for height := (dbblockcount + 1); height < LastBlockNumber; height++ {
			_, ok := DispossHeightMap[height]
			if ok {
				continue
			}

			addtask_func(height)

			allownum--
			if allownum <= 0 {
				break
			}
		}

		break
	}

	currtime := common.GetMillTime()
	if currtime < endtime {
		time.Sleep(time.Millisecond * time.Duration(endtime-currtime))
	}

	return true
}

// 解析指定区块高度到db
func SyncBlockData(tmpval int64) {
	// 解析区块到数据
	log.Debug("start parse block to db index ", tmpval)
	err := parse_data_todb(tmpval, nil)
	log.Debug("end parse block to db index ", tmpval)
	if err != nil {
		beego.Error(err)
	}
}

// 解析指定区块高度到db
func SyncBlockDataHash(blockhash string) {

}

// 解析区块到数据库 result
func parse_data_todb(height int64, eosapi *eos.API) error {
	if eosapi == nil {
		eosapi = eos.New(beego.AppConfig.String("nodeurl"))
	}
	if eosapi == nil {
		return errors.New("connect node error")
	}

	// 区块详情
	result, err := eosapi.GetBlockByNum(uint32(height))
	if err != nil {
		log.Debug(err)
		return err
	}

	blockInfo := Parse_block(result, true, LastBlockNumber)
	if blockInfo == nil {
		log.Debug("block existern !", height)
		return nil
	}

	if blockInfo.Transactions > 0 {
		// 获取交易
		for i := 0; i < len(result.Transactions); i++ {
			txReceipt := &result.Transactions[i]
			txid := txReceipt.Transaction.ID.String()

			log.Debug(height, txid)

			err = Parse_block_tx_todb(txReceipt, blockInfo)
			if err != nil {
				return err
			}
			log.Debug(txid, "finish")
		}
	}

	// 写入区块信息
	num, err := blockInfo.InsertBlockInfo()
	if num <= 0 || err != nil {
		return err
	}

	// 更新区块确认数
	if blockInfo.Confirmations < beego.AppConfig.DefaultInt64("confirmations", 6) {
		go update_confirmations(blockInfo.FrontBlockHash, LastBlockNumber)
	}

	return nil
}

// 解析区块详情到数据库
func Parse_block(result *eos.BlockResp, checkfind bool, cmpheight int64) *dao.BlockInfo {
	if result == nil {
		return nil
	}

	hash := result.ID.String()

	block := dao.NewBlockInfo()
	if checkfind {
		num := block.GetBlockCountByHash(hash)
		if num > 0 {
			return nil
		}
	}

	//log.Debug(result)
	block.Height = int64(result.BlockNum)
	block.Hash = hash
	block.Confirmations = 6
	block.Timestamp = result.Timestamp.Unix()
	block.FrontBlockHash = result.Previous.String()
	block.NextBlockHash = ""
	block.Transactions = len(result.Transactions)

	return block
}

// 解析交易信息到db
func Parse_block_tx_todb(tx *eos.TransactionReceipt, blockInfo *dao.BlockInfo) error {
	if tx == nil || blockInfo == nil {
		return nil
	}

	if tx.TransactionReceiptHeader.Status != eos.TransactionStatusExecuted {
		return nil
	}

	if tx.Transaction.Packed == nil {
		return nil
	}

	var signedTx *eos.SignedTransaction
	var err error
	if tx.Transaction.Packed.Compression == eos.CompressionNone {
		signedTx, err = tx.Transaction.Packed.Unpack()
	} else if tx.Transaction.Packed.Compression == eos.CompressionZlib {
		signedTx, err = tx.Transaction.Packed.UnpackBare()
	}
	if err != nil {
		log.Debug(err)
		return err
	}
	if signedTx == nil {
		return nil
	}
	if len(signedTx.Actions) == 0 {
		return nil
	}

	pushBlockTx := new(models.PushEosBlockInfo)
	pushBlockTx.Type = models.PushTypeEosTX
	pushBlockTx.Height = blockInfo.Height
	pushBlockTx.Hash = blockInfo.Hash
	pushBlockTx.CoinName = beego.AppConfig.String("coin")
	pushBlockTx.Confirmations = blockInfo.Confirmations
	pushBlockTx.Time = blockInfo.Timestamp
	var pushtx models.PushEosTx
	pushtx.Txid = tx.Transaction.ID.String()
	pushtx.Status = "executed"
	pushtx.Fee = 0

	var tmpWatchList map[string]bool = make(map[string]bool)
	for i := 0; i < len(signedTx.Actions); i++ {
		log.Debug(signedTx.Actions[i].Account, signedTx.Actions[i].Name)
		contract_account := string(signedTx.Actions[i].Account)
		if WatchContractList[contract_account] == nil {
			continue
		}
		if signedTx.Actions[i].Name != "transfer" && signedTx.Actions[i].Name != "extransfer" {
			continue
		}

		blocktx := dao.NewBlockTX()
		blocktx.Height = blockInfo.Height
		blocktx.Hash = blockInfo.Hash
		blocktx.Txid = tx.Transaction.ID.String()
		symbol := ""
		if signedTx.Actions[i].Name == "transfer" {
			var res token.Transfer
			err := eos.UnmarshalBinary([]byte(signedTx.Actions[i].HexData), &res)
			if err != nil {
				log.Error(err)
				continue
			}

			if strings.ToLower(res.Quantity.Symbol.Symbol) != "fo" {
				continue
			}

			blocktx.From = string(res.From)
			blocktx.To = string(res.To)
			blocktx.Memo = res.Memo
			blocktx.ContractAddress = contract_account
			blocktx.Amount = res.Quantity.String()
			symbol = res.Quantity.Symbol.Symbol

		} else if signedTx.Actions[i].Name == "extransfer" {
			var res Transfer
			err := eos.UnmarshalBinary([]byte(signedTx.Actions[i].HexData), &res)
			if err != nil {
				log.Error(err)
				continue
			}

			if contract_account == "eosio.token" && strings.ToLower(res.Quantity.Quantity.Symbol.Symbol) == "fo" && res.Quantity.Contract != "eosio" {
				continue
			}

			if strings.ToLower(res.Quantity.Quantity.Symbol.Symbol) != "fo" {
				contract_account = string(res.Quantity.Contract)
			}

			blocktx.From = string(res.From)
			blocktx.To = string(res.To)
			blocktx.Memo = res.Memo
			blocktx.ContractAddress = contract_account
			blocktx.Amount = res.Quantity.Quantity.String()
			symbol = res.Quantity.Quantity.Symbol.Symbol
		}

		if WatchContractList[contract_account] == nil {
			log.Infof("contract_account:%s", contract_account)
			//没有关注的合约，直接忽略
			continue
		}

		// push
		if symbol == WatchContractList[contract_account].Name {
			var action models.PushEosAction
			action.From = blocktx.From
			action.To = blocktx.To
			action.Memo = blocktx.Memo
			action.Contract = blocktx.ContractAddress
			action.Token = symbol
			_tmp := strings.Split(blocktx.Amount, " ")
			action.Amount = _tmp[0]
			pushtx.Actions = append(pushtx.Actions, action)

			if blocktx.From != "" && WatchAddressList[blocktx.From] != nil {
				tmpWatchList[blocktx.From] = true
				log.Debug("watchaddr", blocktx.From)
			}

			if blocktx.To != "" && WatchAddressList[blocktx.To] != nil {
				tmpWatchList[blocktx.To] = true
				log.Debug("watchaddr", blocktx.To)
			}
		}

		num, err := blocktx.Insert()
		if num <= 0 || err != nil {
			beego.Error(err)
		}
	}

	if len(tmpWatchList) > 0 {
		pushBlockTx.Txs = append(pushBlockTx.Txs, pushtx)
		pusdata, err := json.Marshal(&pushBlockTx)
		if err == nil {
			AddPushTask(pushBlockTx.Height, pushBlockTx.Txs[0].Txid, tmpWatchList, pusdata)
		} else {
			log.Debug(err)
		}
	}

	return nil
}

func update_confirmations(frontHash string, cmpheight int64) {

}
