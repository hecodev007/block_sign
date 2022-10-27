package dhx

import (
	"encoding/json"
	_ "encoding/json"
	"errors"
	_ "errors"
	"fmt"
	_ "fmt"
	"reflect"

	//gsrpc "github.com/centrifuge/go-substrate-rpc-client"
	//"github.com/centrifuge/go-substrate-rpc-client/types"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daodhx"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/common/log"
	"github.com/shopspring/decimal"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/astaxie/beego"
)

//replace github.com/centrifuge/go-substrate-rpc-client => ../../../github.com/centrifuge/go-substrate-rpc-client
//replace go.etcd.io/bbolt => go.etcd.io/bbolt@master
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
var ValuePrecision float64 = 1000000000000.0
var NodeOffset int64 = 0
var NodeInfoList []string

func GetValue(value float64) float64 {
	_amount, _ := decimal.NewFromFloat(value).Div(decimal.New(1, int32(12))).Float64()
	return _amount
}

func GetValueStr(value float64) string {
	return decimal.NewFromFloat(value).Div(decimal.New(1, int32(12))).String()
}

func GetValueFromStr(value string) float64 {
	_decimal, _ := decimal.NewFromString(value)
	_amount, _ := _decimal.Div(decimal.New(1, int32(12))).Float64()
	return _amount
}

func GetValueStrFromStr(value string) string {
	_decimal, _ := decimal.NewFromString(value)
	return _decimal.Div(decimal.New(1, int32(12))).String()
}

func GetNode() string {
	if NodeOffset > 99999999999999999 {
		NodeOffset = 0
	}
	index := NodeOffset % int64(len(NodeInfoList))
	NodeOffset++
	//log.Debug(NodeInfoList[index])
	return NodeInfoList[index]
}

// 初始化通道
func InitSync() {
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGTRAP, syscall.SIGHUP, syscall.SIGQUIT)

	// 读取节点列表
	NodeInfoList = strings.Split(beego.AppConfig.String("nodeurl"), ",")

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
		RollbackFromHeight(rollheight)
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
			log.Info("SyncData")
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

	log.Info(11111)
	dhx := NewDhxBlock(beego.AppConfig.String("nodeurl"))
	//starttime := common.GetMillTime()
	//endtime := starttime + 3*1000
	//step1 取节点最新高度
	blockcount, err := dhx.GetblockCount()
	if err != nil {
		time.Sleep(time.Millisecond * 500)
		return true
	}

	//step2 延迟处理
	blockcount = blockcount - beego.AppConfig.DefaultInt64("delayheight", 12)

	log.Info("blockcount:", blockcount)
	//step3获取db高度
	dbblockcount, err2 := dao.GetMaxBlockIndex()
	if err2 != nil {
		log.Error(err2)
		time.Sleep(time.Millisecond * 500)
		return true
	}
	log.Info("dbblockcount:", dbblockcount)
	//step4 对比误差
	if dbblockcount >= blockcount {
		//不需要执行
		time.Sleep(time.Millisecond * 1000)
		return true
	}

	//执行的误差
	num := blockcount - dbblockcount
	log.Info("num:", num)

	for i := int64(0); i < num; i++ {
		blockcount := blockcount + i
		hash, err := dhx.GethashkByHeight(blockcount)
		if err != nil {
			log.Error(err)
			break
		}
		// 解析区块到数据
		log.Debug("start parse block to db index ", blockcount)
		err = parse_data_todb(dhx, blockcount, hash)
		log.Debug("end parse block to db index ", blockcount)
		if err != nil {
			log.Error(err)
			continue
		}
		//currtime := common.GetMillTime()
		//if currtime >= endtime {
		//	break
		//}
	}

	time.Sleep(time.Second * 2)
	//currtime := common.GetMillTime()
	//if (currtime + 10) < endtime {
	//	time.Sleep(time.Millisecond * 1000)
	//}

	return true
}

