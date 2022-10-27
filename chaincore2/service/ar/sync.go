package ar

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	_ "encoding/json"
	"fmt"
	api "github.com/Dev43/arweave-go/api"
	"github.com/Dev43/arweave-go/utils"
	"github.com/astaxie/beego/httplib"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daoar"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/common/log"
	"github.com/shopspring/decimal"
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
// 同步链数据
func SyncData() bool {
	starttime := common.GetMillTime()
	endtime := starttime + 6*1000

	// get fee
	if time.Now().Unix() > LoadTxFeeTime {
		TxFee = GetTxFee()
		LoadTxFeeTime += int64(24 * 3600)
	}

	client, err := api.Dial(beego.AppConfig.String("nodeurl"))
	//client, err := api.Dial("http://ar.rylink.io:1984")//
	if err != nil {
		beego.Error(err)
		time.Sleep(time.Millisecond * 700)
		return true
	}

	info, err := client.GetInfo(context.TODO())
	if info == nil || err != nil {
		beego.Error(err)
		time.Sleep(time.Millisecond * 700)
		return true
	}

	// 获取节点区块高度
	blockcount := int64(info.Height)

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

	tmpcount := dbblockcount
	oncecount, _ := beego.AppConfig.Int("oncecount")
	for i := 0; i < oncecount; i++ {
		// 获取区块数据
		tmpval := tmpcount + 1
		block, err := client.GetBlockByHeight(context.TODO(), tmpval)
		//block, err := getBlockByHeight(tmpval)
		if err != nil {
			log.Errorf("执行ar高度：%d,错误：%s", tmpval, err.Error())
			break
		}
		//t1:= time.Now().Unix()
		//fmt.Println("ar/sync.go 212",t1)

		// 解析区块到数据
		log.Debug("start parse block to db index ", tmpval)
		err = parse_data_todb(client, block, tmpval)
		log.Debug("end parse block to db index ", tmpval)
		if err != nil {
			log.Error(err.Error())
			break
		}
		//fmt.Println("ar/sync.go 222，tmpval=",tmpval,",",time.Now().Unix(),",耗了",time.Now().Unix()-t1,"秒")

		if tmpval >= (blockcount - 1 - beego.AppConfig.DefaultInt64("delayheight", 12)) {
			break
		}

		tmpcount++

		currtime := common.GetMillTime()
		if currtime >= endtime {
			break
		}
	}
	//fmt.Println("ar/sync.go 235")

	currtime := common.GetMillTime()
	if (currtime + 10) < endtime {
		//time.Sleep(time.Millisecond * 100)
		time.Sleep(time.Second * 5)
	}

	return true
}

// 解析指定区块高度到db
//func SyncBlockData(tmpval int64) {
//	client := client.NewHTTP("tcp://"+GetNode(), "/websocket")
//	err := client.Start()
//	if err != nil {
//		// handle error
//	}
//	defer client.Stop()
//
//	respdata, err := client.Block(&tmpval)
//	if respdata == nil || err != nil {
//		beego.Error(err)
//		return
//	}
//
//	info, err := client.ABCIInfo()
//	if info == nil || err != nil {
//		beego.Error(err)
//		return
//	}
//
//	// 解析区块到数据
//	log.Debug("start parse block to db index ", tmpval)
//	err = parse_data_todb(respdata, client, info.Response.LastBlockHeight)
//	log.Debug("end parse block to db index ", tmpval)
//	if err != nil {
//		beego.Error(err)
//	}
//}

