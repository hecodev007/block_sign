package rsk

import (
	"encoding/json"
	_ "encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daorsk"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/common/log"
	"github.com/shopspring/decimal"
	"math/big"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/astaxie/beego"
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

//https://github.com/Dev43/arweave-go
// 同步链数据，链上大概半分钟出一个区块
func SyncData() bool {
	starttime := common.GetMillTime()
	endtime := starttime + 6*1000

	//// get fee
	//if time.Now().Unix() > LoadTxFeeTime {
	//	TxFee = GetTxFee()
	//	LoadTxFeeTime += int64(24 * 3600)
	//}


    rskClient := NewRskBlock(beego.AppConfig.String("nodeurl"))
	blockCount,err := rskClient.GetBlockCount() // 获取节点区块高度
	if err != nil {
		log.Error(err.Error())
		time.Sleep(time.Millisecond * 200)
		return true
	}

	// 获取db区块高度
	dbBlockCount, err2 := dao.GetMaxBlockIndex()
	if err2 != nil {
		log.Error(err2.Error())
		time.Sleep(time.Millisecond * 200)
		return true
	}

	if dbBlockCount >= (blockCount - beego.AppConfig.DefaultInt64("delayheight", 12)) {
		time.Sleep(time.Millisecond * 500)
		return true
	}

	tmpcount := dbBlockCount
	oncecount, _ := beego.AppConfig.Int("oncecount")
	for i := 0; i < oncecount; i++ {
		// 获取区块数据
		tmpval := tmpcount + 1

		// 解析区块到数据
		log.Debug("start parse block to db index ", tmpval)
		err = parse_data_todb(rskClient, tmpval)
		log.Debug("end parse block to db index ", tmpval)
		if err != nil {
			log.Error(err.Error())
			break
		}

		if tmpval >= (blockCount - 1 - beego.AppConfig.DefaultInt64("delayheight", 12)) {
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
		//time.Sleep(time.Millisecond * 100)
		time.Sleep(time.Second * 5)
	}

	return true
}

// 解析区块到数据库 result
func parse_data_todb(client *RskClient, height int64) error {
	blockData, err := client.GetBlockDataByNumber(height)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	blockInfo := dao.NewBlockInfo()
	highindex := common.StrBaseToInt64(blockData.Number,16)
	blockInfo.Height = highindex
	blockInfo.Timestamp =  common.TimeToStr(common.StrBaseToInt64(blockData.Timestamp,16))
	blockInfo.FrontBlockHash = blockData.ParentHash
	blockInfo.Hash = blockData.Hash
	blockInfo.Transactions = int64(len(blockData.Transactions))

	// 区块交易信息
	for i := 0; i < len(blockData.Transactions); i++ {

		transaction := blockData.Transactions[i]
		if "0x0000000000000000000000000000000000000000" == transaction.From { //如果输入地址为0，则舍弃掉
			continue
		}


		// 获取原始交易信息
		err :=parse_block_tx_todb(blockData.Hash, highindex, transaction,blockData.GasUsed, blockInfo,blockData.Timestamp)
		if err != nil {
			log.Debug(err)
			return err
		}
	}

	// 写入区块信息
	num, err := blockInfo.InsertBlockInfo()
	if num <= 0 || err != nil {
		return err
	}

	// 更新区块确认数
	if blockInfo.Confirmations < beego.AppConfig.DefaultInt64("confirmations", 6) {
		//go update_confirmations(blockInfo.Hash, blockInfo.Height)
		go update_confirmations(blockInfo.Hash, blockInfo.Height, client)
	}

	return nil
}

// 解析区块详情到数据库
func parse_block(block *BlockDataResult, checkfind bool, cmpheight int64) *dao.BlockInfo {
	if block == nil {
		return nil
	}

	blockInfo := dao.NewBlockInfo()

	//log.Debug(result)
	blockInfo.Height = common.StrBaseToInt64(block.Number,16)
	blockInfo.Hash = block.Hash
	blockInfo.Confirmations = cmpheight - blockInfo.Height + 1
	blockInfo.Timestamp = common.TimeToStr(common.StrBaseToInt64(block.Timestamp,16))
	blockInfo.FrontBlockHash = block.ParentHash
	blockInfo.NextBlockHash = ""
	blockInfo.Transactions = int64(len(block.Transactions))

	return blockInfo
}

// 解析交易信息到block_tx表
//parse_block_tx_todb(blockData.Hash, highindex, transaction)
func parse_block_tx_todb(hash string, highindex int64, transaction Transaction,gasUsed string, blockInfo *dao.BlockInfo,timestamp string) error {
	//WatchContractList
	//var tmpWatchContractList map[string]bool = make(map[string]bool)

	// 获取详细交易数据
	blocktx := dao.NewBlockTX()
	blocktx.FromAddress = transaction.From
	blocktx.BlockHash = hash
	blocktx.BlockHeight = highindex
	blocktx.Txid = transaction.Hash
	blocktx.GasUsed = common.StrBaseToInt64(transaction.Gas,16)
	blocktx.GasPrice = common.StrBaseToInt64(transaction.GasPrice,16)
	blocktx.Nonce = common.StrBaseToInt64(transaction.Nonce,16)
	blocktx.Input = transaction.Input
	blocktx.Timestamp = common.TimeToStr(common.StrBaseToInt64(timestamp,16))

	//Status          int8 //'0代表 失败,1代表成功,2代表上链成功但交易失败',
	//Logs            string
	if highindex == 2376617 {
		fmt.Println("313 From:",transaction.From,",To:",transaction.To)
		highindex = 2376617
	}
	rsk := NewRskBlock(beego.AppConfig.String("nodeurl"))
	has,contactInfo :=HasContact(transaction.To)
	if has { //这种情况表明是关注的代币交易
		toAddress, amount, err := rsk.ParseTransferData(transaction.Input)
		if err != nil {
			log.Error(err)
			return err
		}

		blocktx.ToAddress = toAddress
		blocktx.ContractAddress = transaction.To
		blocktx.CoinName = contactInfo.Name
		blocktx.Decimal = int8(contactInfo.Decimal)

		//以下这段代码是获取logs和status字段的值
		result,err:=rsk.GetTransactionReceipt(transaction.Hash)
		if err != nil {
			log.Error(err)
			return err
		}
		jsonBytes, err := json.Marshal(result.Logs)
		if err != nil {
			fmt.Println(err)
		}
		blocktx.Logs = string(jsonBytes)
		blocktx.Status = int8(common.StrBaseToInt(result.Status,16))

		d := decimal.NewFromBigInt(amount,int32(-1 * contactInfo.Decimal)) //根据从contract_info数据库读取到的精度计算
		fmt.Println("342,d:",d)
		blocktx.Amount = d.String()
		fmt.Println("344,d:",blocktx.Amount)

		//////////////////////
		var tmpWatchList map[string]bool = make(map[string]bool)
		if WatchAddressList[blocktx.FromAddress] != nil {
			log.Debug("watchaddr", blocktx.FromAddress)
			tmpWatchList[blocktx.FromAddress] = true
		}

		if WatchAddressList[blocktx.ToAddress] != nil {
			log.Debug("watchaddr", blocktx.ToAddress)
			tmpWatchList[blocktx.ToAddress] = true
		}

		// push
		if len(tmpWatchList) > 0 {
			pushBlockTx := new(models.PushAccountBlockInfo)
			pushBlockTx.Type = models.PushTypeAccountTX
			pushBlockTx.Height = highindex
			pushBlockTx.Hash = transaction.BlockHash
			pushBlockTx.CoinName = contactInfo.Name
			pushBlockTx.Confirmations = blockInfo.Confirmations
			pushBlockTx.Time = common.StrToTime(blockInfo.Timestamp)

			var pushtx models.PushAccountTx
			pushtx.Txid = blocktx.Txid
			d := decimal.NewFromInt(blocktx.GasPrice).Mul(decimal.NewFromInt(blocktx.GasUsed)).Shift(-18)
			pushtx.Fee,_ = d.Float64()
			pushtx.From = blocktx.FromAddress
			pushtx.To = blocktx.ToAddress
			pushtx.Amount = blocktx.Amount
			//pushtx.Memo = blocktx.Memo
			pushtx.Contract = blocktx.ContractAddress
			pushBlockTx.Txs = append(pushBlockTx.Txs, pushtx)

			pusdata, err := json.Marshal(&pushBlockTx)
			if err == nil {
				AddPushTask(blocktx.BlockHeight, blocktx.Txid, tmpWatchList, pusdata)
			} else {
				log.Debug(err)
			}
		}

		if len(tmpWatchList) > 0 {
			num, err := blocktx.Insert()
			if num <= 0 || err != nil {
				beego.Error(err)
			}
		}

	} else { //如果不是关注的合约地址情况
		isContact,err:=rsk.IsContact(transaction.To)
		if err != nil {
			log.Error(err)
			return err
		} else if !isContact{ //如果不是合约地址，说明是一笔普通交易
			blocktx.ToAddress = transaction.To
			blocktx.CoinName = beego.AppConfig.String("coin") //链上默认的货币，应该是RBTC
			blocktx.Decimal = 18

			result,err:=rsk.GetTransactionReceipt(transaction.Hash)
			if err != nil {
				log.Error(err)
				return err
			}
			jsonBytes, err := json.Marshal(result.Logs)
			if err != nil {
				fmt.Println(err)
			}
			blocktx.Logs = string(jsonBytes)
			blocktx.Status = int8(common.StrBaseToInt(result.Status,16))

			amount,_ := common.StrBaseToBigInt(transaction.Value,16)
			d := decimal.NewFromBigInt(amount,-18) //根据从contract_info数据库读取到的精度计算
			blocktx.Amount = d.String()

			var tmpWatchList map[string]bool = make(map[string]bool)
			if WatchAddressList[blocktx.FromAddress] != nil {
				log.Debug("watchaddr", blocktx.FromAddress)
				tmpWatchList[blocktx.FromAddress] = true
			}

			if WatchAddressList[blocktx.ToAddress] != nil {
				log.Debug("watchaddr", blocktx.ToAddress)
				tmpWatchList[blocktx.ToAddress] = true
			}

			if len(tmpWatchList) < 0 { //如果没有关注的地址直接返回，否则进行交易解析
				return nil
			}

			// push
			if len(tmpWatchList) > 0 {
				pushBlockTx := new(models.PushAccountBlockInfo)
				pushBlockTx.Type = models.PushTypeAccountTX
				pushBlockTx.Height = highindex
				pushBlockTx.Hash = transaction.BlockHash
				pushBlockTx.CoinName = beego.AppConfig.String("coin") ///?
				pushBlockTx.Confirmations = blockInfo.Confirmations
				pushBlockTx.Time = common.StrToTime(blockInfo.Timestamp)

				var pushtx models.PushAccountTx
				pushtx.Txid = blocktx.Txid
				d := decimal.NewFromInt(blocktx.GasPrice).Mul(decimal.NewFromInt(blocktx.GasUsed)).Shift(-18)
				pushtx.Fee,_ = d.Float64()
				pushtx.From = blocktx.FromAddress
				pushtx.To = blocktx.ToAddress
				pushtx.Amount = blocktx.Amount
				//pushtx.Memo = blocktx.Memo
				pushtx.Contract = blocktx.ContractAddress
				pushBlockTx.Txs = append(pushBlockTx.Txs, pushtx)

				pusdata, err := json.Marshal(&pushBlockTx)
				if err == nil {
					AddPushTask(blocktx.BlockHeight, blocktx.Txid, tmpWatchList, pusdata)
				} else {
					log.Debug(err)
				}
				if len(tmpWatchList) > 0 {
					num, err := blocktx.Insert()
					if num <= 0 || err != nil {
						beego.Error(err)
					}
				}
			}
		}
	}

	return nil
}

func update_confirmations(hash string, height int64,client *RskClient) {
	// 更新确认数
	confirmations := beego.AppConfig.DefaultInt64("confirmations", 6)
	previousblockhash := hash
	for i := int64(0); i < confirmations; i++ {
		frontHeight := height - i - 1
		block, err := client.GetBlockDataByNumber(frontHeight)
		if block == nil || err != nil {
			log.Error(err.Error())
			break
		}

		// 区块详情
		prevBlockInfo := parse_block(block, false, height)
		if prevBlockInfo == nil {
			log.Debug("block existern !")
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
		pushBlockTx.Time = common.StrToTime(prevBlockInfo.Timestamp)
		//pusdata, err := json.Marshal(&pushBlockTx)
		//if err == nil {
		//	AddPushUserTask(prevBlockInfo.Height, pusdata)
		//}

		previousblockhash = prevBlockInfo.Hash
		if prevBlockInfo.Confirmations > confirmations {
			break
		}
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

/////////////////////////////////
func StartBlockAmountTimer() {
	/*
	blockID := int64(767713)
	block:= NewRifBlock(beego.AppConfig.String("nodeurl"))
	for{
		fmt.Println("blockID:",blockID)
		result,err:= block.GetBlockDataByNumber(blockID)
		if err != nil {
			continue
		}
		for i:=0;i< len(result.Transactions);i++ {
			if result.Transactions[i].From != "0x0000000000000000000000000000000000000000" &&
				result.Transactions[i].To != "0x0000000000000000000000000000000001000008" &&
				result.Transactions[i].To != "0x0000000000000000000000000000000001000006" &&
				result.Transactions[i].From != "" &&
				result.Transactions[i].To != ""{
				fmt.Println("捕获了一个ID是：",blockID)
				return
			}
		}

		blockID++
	}
	*/

	to,ammount,_:=ParseTransferData("0xa9059cbb00000000000000000000000048bd015a8d33c43de46a6709037171f3e2014c87000000000000000000000000000000000000000000000573b330b335c4140000")
	fmt.Println(to,",",ammount)
}

func ParseTransferData(input string) (to string, amount *big.Int, err error) {

	if strings.Index(input, "0xa9059cbb") != 0 {
		return to, amount, errors.New("input is not transfer data")
	}
	if len(input) < 138 {
		return to, amount, fmt.Errorf("input data isn't 138 , size %d ", 138)
	}
	to = "0x" + input[34:74]
	amount = new(big.Int)
	amount.SetString(input[74:138], 16)
	if amount.Sign() < 0 {
		return to, amount, errors.New("bad amount data")
	}
	return to, amount, nil
}
////////////////////////////////
func Test() {

	//amount,_ := common.StrBaseToBigInt("0x0",16)
	//fmt.Println("562:",amount)
	//d := decimal.NewFromBigInt(amount,18) //根据从contract_info数据库读取到的精度计算
	//fmt.Println("564",d)

	//for i,v := range WatchContractList {
	//	fmt.Println(i,",",v)
	//}

}