// 解析指定区块高度到db
func SyncBlockData(tmpval int64) {
	// 解析区块到数据
	log.Debug("start parse block to db index ", tmpval)
	dhx := NewDhxBlock(beego.AppConfig.String("nodeurl"))
	hash, err := dhx.GethashkByHeight(tmpval)
	if err != nil {
		log.Error(err)
		return
	}
	err = parse_data_todb(dhx, tmpval, hash)
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

			// nothing

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

type MyJsonName struct {
	ID      string `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  string `json:"result"`
}

func parse_data_todb(dhx *DhxBlock, cmpheight int64, hash string) error {
	var hasTrans = false

	blockInfo, extrinsics := parse_block(dhx, true, cmpheight, hash)
	if blockInfo == nil {
		//log.Error("block existern !")
		//无交易 无需解析
		return errors.New("block existern")
	}
	blockInfo.Confirmations = 13
	blockInfo.Height = cmpheight

	//第一步，查看是否包含交易信息
	block1, err := GetBlockTransMethodsByHeight(cmpheight)
	if err != nil {
		log.Error(err)
		return err
	}
	blockInfo.Transactions = int64(len(block1.Extrinsics)) //交易确认数
	if len(extrinsics) != len(block1.Extrinsics) {
		//
		log.Errorf("数据对比异常中断：【%s】", blockInfo.Hash)
		return fmt.Errorf("数据对比异常中断：【%s】", blockInfo.Hash)
	}
	for _, v := range block1.Extrinsics { //遍历交易数据，获取时间
		log.Info()
		if !v.Success {
			//失败交易忽略
			log.Infof("txid:%s,事件：%s,失败交易，忽略", v.Hash, v.Method.Pallet+v.Method.Method)
			continue
		}

		if v.Method.Pallet == "timestamp" && v.Method.Method == "set" {
			timeStr := v.Args.Now
			temp := common.StrToInt64(timeStr) / 1000
			blockInfo.Timestamp = common.TimeToStr(temp)

			//args := v.Args.(map[string]interface{})
			//if len(args) > 0 {
			//	timeStr := args["now"].(string)
			//	temp := common.StrToInt64(timeStr) / 1000
			//	blockInfo.Timestamp = common.TimeToStr(temp)
			//}

			//if len(v.Args) > 0 {
			//	timeStr := v.Args[0].(string)
			//	temp := common.StrToInt64(timeStr) / 1000
			//	blockInfo.Timestamp = common.TimeToStr(temp)
			//}
		}
		if v.Method.Pallet == "balances" && v.Method.Method == "transfer" {
			hasTrans = true
		} else if v.Method.Pallet == "balances" && v.Method.Method == "transferKeepAlive" {
			hasTrans = true
		}
	}

	if hasTrans { //当有交易信息时候捕捉交易信息
		block, err := GetBlockTransByHeight(cmpheight)
		if err != nil {
			log.Error("parse_data_todb() return err:", err.Error())
			return err
		}

		for k, v := range block.Extrinsics { //遍历交易数据,获取真正的转账数据
			//successInfo := reflect.TypeOf(v.Success).String() //要为bool型并且为true才解析到数据库
			//if successInfo == "bool" {
			//
			//}
			//if v.Method == "balances.transfer" || v.Method == "balances.transferKeepAlive" { //只检查有效的交易
			if v.Method.Pallet == "balances" && (v.Method.Method == "transfer" || v.Method.Method == "transferKeepAlive") { //只检查有效的交易
				//if v.Method.Pallet == "balances"  && v.Method.Method == "transfer"   { //只检查有效的交易

				txinfo := make(map[string]interface{})

				s := v.Args.Dest
				dd, _ := json.Marshal(s)
				if reflect.TypeOf(s) != nil && reflect.TypeOf(s).String() != "map[string]interface {}" {
					log.Infof("不支持的数据类型 %s，暂时不解析,内容：%s", reflect.TypeOf(s), string(dd))
					continue
				}

				dhxDest := new(Dest)
				err = json.Unmarshal(dd, dhxDest)
				if err != nil || dhxDest.ID == "" {
					log.Infof("错误解析内容:%s", string(dd))
					continue
				}

				if v.Signature.Signer.Id == "" || dhxDest.ID == "" || v.Args.Value == "" {
					log.Debug("交易参数异常")
					continue
				}
				txinfo["from"] = v.Signature.Signer.Id
				txinfo["to"] = dhxDest.ID
				txinfo["amount"], err = decimal.NewFromString(v.Args.Value)
				if err != nil {
					log.Debug(err)
					continue
				}

				//fee, _ := decimal.NewFromString(v.Info.PartialFee)
				fee, _ := dhx.GetPaymentQueryFeeDetails(extrinsics[k], hash)
				if fee.IsZero() {
					log.Infof("HASH[%s],获取是手续费异常，丢弃", hash)
					continue
				}
				txinfo["fee"] = fee //手续费
				txinfo["nonce"] = v.Nonce
				txinfo["signature"] = v.Signature.Signature
				txinfo["txid"] = v.Hash
				txinfo["suc_info"] = v.Success

				parse_block_tx_todb(0, hash, cmpheight, txinfo, blockInfo)
			}
		}
	}

	// 写入区块信息
	num, err := blockInfo.InsertBlockInfo()
	if num <= 0 || err != nil {
		log.Error(err)
		return err
	}

	// 更新区块确认数 bug ，会被误判
	//if blockInfo.Confirmations < beego.AppConfig.DefaultInt64("confirmations", 6) {
	//	go update_confirmations(blockInfo.Hash, blockInfo.Height)
	//}

	return nil
}

func GetAmountAndAdressFromArgs(args []string) (decimal.Decimal, string, error) {
	isNum := common.IsDigit(args[0]) //判断args[0]是否为number
	if isNum {
		d, err := decimal.NewFromString(args[0])
		if err != nil {
			log.Error(err)
			return d, args[1], err
		}
		return d, args[1], nil
	} else {
		d, err := decimal.NewFromString(args[1])
		if err != nil {
			log.Error(err)
			return d, args[0], err
		}
		return d, args[0], nil
	}
}

// 解析区块详情到数据库
func parse_block(dhx *DhxBlock, checkfind bool, cmpheight int64, hash string) (block *dao.BlockInfo, extrinsics []string) {
	ksBlock, err := dhx.GetBlockData(hash)
	if err != nil {
		log.Error(err)
		return nil, nil
	}
	//if len(ksBlock.Result.Block.Extrinsics) > 0 {
	//	log.Errorf("hash[%s]数据长度不足，跳过区块",hash)
	//	return nil,""
	//}
	//metaData = ksBlock.Result.Block.Extrinsics[1]

	block = dao.NewBlockInfo()
	if checkfind {
		num := block.GetBlockCountByHash(hash)
		if num > 0 {
			return nil, nil
		}
	}
	currentheight := common.StrBaseToInt(ksBlock.Result.Block.Header.Number, 16) //通过返回值获取区块高度
	block.Height = int64(currentheight)
	block.Hash = hash
	block.Confirmations = cmpheight - block.Height + 1
	block.Timestamp = time.Now().String()
	block.FrontBlockHash = ksBlock.Result.Block.Header.ParentHash //result.Block.Header.ParentHash.Hex()
	block.NextBlockHash = ""
	block.Transactions = int64(len(ksBlock.Result.Block.Extrinsics))
	return block, ksBlock.Result.Block.Extrinsics
}

// 解析交易信息到db
// hash区块HASH
func parse_block_tx_todb(id int, hash string, index int64, txinfo map[string]interface{}, blockInfo *dao.BlockInfo) error {
	if txinfo == nil {
		return nil
	}

	var tmpWatchList map[string]bool = make(map[string]bool)
	blocktx := dao.NewBlockTX()

	blocktx.Hash = blockInfo.Hash
	blocktx.Txid = txinfo["txid"].(string)
	blocktx.Sysfee = txinfo["fee"].(decimal.Decimal).Shift(-1 * models.DHX_DECIMAL).String()
	blocktx.Height = blockInfo.Height
	blocktx.From = txinfo["from"].(string)
	blocktx.To = txinfo["to"].(string)

	if WatchAddressList[blocktx.From] != nil {
		log.Debug("watchaddr", blocktx.From)
		tmpWatchList[blocktx.From] = true
	}

	if WatchAddressList[blocktx.To] != nil {
		log.Debug("watchaddr", blocktx.To)
		tmpWatchList[blocktx.To] = true
	}

	// push
	if len(tmpWatchList) > 0 {
		success_info := reflect.TypeOf(txinfo["suc_info"]).String() //要为bool型并且为true才解析到数据库
		if success_info == "bool" {
			blocktx.Succuss = reflect.ValueOf(txinfo["suc_info"]).Bool() //判断是否成功的内容
			if blocktx.Succuss {
				blocktx.Amount = txinfo["amount"].(decimal.Decimal).Shift(-1 * models.DHX_DECIMAL).String()
			} else {
				blocktx.Amount = "0"
			}

			pushBlockTx := new(models.PushAccountBlockInfo)
			pushBlockTx.Type = models.PushTypeAccountTX
			pushBlockTx.Height = blockInfo.Height
			pushBlockTx.Hash = blockInfo.Hash
			pushBlockTx.CoinName = beego.AppConfig.String("coin")
			pushBlockTx.Confirmations = blockInfo.Confirmations
			pushBlockTx.Time = common.StrToTime(blockInfo.Timestamp)

			var pushtx models.PushAccountTx
			pushtx.Txid = blocktx.Txid
			pushtx.Fee, _ = txinfo["fee"].(decimal.Decimal).Shift(-1 * models.DHX_DECIMAL).Float64() //blocktx.Sysfee
			pushtx.From = blocktx.From
			pushtx.To = blocktx.To
			pushtx.Amount = blocktx.Amount
			pushtx.Memo = blocktx.Memo
			pushtx.Contract = ""
			pushBlockTx.Txs = append(pushBlockTx.Txs, pushtx)

			pusdata, err := json.Marshal(&pushBlockTx)
			if err == nil {
				AddPushTask(blocktx.Height, blocktx.Txid, tmpWatchList, pusdata)
			} else {
				log.Debug(err)
			}

			num, err := blocktx.Insert()
			if num <= 0 || err != nil {
				log.Debug(err)
			}
		} else {
			blockAbnormal := dao.NewBlockTXAbnormal()
			blockAbnormal.Txid = blocktx.Txid
			blockAbnormal.Hash = blocktx.Hash
			blockAbnormal.Amount = blocktx.Amount
			blockAbnormal.Sysfee = blocktx.Sysfee
			blockAbnormal.Height = blocktx.Height
			blockAbnormal.From = blocktx.From
			blockAbnormal.To = blocktx.To
			blockAbnormal.SucInfo = reflect.ValueOf(txinfo["suc_info"]).String()
			num, err := blockAbnormal.Insert()
			if num <= 0 || err != nil {
				log.Debug(err)
			}
		}
	}

	//if len(tmpWatchList) > 0 {
	//	num, err := blocktx.Insert()
	//	if num <= 0 || err != nil {
	//		log.Debug(err)
	//	}
	//}

	return nil
}

func update_confirmations(hash string, cmpheight int64) {
	// 更新确认数
	confirmations := beego.AppConfig.DefaultInt64("confirmations", 6)
	preNexthash := hash
	for i := int64(0); i < confirmations; i++ {
		frontHeight := cmpheight - i - 1

		// 区块详情
		dhx := NewDhxBlock(beego.AppConfig.String("nodeurl"))
		previousblockhash, err := dhx.GethashkByHeight(frontHeight)
		if previousblockhash == "" || err != nil {
			break
		}
		prevBlockInfo, _ := parse_block(dhx, false, cmpheight, previousblockhash)
		if prevBlockInfo == nil {
			log.Debug("block existern !")
			continue
		}
		//UpdateConfirmations(height int64, confirmations int64, nextblockhash string)
		// update db
		dao.UpdateConfirmations(prevBlockInfo.Height, prevBlockInfo.Confirmations, preNexthash)

		pushBlockTx := new(models.PushUtxoBlockInfo)
		pushBlockTx.Type = models.PushTypeAccountConfir
		pushBlockTx.Height = prevBlockInfo.Height
		pushBlockTx.Hash = prevBlockInfo.Hash
		pushBlockTx.CoinName = beego.AppConfig.String("coin")
		pushBlockTx.Confirmations = prevBlockInfo.Confirmations
		pushBlockTx.Time = time.Now().Unix() //common.StrToTime(prevBlockInfo.Timestamp)
		pusdata, err := json.Marshal(&pushBlockTx)
		if err == nil {
			AddPushUserTask(prevBlockInfo.Height, pusdata)
		}

		previousblockhash = prevBlockInfo.Hash
		preNexthash = prevBlockInfo.Hash

		if prevBlockInfo.Confirmations > confirmations {
			break
		}
	}
}