func worker_tx(id int, jobs <-chan string, results chan<- int, hash string, height int64, blockInfo *dao.BlockInfo) {
	count := len(jobs)
	offset := 0
	for i := 0; i < count; i++ {
		select {
		case txid := <-jobs:
			//log.Debug(txid, id)
			offset += 1

			client, err := api.Dial(beego.AppConfig.String("nodeurl"))
			if err != nil {
				log.Debug(err)
				continue
			}
			transaction, err := client.GetTransaction(context.TODO(), txid)
			if err != nil {
				log.Debug(err)
				continue
			}
			if "" == transaction.Target() { //如果目标地址为空则表明不是交易类型
				continue
			}

			err = parse_block_tx_todb(0, hash, height, txid, blockInfo)
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

	results <- 1
}

// 解析区块到数据库 result
func parse_data_todb(client *api.Client, block *api.Block, cmpheight int64) error {

	if block == nil {
		return nil
	}

	blockInfo := dao.NewBlockInfo()
	highindex := int64(block.Height)
	blockInfo.Height = highindex
	blockInfo.Timestamp = int64(block.Timestamp)
	blockInfo.FrontBlockHash = block.PreviousBlock
	blockInfo.Hash = block.IndepHash
	blockInfo.Transactions = int64(len(block.Txs))
	blockInfo.Confirmations = 1

	// 区块交易信息
	enablegoroutine := beego.AppConfig.DefaultBool("enablegoroutine", false)
	cpus := runtime.NumCPU()

	for i, v := range block.Txs {
		txid := fmt.Sprintf("%s", v)
		// 投递到通道
		if enablegoroutine {
			index := i % cpus
			JobsTaskList[index].Txids <- txid
		} else {
			// 获取原始交易信息
			//log.Debug(txid)

			err := parse_block_tx_todb(0, block.IndepHash, highindex, txid, blockInfo)
			if err != nil {
				log.Debug(err)
				return err
			}
		}
	}

	if enablegoroutine {
		// 开始执行任务
		for w := 0; w < cpus; w++ {
			go worker_tx(w, JobsTaskList[w].Txids, JobsTaskList[w].TxResult, block.IndepHash, highindex, blockInfo)
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
	if blockInfo.Confirmations < beego.AppConfig.DefaultInt64("confirmations", 6) {
		//go update_confirmations(blockInfo.Hash, blockInfo.Height)
		go update_confirmations(blockInfo.Hash, blockInfo.Height, client)
	}

	return nil
}

// 解析区块详情到数据库
func parse_block(block *api.Block, checkfind bool, cmpheight int64) *dao.BlockInfo {
	if block == nil {
		return nil
	}

	hash := block.IndepHash

	blockInfo := dao.NewBlockInfo()
	//if checkfind {
	//	num := block.GetBlockCountByHash(hash)
	//	if num > 0 {
	//		return nil
	//	}
	//}

	//log.Debug(result)
	blockInfo.Height = int64(block.Height)
	blockInfo.Hash = hash
	blockInfo.Confirmations = cmpheight - int64(block.Height) + 1
	blockInfo.Timestamp = int64(block.Timestamp)
	blockInfo.FrontBlockHash = block.PreviousBlock
	blockInfo.NextBlockHash = ""
	blockInfo.Transactions = int64(len(block.Txs))

	return blockInfo
}

// 解析交易信息到block_tx表
func parse_block_tx_todb(id int, hash string, index int64, tx string, blockInfo *dao.BlockInfo) error {

	//block, err := client.GetBlockByHeight(context.TODO(),tmpval)
	client, err := api.Dial(beego.AppConfig.String("nodeurl"))
	if err != nil {
		log.Debug(err)
		return err
	}
	transaction, err := client.GetTransaction(context.TODO(), tx)
	if err != nil {
		log.Debug(err)
		return err
	}
	if "" == transaction.Target() { //如果目标地址为空则表明不是交易类型
		return nil
	}

	// 获取详细交易数据
	blocktx := dao.NewBlockTX()
	//From地址
	data, err := base64.RawURLEncoding.DecodeString(transaction.Owner()) //
	h := sha256.New()
	h.Write(data)
	blocktx.From = utils.EncodeToBase64(h.Sum(nil))

	//转到哪而去
	blocktx.To = transaction.Target()
	blocktx.Hash = transaction.Hash()

	fee, err := decimal.NewFromString(transaction.Reward()) //获取手续费
	if err != nil {
		log.Debug(err)
		return err
	}
	blocktx.Sysfee = fee.Div(decimal.New(1, 12)).String()

	//获取交易金额
	quantity, err := decimal.NewFromString(transaction.Quantity())
	if err != nil {
		log.Debug(err)
		return err
	}
	blocktx.Amount = quantity.Div(decimal.New(1, 12)).String()
	blocktx.Height = blockInfo.Height //index
	blocktx.Hash = hash
	blocktx.Txid = utils.EncodeToBase64(transaction.ID())

	var tmpWatchList map[string]bool = make(map[string]bool)
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
		pushtx.Fee = common.StrToFloat64(blocktx.Sysfee)
		pushtx.From = blocktx.From
		pushtx.To = blocktx.To
		pushtx.Amount = blocktx.Amount
		pushtx.Memo = blocktx.Memo
		pushtx.Contract = blocktx.ContractAddress
		pushBlockTx.Txs = append(pushBlockTx.Txs, pushtx)

		pusdata, err := json.Marshal(&pushBlockTx)
		if err == nil {
			AddPushTask(blocktx.Height, blocktx.Txid, tmpWatchList, pusdata)
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

	return nil
}

func update_confirmations(hash string, height int64, client *api.Client) {
	// 更新确认数
	confirmations := beego.AppConfig.DefaultInt64("confirmations", 6)
	previousblockhash := hash
	for i := int64(0); i < confirmations; i++ {
		frontHeight := height - i - 1
		client2, err := api.Dial(beego.AppConfig.String("nodeurl"))
		if err != nil {
			log.Error(err.Error())
			continue
		}
		block, err := client2.GetBlockByHeight(context.TODO(), frontHeight)
		if block == nil || err != nil {
			log.Error(err.Error())
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

//
//
////====================私有化===================
//
//
//// Block struct
//type MyArBlock struct {
//	HashList      []string      `json:"hash_list"`
//	Nonce         string        `json:"nonce"`
//	PreviousBlock string        `json:"previous_block"`
//	Timestamp     int           `json:"timestamp"`
//	LastRetarget  int           `json:"last_retarget"`
//	Diff          string         `json:"diff"` //注意这里不再是int ，但是sdk没有更新
//	Height        int           `json:"height"`
//	Hash          string        `json:"hash"`
//	IndepHash     string        `json:"indep_hash"`
//	Txs           []interface{} `json:"txs"`
//	WalletList    []struct {
//		Wallet   string `json:"wallet"`
//		Quantity int64  `json:"quantity"`
//		LastTx   string `json:"last_tx"`
//	} `json:"wallet_list"`
//	RewardAddr string        `json:"reward_addr"`
//	Tags       []interface{} `json:"tags"`
//	RewardPool int           `json:"reward_pool"`
//	WeaveSize  int           `json:"weave_size"`
//	BlockSize  int           `json:"block_size"`
//}
//
//// GetBlockByHeight requests a block by its height
//func getBlockByHeight(height int64) (*api.Block, error) {
//
//	body, err := get(fmt.Sprintf(+"block/height/%d", height))
//	if err != nil {
//		return nil, err
//	}
//
//	//注意diff转换
//	result := MyArBlock{}
//	err = json.Unmarshal(body, &result)
//	if err != nil {
//		return nil, err
//	}
//	result.Diff = ""
//
//	body,_ = json.Marshal(result)
//	block := api.Block{}
//	err = json.Unmarshal(body, &body)
//	if err != nil {
//		return nil, err
//	}
//	return &block, nil
//}
//
//
//
//// 发送GET请求
//// url：         请求地址
//// response：    请求返回的内容
//func get(url string) ([]byte, error) {
//
//	// 超时时间：60秒
//	client := &http.Client{Timeout: 60 * time.Second}
//	resp, err := client.Get(url)
//	if err != nil {
//		return nil, err
//	}
//	defer resp.Body.Close()
//	var buffer [512]byte
//	result := bytes.NewBuffer(nil)
//	for {
//		n, err := resp.Body.Read(buffer[0:])
//		result.Write(buffer[0:n])
//		if err != nil && err == io.EOF {
//			break
//		} else if err != nil {
//			return nil, err
//		}
//	}
//
//	return result.Bytes(), nil
//}
