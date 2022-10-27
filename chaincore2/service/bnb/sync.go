package bnb

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	_ "encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/binance-chain/go-sdk/types/tx"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daobnb"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/common/log"
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

// Coin def
type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// Transaction Input
type Input struct {
	Address string `json:"address"`
	Coins   []Coin `json:"coins"`
}

// Transaction Output
type Output struct {
	Address string `json:"address"`
	Coins   []Coin `json:"coins"`
}

// SendMsg - high level transaction of the coin module
type SendMsg struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

type SendMsgTx struct {
	Type string  `json:"type"`
	Msg  SendMsg `json:"value"`
}

// 通道列表
var JobsTaskList []*Task
var c = make(chan os.Signal, 10)
var ValuePrecision float64 = 100000000.0
var TxFee float64 = 0.000375
var LoadTxFeeTime int64 = 0
var NodeOffset int64 = 0
var NodeInfoList []string

func GetValue(value float64) float64 {
	_value, _ := strconv.ParseFloat(fmt.Sprintf("%.8f", value/ValuePrecision), 64)
	return _value
}

func GetValueStr(value float64) string {
	return fmt.Sprintf("%.8f", value/ValuePrecision)
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

// 同步链数据
func SyncData() bool {
	starttime := common.GetMillTime()
	endtime := starttime + 3*1000

	// get fee
	if time.Now().Unix() > LoadTxFeeTime {
		TxFee = GetTxFee()
		LoadTxFeeTime += int64(24 * 3600)
	}

	req := httplib.Get(fmt.Sprintf("http://%s/abci_info", GetNode()))
	respbyte, err := req.Bytes()
	if err != nil {
		beego.Error(err)
		time.Sleep(time.Millisecond * 500)
		return true
	}
	var respdata map[string]interface{}
	json.Unmarshal(respbyte, &respdata)

	if respdata == nil {
		return true
	}

	result := respdata["result"].(map[string]interface{})
	if result == nil {
		return true
	}
	response := result["response"].(map[string]interface{})
	if response == nil {
		return true
	}

	// 获取节点区块高度
	blockcount := common.StrToInt64(response["last_block_height"].(string))

	// 获取db区块高度
	dbblockcount, err2 := dao.GetMaxBlockIndex()
	if err2 != nil {
		beego.Error(err2)
		time.Sleep(time.Millisecond * 500)
		return true
	}

	if dbblockcount >= (blockcount - beego.AppConfig.DefaultInt64("delayheight", 12)) {
		time.Sleep(time.Millisecond * 500)
		return true
	}
	log.Debug(blockcount, dbblockcount)

	tmpcount := dbblockcount
	oncecount, _ := beego.AppConfig.Int("oncecount")
	for i := 0; i < oncecount; i++ {
		// 获取区块数据
		tmpval := tmpcount + 1
		//fmt.Printf("处理最新高度： %d\n",tmpval)
		req := httplib.Get(fmt.Sprintf("http://%s/block?height=%d", GetNode(), tmpval))
		blockbyte, err := req.Bytes()
		if err != nil {
			beego.Error(err)
			break
		}

		var respblock map[string]interface{}
		err = json.Unmarshal(blockbyte, &respblock)
		if err != nil {
			beego.Error(err)
			break
		}

		result := respblock["result"].(map[string]interface{})

		// 解析区块到数据
		log.Debug("start parse block to db index ", tmpval)
		err = parse_data_todb(result, tmpval)
		log.Debug("end parse block to db index ", tmpval)
		if err != nil {
			beego.Error(err)
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
		time.Sleep(time.Millisecond * 100)
	}

	return true
}

// 解析指定区块高度到db
func SyncBlockData(tmpval int64) {
	req := httplib.Get(fmt.Sprintf("http://%s/block?height=%d", GetNode(), tmpval))
	blockbyte, err := req.Bytes()
	if err != nil {
		beego.Error(err)
		return
	}

	var respblock map[string]interface{}
	err = json.Unmarshal(blockbyte, &respblock)
	if err != nil {
		beego.Error(err)
		return
	}

	req = httplib.Get(fmt.Sprintf("http://%s/abci_info", GetNode()))
	respbyte, err := req.Bytes()
	if err != nil {
		beego.Error(err)
		time.Sleep(time.Millisecond * 500)
		return
	}
	var respdata map[string]interface{}
	json.Unmarshal(respbyte, &respdata)
	result := respdata["result"].(map[string]interface{})
	response := result["response"].(map[string]interface{})

	// 获取节点区块高度
	lastBlockHeight := common.StrToInt64(response["last_block_height"].(string))

	// 解析区块到数据
	log.Debug("start parse block to db index ", tmpval)
	result = respblock["result"].(map[string]interface{})
	err = parse_data_todb(result, lastBlockHeight)
	log.Debug("end parse block to db index ", tmpval)
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

// 解析区块到数据库 result
func parse_data_todb(data map[string]interface{}, cmpheight int64) error {
	if data == nil {
		return nil
	}
	blockMeta := data["block_meta"].(map[string]interface{})
	block_id := blockMeta["block_id"].(map[string]interface{})
	header := blockMeta["header"].(map[string]interface{})
	highindex, hash := common.StrToInt64(header["height"].(string)), block_id["hash"].(string)
	blockInfo := parse_block(data, true, cmpheight)
	if blockInfo == nil {
		log.Debug("block existern !")
		return errors.New("block existern !")
	}

	// 区块交易信息
	enablegoroutine := beego.AppConfig.DefaultBool("enablegoroutine", false)
	cpus := runtime.NumCPU()

	if blockInfo.Transactions > 0 {
		//q, err := tmquery.New(fmt.Sprintf("tx.height=%d", 30664539))
		//tx, err := client.TxSearch(q, false, 1, 100)
		url := fmt.Sprintf("http://%s/tx_search?query=\"tx.height=%d\"&per_page=100", GetNode(), highindex)
		req := httplib.Get(url)
		respdata, err := req.Bytes()
		if err != nil {
			log.Debug(err)
			return nil
		}

		var result map[string]interface{}
		err = json.Unmarshal(respdata, &result)
		if err != nil {
			beego.Error(err)
			return nil
		}

		_result := result["result"].(map[string]interface{})
		txs := _result["txs"].([]interface{})
		for i := 0; i < len(txs); i++ {
			tx := txs[i].(map[string]interface{})
			txid := tx["hash"].(string)

			// 投递到通道
			if enablegoroutine {
				index := i % cpus
				JobsTaskList[index].Txids <- txid
			} else {
				// 获取原始交易信息
				log.Debug(txid)

				err = parse_block_tx_todb(0, hash, highindex, tx, blockInfo)
				if err != nil {
					log.Debug(err)
					return err
				}
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

	//更新区块确认数
	if blockInfo.Confirmations < beego.AppConfig.DefaultInt64("confirmations", 6) {
		//go update_confirmations(blockInfo.Hash, blockInfo.Height)
		update_confirmations(blockInfo.Hash, blockInfo.Height)
	}

	return nil
}

// 解析区块详情到数据库
func parse_block(data map[string]interface{}, checkfind bool, cmpheight int64) *dao.BlockInfo {
	if data == nil {
		return nil
	}

	blockMeta := data["block_meta"].(map[string]interface{})
	block_id := blockMeta["block_id"].(map[string]interface{})
	hash := block_id["hash"].(string)

	block := dao.NewBlockInfo()
	if checkfind {
		num := block.GetBlockCountByHash(hash)
		if num > 0 {
			return nil
		}
	}

	//log.Debug(result)
	header := blockMeta["header"].(map[string]interface{})
	next_block_id := header["last_block_id"].(map[string]interface{})
	block.Height = common.StrToInt64(header["height"].(string))
	block.Hash = hash
	//block.Confirmations = cmpheight - block.Height + 1
	block.Confirmations = beego.AppConfig.DefaultInt64("confirmations", 12) + 1
	_time, _ := time.Parse(time.RFC3339Nano, header["time"].(string))
	block.Timestamp = _time.Unix()
	block.FrontBlockHash = next_block_id["hash"].(string)
	block.NextBlockHash = ""
	block.Transactions = common.StrToInt64(header["num_txs"].(string))

	return block
}

// 解析交易信息到db
func parse_block_tx_todb(id int, hash string, index int64, txinfo map[string]interface{}, blockInfo *dao.BlockInfo) error {
	if txinfo == nil {
		return nil
	}

	txid := txinfo["hash"].(string)
	_tx := txinfo["tx"].(string)
	b, _ := base64.StdEncoding.DecodeString(_tx)
	//log.Debug(hex.EncodeToString(b))

	var stdtx tx.StdTx
	err := tx.Cdc.UnmarshalBinaryLengthPrefixed(b, &stdtx)
	//if err != nil {
	//	beego.Error(err)
	//	return nil
	//}
	if len(stdtx.Msgs) == 0 {
		return nil
	}
	if stdtx.Msgs[0].Type() != "send" {
		return nil
	}

	d1, err := tx.Cdc.MarshalJSON(stdtx.Msgs[0])
	if err != nil {
		log.Debug(err)
		return nil
	}
	//log.Debug(string(d1))

	var sendmsg SendMsgTx
	err = json.Unmarshal(d1, &sendmsg)
	if err != nil {
		log.Debug(err)
		return nil
	}

	//log.Debug(stdtx, stdtx.Msgs[0].Type(), stdtx.Msgs[0].Route())
	//log.Debug(sendmsg)

	// check coin
	denom := sendmsg.Msg.Inputs[0].Coins[0].Denom
	if WatchContractList[denom] == nil {
		return nil
	}

	//log.Debug(tx)
	var tmpWatchList map[string]bool = make(map[string]bool)

	blocktx := dao.NewBlockTX()
	blocktx.Height = index
	blocktx.Hash = hash
	blocktx.Txid = txid
	blocktx.Sysfee = TxFee
	blocktx.From = sendmsg.Msg.Inputs[0].Address
	blocktx.To = sendmsg.Msg.Outputs[0].Address
	blocktx.Memo = GetMemo(hex.EncodeToString(b))
	blocktx.Amount = common.StrToFloat64(sendmsg.Msg.Inputs[0].Coins[0].Amount)
	blocktx.ContractAddress = denom

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
		pushtx.Amount = GetValueStr(blocktx.Amount)
		pushtx.Memo = blocktx.Memo
		pushtx.Contract = denom
		pushBlockTx.Txs = append(pushBlockTx.Txs, pushtx)

		pusdata, err := json.Marshal(&pushBlockTx)
		if err == nil {
			log.Debugf("pusdata:%s", pusdata)
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

func update_confirmations(hash string, height int64) {
	// 更新确认数
	confirmations := beego.AppConfig.DefaultInt64("confirmations", 6)
	previousblockhash := hash
	for i := int64(0); i < confirmations; i++ {
		frotheight := height - i - 1

		req := httplib.Get(fmt.Sprintf("http://%s/block?height=%d", GetNode(), frotheight))
		blockbyte, err := req.Bytes()
		if err != nil {
			beego.Error(err)
			break
		}

		var respblock map[string]interface{}
		err = json.Unmarshal(blockbyte, &respblock)
		if err != nil {
			beego.Error(err)
			break
		}

		// 区块详情
		result := respblock["result"].(map[string]interface{})
		prevBlockInfo := parse_block(result, false, height)
		if prevBlockInfo == nil {
			log.Debug("block existern !")
			continue
		}

		// update db
		//log.Debug(prevBlockInfo.Height, prevBlockInfo.Confirmations, prevBlockInfo.NextBlockHash)
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
