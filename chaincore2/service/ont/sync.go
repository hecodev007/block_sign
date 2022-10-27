package ont

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daoont"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/common/log"
	"github.com/itchyny/base58-go"
	"github.com/shopspring/decimal"
	"math/big"
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
var ValuePrecision float64 = 1000000000.0

// ong
func GetValue(value float64) float64 {
	_value, _ := strconv.ParseFloat(fmt.Sprintf("%.9f", value/ValuePrecision), 64)
	return _value
}

// ong
func GetValueStr(value float64) string {
	return fmt.Sprintf("%.9f", value/ValuePrecision)
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
	if err != nil || datas["desc"] != "SUCCESS" {
		log.Debug(err, string(respdata))
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

	if dbblockcount >= (blockcount - 1) {
		time.Sleep(time.Millisecond * 500)
		return true
	}
	log.Debug(blockcount, dbblockcount)

	tmpcount := dbblockcount
	oncecount, _ := beego.AppConfig.Int("oncecount")
	for i := 0; i < oncecount; i++ {
		// 获取区块数据
		tmpheight := tmpcount + 1
		respdata, err := common.RequestStr("getblock", []interface{}{tmpheight, 1})
		if err != nil {
			beego.Error(err)
			break
		} else {
			//log.Debug(respdata)
		}

		// 解析区块到数据
		log.Debug("start parse block to db index ", tmpheight)
		err = parse_data_todb(respdata, tmpheight)
		log.Debug("end parse block to db index ", tmpheight)
		if err != nil {
			beego.Error(err)
			break
		}

		if tmpheight >= (blockcount - 1) {
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
			respdata, err := common.Request("getsmartcodeevent", []interface{}{txid})
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
			err = parse_block_tx_todb(id, hash, highindex, tx, blockInfo)
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
func parse_data_todb(blockdata string, parseheight int64) error {
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
	blockInfo := parse_block(result, true, parseheight)
	if blockInfo == nil {
		log.Debug("block existern !")
		return errors.New("block existern !")
	}

	// 区块交易信息
	highindex, hash := blockInfo.Height, blockInfo.Hash
	enablegoroutine := beego.AppConfig.DefaultBool("enablegoroutine", false)
	cpus := runtime.NumCPU()
	txs := result["Transactions"].([]interface{})
	for i := 0; i < len(txs); i++ {
		_tx := txs[i].(map[string]interface{})
		txid := _tx["Hash"].(string)

		// 投递到通道
		if enablegoroutine {
			index := i % cpus
			JobsTaskList[index].Txids <- txid
		} else {
			// 获取原始交易信息
			log.Debug(txid)
			var datas map[string]interface{}

			// 解析原始交易信息
			respdata, err := common.Request("getsmartcodeevent", []interface{}{txid})
			if err != nil {
				beego.Error(err)
				return err
			} else {
				//log.Debug(string(respdata))
			}

			err = json.Unmarshal(respdata, &datas)
			if err != nil {
				log.Debug(err)
				return err
			}

			tx := datas["result"].(map[string]interface{})
			err = parse_block_tx_todb(0, hash, highindex, tx, blockInfo)
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
		go update_confirmations(blockInfo.FrontBlockHash, blockInfo.Height)
	}

	return nil
}

// 解析区块详情到数据库
func parse_block(result map[string]interface{}, checkfind bool, cmpheight int64) *dao.BlockInfo {
	if result == nil {
		return nil
	}

	hash := result["Hash"].(string)

	block := dao.NewBlockInfo()
	if checkfind {
		num := block.GetBlockCountByHash(hash)
		if num > 0 {
			return nil
		}
	}

	blockhead := result["Header"].(map[string]interface{})
	block.Height = int64(blockhead["Height"].(float64))
	block.Hash = hash
	block.Confirmations = cmpheight - block.Height + 1
	block.Timestamp = int64(blockhead["Timestamp"].(float64))
	if blockhead["PrevBlockHash"] != nil {
		block.FrontBlockHash = blockhead["PrevBlockHash"].(string)
	}
	if blockhead["NextBlockHash"] != nil {
		block.NextBlockHash = blockhead["NextBlockHash"].(string)
	}
	block.Transactions = len(result["Transactions"].([]interface{}))

	return block
}

// 解析交易信息到db
func parse_block_tx_todb(id int, hash string, height int64, tx map[string]interface{}, blockInfo *dao.BlockInfo) error {
	if tx == nil {
		return nil
	}

	//log.Debug(tx)
	var tmpWatchList map[string]bool = make(map[string]bool)

	var pushtxs []models.PushAccountTx
	{
		blocktx := dao.NewBlockTX()
		blocktx.Height = height
		blocktx.Hash = hash
		blocktx.Txid = tx["TxHash"].(string)
		blocktx.Sysfee = GetValue(tx["GasConsumed"].(float64))
		blocktx.Status = int(tx["State"].(float64))

		notifys := tx["Notify"].([]interface{})
		for j := 0; j < len(notifys); j++ {
			notify := notifys[j].(map[string]interface{})
			contractAddr := notify["ContractAddress"].(string)
			if WatchContractList[contractAddr] != nil {
				States := notify["States"].([]interface{})

				// contract transfer
				//res, err := hex.DecodeString("7472616e73666572")
				//log.Debug(string(res))

				if len(States) == 4 && (States[0].(string) == "transfer" || States[0].(string) == "7472616e73666572") {
					blocktx.Contract = contractAddr
					if States[0].(string) == "transfer" {
						blocktx.From = States[1].(string)
						blocktx.To = States[2].(string)
						blocktx.Amount = int64(States[3].(float64))
					} else {
						blocktx.From = AddressFromHexString(States[1].(string))
						blocktx.To = AddressFromHexString(States[2].(string))
						res, _ := hex.DecodeString(States[3].(string))
						blocktx.Amount = BigIntFromNeoBytes(res).Int64()
					}
					blocktx.Insert()

					if blocktx.Status != 1 {
						continue
					}

					var watch bool = false
					if WatchAddressList[blocktx.From] != nil {
						log.Debug("watchaddr", blocktx.From)
						tmpWatchList[blocktx.From] = true
						watch = true
					}

					if WatchAddressList[blocktx.To] != nil {
						log.Debug("watchaddr", blocktx.To)
						tmpWatchList[blocktx.To] = true
						watch = true
					}

					if watch {
						var pushtx models.PushAccountTx
						pushtx.Txid = blocktx.Txid
						pushtx.Fee = blocktx.Sysfee
						pushtx.Contract = blocktx.Contract
						pushtx.From = blocktx.From
						pushtx.To = blocktx.To
						if blocktx.Contract == "0100000000000000000000000000000000000000" {
							pushtx.Amount = fmt.Sprintf("%d", blocktx.Amount)
						} else if blocktx.Contract == "0200000000000000000000000000000000000000" {
							pushtx.Amount = GetValueStr(float64(blocktx.Amount))
						} else {
							// contract
							pushtx.Amount = decimal.NewFromFloat(float64(blocktx.Amount)).Div(decimal.New(1, int32(WatchContractList[contractAddr].Decimal))).String()
						}
						pushtxs = append(pushtxs, pushtx)
					}
				}
			}
		}

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
		pushBlockTx.Txs = pushtxs

		pusdata, err := json.Marshal(&pushBlockTx)
		if err == nil {
			AddPushTask(height, hash, tmpWatchList, pusdata)
		} else {
			log.Debug(err)
		}
	}

	return nil
}

func update_confirmations(frontHash string, height int64) {
	// 更新确认数
	confirmations := beego.AppConfig.DefaultInt64("confirmations", 6)
	previousblockhash := frontHash
	for i := int64(0); i < confirmations; i++ {
		respdata, err := common.Request("getblock", []interface{}{previousblockhash, 1})
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
		prevBlockInfo := parse_block(result, false, height)
		if prevBlockInfo == nil {
			log.Debug("block existern !")
			continue
		}

		// update db
		//log.Debug(prevBlockInfo.Height, prevBlockInfo.Confirmations, prevBlockInfo.NextBlockHash)
		dao.UpdateConfirmations(prevBlockInfo.Height, prevBlockInfo.Confirmations, prevBlockInfo.NextBlockHash)

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

func ToArrayReverse(arr []byte) []byte {
	l := len(arr)
	x := make([]byte, 0)
	for i := l - 1; i >= 0; i-- {
		x = append(x, arr[i])
	}
	return x
}

// AddressParseFromHexString returns parsed Address
func AddressFromHexString(s string) string {
	hx, err := hex.DecodeString(s)
	if err != nil {
		log.Debug(err)
		return ""
	}
	if len(hx) != 20 {
		log.Debug("[Common]: AddressParseFromBytes err, len != 20")
		return ""
	}

	data := append([]byte{23}, hx[:]...)
	temp := sha256.Sum256(data)
	temps := sha256.Sum256(temp[:])
	data = append(data, temps[0:4]...)

	bi := new(big.Int).SetBytes(data).String()
	encoded, _ := base58.BitcoinEncoding.Encode([]byte(bi))
	return string(encoded)

	//addr, err := utils.AddressParseFromBytes(ToArrayReverse(hx))
	//addr, err := utils.AddressParseFromBytes(hx)
	//return addr.ToBase58(), nil
}

func bytesReverse(u []byte) []byte {
	for i, j := 0, len(u)-1; i < j; i, j = i+1, j-1 {
		u[i], u[j] = u[j], u[i]
	}
	return u
}

var bigOne = big.NewInt(1)

func BigIntFromNeoBytes(ba []byte) *big.Int {
	res := big.NewInt(0)
	l := len(ba)
	if l == 0 {
		return res
	}

	bytes := make([]byte, 0, l)
	bytes = append(bytes, ba...)
	bytesReverse(bytes)

	if bytes[0]>>7 == 1 {
		for i, b := range bytes {
			bytes[i] = ^b
		}

		temp := big.NewInt(0)
		temp.SetBytes(bytes)
		temp.Add(temp, bigOne)
		bytes = temp.Bytes()
		res.SetBytes(bytes)
		return res.Neg(res)
	}

	res.SetBytes(bytes)
	return res
}
