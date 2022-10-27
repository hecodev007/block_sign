package ve

import (
	"encoding/json"
	_ "encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/dao_ve"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/common/log"
	"github.com/shopspring/decimal"

	//"github.com/vechain/thor/state"
	"github.com/vechain/thor/thor"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// 链同步服务动作
type BnbSyncAction struct {
}

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
var TxFee float64 = 0.000375
var LoadTxFeeTime int64 = 0
var NodeOffset int64 = 0
var NodeInfoList []string
var InitHeight int64 = 0

func GetValue(value float64) float64 {
	_value, _ := strconv.ParseFloat(fmt.Sprintf("%.6f", value/ValuePrecision), 64)
	return _value
}

func GetValueStr(value float64) string {
	return fmt.Sprintf("%.6f", value/ValuePrecision)
}

func GetStrValueStr(value string) string {
	_tmp, _ := decimal.NewFromString(value)
	return _tmp.Div(decimal.New(1, 6)).String()
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
	//初始化init高度
	InitHeight = beego.AppConfig.DefaultInt64("initheight", 0)
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
	starttime := common.GetMillTime()
	endtime := starttime + 6*1000

	// get fee
	if time.Now().Unix() > LoadTxFeeTime {
		TxFee = GetTxFee()
		LoadTxFeeTime += int64(24 * 3600)
	}

	client := NewRskBlock(beego.AppConfig.String("nodeurl"))
	// 获取节点区块高度
	blockcount, err := client.GetNodeHeight()
	if err != nil {
		log.Error(err)
		time.Sleep(time.Millisecond * 100)
		return true
	}

	// 获取db区块高度
	dbblockcount, err2 := dao.GetMaxBlockIndex()
	if err2 != nil {
		log.Error(err)
		time.Sleep(time.Millisecond * 100)
		return true
	}
	if InitHeight > dbblockcount {
		dbblockcount = InitHeight
	}

	if dbblockcount >= (blockcount - beego.AppConfig.DefaultInt64("delayheight", 12)) {
		time.Sleep(time.Millisecond * 500)
		return true
	}

	tmpcount := dbblockcount
	oncecount, _ := beego.AppConfig.Int("oncecount")
	for i := 0; i < oncecount; i++ {
		// 获取区块数据
		tmpval := tmpcount + 1

		//tmpval = 6136819
		//tmpval = 6127228   //for zenghao test
		block, err := client.GetBlockInfoByHeight(tmpval)
		if err != nil {
			log.Error(err)
			break
		}

		// 解析区块到数据
		log.Debug("start parse block to db index ", tmpval)
		err = parse_data_todb(block, tmpval)
		log.Debug("end parse block to db index ", tmpval)
		if err != nil {
			log.Error(err)
			break
		}

		if tmpval >= (blockcount - 1 - beego.AppConfig.DefaultInt64("delayheight", 12)) {
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
		time.Sleep(time.Second * 5)
	}

	return true
}

func worker_tx(id int, jobs <-chan string, results chan<- int, hash string, height int64, blockInfo *dao.BlockInfo) {
	count := len(jobs)
	offset := 0
	for i := 0; i < count; i++ {
		select {
		case txid := <-jobs:
			offset += 1

			err := parse_block_tx_todb(0, hash, height, txid, blockInfo)
			if err != nil {
				log.Debug(err)
				continue
			}
			//log.Debug(txid, id, "finish")
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

	results <- 1
}

// 解析区块到数据库 result
func parse_data_todb(block *BlockInfo, cmpheight int64) error {

	if block == nil {
		return nil
	}

	blockInfo := dao.NewBlockInfo()
	blockInfo.Height = block.Height
	blockInfo.Timestamp = block.Timestamp
	blockInfo.FrontBlockHash = block.ParentID
	blockInfo.Hash = block.ID
	blockInfo.Transactions = int64(len(block.Transactions))
	blockInfo.Confirmations = 1

	// 区块交易信息
	enablegoroutine := beego.AppConfig.DefaultBool("enablegoroutine", false)
	cpus := runtime.NumCPU()

	for i, v := range block.Transactions {
		txid := fmt.Sprintf("%s", v)
		// 投递到通道
		if enablegoroutine {
			index := i % cpus
			JobsTaskList[index].Txids <- txid
		} else {
			// 获取原始交易信息

			err := parse_block_tx_todb(0, block.ID, block.Height, txid, blockInfo)
			if err != nil {
				log.Debug(err)
				return err
			}
		}
	}

	if enablegoroutine {
		// 开始执行任务
		for w := 0; w < cpus; w++ {
			go worker_tx(w, JobsTaskList[w].Txids, JobsTaskList[w].TxResult, block.ID, block.Height, blockInfo)
		}

		for a := 0; a < cpus; a++ {
			<-JobsTaskList[a].TxResult
		}
	}

	// 写入区块信息
	num, err := blockInfo.InsertBlockInfo()
	if num <= 0 || err != nil {
		return err
	}

	// 更新区块确认数
	if blockInfo.Confirmations < beego.AppConfig.DefaultInt64("confirmations", 12) {
		//go update_confirmations(blockInfo.Hash, blockInfo.Height)
		go update_confirmations(blockInfo.Hash, blockInfo.Height)
	}

	return nil
}

// 解析区块详情到数据库
func parse_block(block *BlockInfo, checkfind bool, cmpheight int64) *dao.BlockInfo {
	if block == nil {
		return nil
	}

	hash := block.ID

	blockInfo := dao.NewBlockInfo()

	blockInfo.Height = block.Height
	blockInfo.Hash = hash
	blockInfo.GetBlockInfoByIndex(block.Height) //从数据库去拿取区块的确认高度
	blockInfo.Confirmations++
	if blockInfo.Confirmations < cmpheight-block.Height+1 {
		blockInfo.Confirmations = cmpheight - block.Height + 1
	}
	blockInfo.Timestamp = block.Timestamp
	blockInfo.FrontBlockHash = block.ParentID
	blockInfo.NextBlockHash = ""
	blockInfo.Transactions = int64(len(block.Transactions))

	return blockInfo
}

// 解析交易信息到block_tx表
func parse_block_tx_todb(id int, hash string, index int64, txId string, blockInfo *dao.BlockInfo) error {

	client := NewRskBlock(beego.AppConfig.String("nodeurl"))
	transaction, err := client.GetTransactionsByID(txId)
	if err != nil {
		log.Debug(err)
		return err
	}
	if transaction.Reverted { //无效交易
		gasFee := transaction.GasPayer
		if WatchAddressList[gasFee] != nil {
			url := fmt.Sprintf("%s/transactions/%s", beego.AppConfig.String("nodeurl"), txId)
			req := httplib.Get(url)
			bytes, err := req.Bytes()
			if err != nil {
				return err
			}
			var resp map[string]interface{}
			json.Unmarshal(bytes, &resp)
			from := resp["origin"].(string)

			paid, _ := common.StrBaseToBigInt(transaction.Paid, 16)
			paid_ := decimal.NewFromBigInt(paid, -int32(18))
			feeAmount := paid_.String() //金额，18位小数

			var tmpWatchList map[string]bool = make(map[string]bool)
			pushBlockTx := new(models.PushAccountBlockInfo)
			pushBlockTx.Type = models.PushTypeAccountTX
			pushBlockTx.Height = blockInfo.Height
			pushBlockTx.Hash = blockInfo.Hash
			pushBlockTx.CoinName = beego.AppConfig.String("coin")
			pushBlockTx.Confirmations = beego.AppConfig.DefaultInt64("confirmations", 12)
			pushBlockTx.Time = blockInfo.Timestamp

			var pushtx models.PushAccountTx
			pushtx.From = from
			pushtx.FeePayer = gasFee
			pushtx.Fee = common.StrToFloat64(feeAmount)
			pushtx.Txid = txId
			pushBlockTx.Txs = []models.PushAccountTx{pushtx}
			pusdata, err := json.Marshal(&pushBlockTx)

			tmpWatchList[from] = true
			if err == nil {
				AddPushTask(blockInfo.Height, txId, tmpWatchList, pusdata)
			} else {
				log.Debug(err)
			}
		}

		return nil
	}

	for i := 0; i < len(transaction.Outputs); i++ {
		for _, event := range transaction.Outputs[i].Events { //这个循环是检测代币的交易信息

			blocktx := dao.NewBlockTX()
			blocktx.Height = blockInfo.Height
			blocktx.Hash = blockInfo.Hash
			blocktx.Txid = txId

			has, contactInfo := HasContact(event.Address)
			if !has || len(event.Topics) <= 0 {
				continue
			}

			paid, _ := common.StrBaseToBigInt(transaction.Paid, 16)
			paid_ := decimal.NewFromBigInt(paid, -int32(contactInfo.Decimal))
			blocktx.Sysfee = paid_.String() //金额，18位小数

			blocktx.FeeName = contactInfo.Name
			blocktx.GasUsed = transaction.GasUsed
			blocktx.ContractAddress = event.Address
			blocktx.GasPayer = transaction.GasPayer

			addrs := make([]string, 0, 3)
			for _, topic := range event.Topics {
				if len(topic) > 0 {
					a, _ := thor.ParseBytes32(topic)
					c := thor.BytesToAddress(a.Bytes())
					addrs = append(addrs, c.String())
				}
			}

			blocktx.From = transaction.Meta.TxOrigin
			//下面这段代码是通过请求另一个获取交易data数据来比较，找出to地址
			toAddr, gas, gasPrice, err := client.GetTransactionsByPending(txId, addrs)

			if err != nil {
				log.Error(err)
				blocktx.To = addrs[len(addrs)-1]
			} else {
				blocktx.To = toAddr
			}
			blocktx.Gas = gas
			blocktx.GasPrice = gasPrice
			if transaction.Meta.TxOrigin != transaction.GasPayer {
				blocktx.GasPayer = transaction.GasPayer
			}

			ammount, _ := common.StrBaseToBigInt(event.Data, 16)
			d := decimal.NewFromBigInt(ammount, -int32(contactInfo.Decimal))
			blocktx.Amount = d.String() //获得金额
			var tmpWatchList map[string]bool = make(map[string]bool)
			if WatchAddressList[blocktx.From] != nil {
				log.Debug("watchaddr", blocktx.From)
				tmpWatchList[blocktx.From] = true
			}
			if WatchAddressList[blocktx.To] != nil {
				log.Debug("watchaddr", blocktx.To)
				tmpWatchList[blocktx.To] = true
			}

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
				pushtx.Fee = common.StrToFloat64(blocktx.Sysfee)
				pushtx.From = blocktx.From
				pushtx.To = blocktx.To
				if transaction.Reverted {
					pushtx.Amount = "0"
				} else {
					pushtx.Amount = blocktx.Amount
				}
				pushtx.Memo = blocktx.Memo
				pushtx.Contract = blocktx.ContractAddress
				pushtx.FeePayer = transaction.GasPayer
				pushBlockTx.Txs = append(pushBlockTx.Txs, pushtx)

				pusdata, err := json.Marshal(&pushBlockTx)
				if err == nil {
					AddPushTask(blocktx.Height, blocktx.Txid, tmpWatchList, pusdata)
				} else {
					log.Debug(err)
				}

				num, err := blocktx.Insert()
				if num <= 0 || err != nil {
					log.Error(err)
				}
			}
		}

		//这是遍历正常转正交易记录
		for _, transfer := range transaction.Outputs[i].Transfers {

			blocktx := dao.NewBlockTX()
			blocktx.Height = blockInfo.Height
			blocktx.Hash = blockInfo.Hash
			blocktx.Txid = txId

			ammount, _ := common.StrBaseToBigInt(transfer.Amount, 16)
			d := decimal.NewFromBigInt(ammount, -18)
			blocktx.Amount = d.String() //金额，18位小数

			paid, _ := common.StrBaseToBigInt(transaction.Paid, 16)
			paid_ := decimal.NewFromBigInt(paid, -18)
			blocktx.Sysfee = paid_.String() //金额，18位小数
			blocktx.FeeName = beego.AppConfig.String("fee_coin")

			blocktx.GasUsed = transaction.GasUsed

			blocktx.From = transfer.Sender  //发送者
			blocktx.To = transfer.Recipient //接收者
			if transfer.Sender != transaction.GasPayer {
				blocktx.GasPayer = transaction.GasPayer
			}
			//beego.Debug(blocktx.From,blocktx.To,blocktx.Height)
			var tmpWatchList map[string]bool = make(map[string]bool)
			if WatchAddressList[blocktx.From] != nil {
				log.Debug("watchaddr", blocktx.From)
				tmpWatchList[blocktx.From] = true
			}
			if WatchAddressList[blocktx.To] != nil {
				log.Debug("watchaddr", blocktx.To)
				tmpWatchList[blocktx.To] = true
			}

			if len(tmpWatchList) > 0 {
				pushBlockTx := new(models.PushAccountBlockInfo)
				pushBlockTx.Type = models.PushTypeAccountTX
				pushBlockTx.Height = blockInfo.Height
				pushBlockTx.Hash = blockInfo.Hash
				pushBlockTx.CoinName = beego.AppConfig.String("coin")
				pushBlockTx.Confirmations = blockInfo.Confirmations
				pushBlockTx.Time = blockInfo.Timestamp

				_, gas, gasPrice, _ := client.GetTransactionsByPending(txId, nil)
				blocktx.Gas = gas
				blocktx.GasPrice = gasPrice

				var pushtx models.PushAccountTx
				pushtx.Txid = blocktx.Txid
				pushtx.Fee = common.StrToFloat64(blocktx.Sysfee)
				pushtx.From = blocktx.From
				pushtx.To = blocktx.To
				if transaction.Reverted {
					pushtx.Amount = "0"
				} else {
					pushtx.Amount = blocktx.Amount
				}
				pushtx.Memo = blocktx.Memo
				pushtx.Contract = blocktx.ContractAddress
				if blocktx.GasPayer != "" {
					pushtx.FeePayer = blocktx.GasPayer
				}
				pushBlockTx.Txs = append(pushBlockTx.Txs, pushtx)

				pusdata, err := json.Marshal(&pushBlockTx)
				//beego.Debug("YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY")
				//beego.Debug(string(pusdata))
				if err == nil {
					AddPushTask(blocktx.Height, blocktx.Txid, tmpWatchList, pusdata)
				} else {
					log.Debug(err)
				}
				num, err := blocktx.Insert()
				if num <= 0 || err != nil {
					log.Error(err)
				}
			}
		}
	}

	return nil
}

func update_confirmations(hash string, height int64) {
	// 更新确认数
	confirmations := beego.AppConfig.DefaultInt64("confirmations", 12)
	previousblockhash := hash
	for i := int64(0); i < confirmations; i++ {
		frontHeight := height - i - 1

		client := NewRskBlock(beego.AppConfig.String("nodeurl"))
		block, err := client.GetBlockInfoByHeight(frontHeight)
		if err != nil {
			log.Error(err)
			continue
		}
		// 区块详情
		prevBlockInfo := parse_block(block, false, height)
		if prevBlockInfo == nil {
			log.Debug("block existern !")
			continue
		}

		if prevBlockInfo.Confirmations > confirmations {
			continue
		}
		// update db
		dao.UpdateConfirmations(prevBlockInfo.Height, prevBlockInfo.Confirmations, previousblockhash)

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

		previousblockhash = prevBlockInfo.Hash
		//if prevBlockInfo.Confirmations > confirmations {
		//	fmt.Println(height,",571:",prevBlockInfo.Confirmations,",",i)
		//	break
		//}
	}

}

func GetMemo(tx string) string {
	url := beego.AppConfig.String("txparseurl")
	req := httplib.Post(url).SetTimeout(time.Second*3, time.Second*10)
	req.JSONBody(map[string]interface{}{
		"tx":     tx,
		"base64": false,
	})
	result, err := req.Bytes()

	if err != nil {
		return ""
	} else {
		resp, _ := req.Response()
		if resp.StatusCode != 200 {
			return ""
		} else {
			var tmp map[string]interface{}
			json.Unmarshal(result, &tmp)
			return tmp["data"].(string)
		}
	}
}

func GetTxFee() float64 {
	var fee float64 = 0.000375
	url := beego.AppConfig.String("https://dex.binance.org/api/v1/fees")
	req := httplib.Get(url).SetTimeout(time.Second*3, time.Second*10)
	result, err := req.Bytes()
	if err != nil {
		return fee
	} else {
		resp, _ := req.Response()
		if resp.StatusCode != 200 {
			return fee
		} else {
			var tmp []interface{}
			json.Unmarshal(result, &tmp)
			for i := 0; i < len(tmp); i++ {
				info := tmp[i].(map[string]interface{})
				if info["fixed_fee_params"] != nil {
					fixed_fee_params := info["fixed_fee_params"].(map[string]interface{})
					return GetValue(fixed_fee_params["fee"].(float64))
				}
			}
		}
	}
	return fee
}
