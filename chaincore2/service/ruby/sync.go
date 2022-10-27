package ruby

import (
	"encoding/json"
	"errors"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daoruby"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/common/log"
	"github.com/shopspring/decimal"
	"os"
	"os/signal"
	"runtime"
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
var ValuePrecision float64 = 1000000.0

func GetValue(value float64) float64 {
	//_value, _ := strconv.ParseFloat(fmt.Sprintf("%.6f", value / ValuePrecision), 64)
	//return _value
	amount := decimal.NewFromFloat(value)
	_value, _ := amount.Div(decimal.NewFromFloat(ValuePrecision)).Float64()
	return _value
}

func GetValueStr(value float64) string {
	amount := decimal.NewFromFloat(value)
	return amount.Div(decimal.NewFromFloat(ValuePrecision)).String()

	//return fmt.Sprintf("%.6f", value / ValuePrecision)
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
	var blockCountResult models.GetBlockCountResult
	err := RequestObject("showblockcount", nil, &blockCountResult)
	if err != nil {
		beego.Error(err)
		time.Sleep(time.Millisecond * 500)
		return true
	} else {
		//log.Debug(blockCountResult.Result)
	}
	blockcount := blockCountResult.Result //common.StrToInt64(blockCountResult.Result)

	// 获取db区块高度
	dbblockcount, err2 := dao.GetMaxBlockIndex()
	if err2 != nil {
		beego.Error(err2)
		time.Sleep(time.Millisecond * 500)
		return true
	}

	if dbblockcount >= (blockcount - beego.AppConfig.DefaultInt64("delayheight", 2)) {
		time.Sleep(time.Millisecond * 500)
		return true
	}
	log.Debug(blockcount, dbblockcount)

	tmpcount := dbblockcount
	oncecount, _ := beego.AppConfig.Int("oncecount")
	for i := 0; i < oncecount; i++ {
		// 获取区块数据
		tmpval := tmpcount + 1

		respdata, err := RequestStr("showblock", []interface{}{common.Int64ToString(tmpval, 10), 1})
		if err != nil {
			beego.Error(err)
			return true
		} else {
			//log.Debug(string(respdata))
		}

		// 解析区块到数据
		log.Debug("start parse block to db index ", tmpval)
		err = parse_data_todb(respdata)
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
	Rollback(tmpval)
	respdata, err := RequestStr("showblock", []interface{}{common.Int64ToString(tmpval, 10), 1})
	if err != nil {
		beego.Error(err)
		return
	} else {
		//log.Debug(string(respdata))
	}

	// 解析区块到数据
	log.Debug("start parse block to db index ", tmpval)
	err = parse_data_todb(respdata)
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

			var datas map[string]interface{}

			// 解析原始交易信息
			respdata, err := Request("showrawdeal", []interface{}{txid, 1})
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
			if datas["result"] == nil {
				log.Debug("showrawdeal not found", txid, "reindex = 1 and txindex = 1 ?")
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
func parse_data_todb(blockdata string) error {
	var datas map[string]interface{}
	err := json.Unmarshal([]byte(blockdata), &datas)
	if err != nil {
		log.Debug(err)
		return err
	}

	// 区块详情
	result := datas["result"].(map[string]interface{})

	highindex, hash := int64(result["height"].(float64)), result["hash"].(string)
	blockInfo := parse_block(result, true)
	if blockInfo == nil {
		log.Debug("block existern !")
		return errors.New("block existern !")
	}

	// 区块交易信息
	enablegoroutine := beego.AppConfig.DefaultBool("enablegoroutine", false)
	cpus := runtime.NumCPU()
	txs := result["tx"].([]interface{})
	for i := 0; i < len(txs); i++ {
		txid := txs[i].(string)

		// 投递到通道
		if enablegoroutine {
			index := i % cpus
			JobsTaskList[index].Txids <- txid
		} else {
			// 获取原始交易信息
			log.Debug(txid)
			respdata, err := Request("showrawdeal", []interface{}{txid, 1})
			if err != nil {
				beego.Error(err)
				return err
			} else {
				//log.Debug(string(respdata))
			}

			var datas map[string]interface{}
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
		go update_confirmations(blockInfo.FrontBlockHash)
	}

	return nil
}

// 解析区块详情到数据库
func parse_block(result map[string]interface{}, checkfind bool) *dao.BlockInfo {
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
	block.Height = int64(result["height"].(float64))
	block.Hash = hash
	block.Confirmations = int64(result["confirmations"].(float64))
	block.Timestamp = int64(result["time"].(float64))
	if result["prevblockhash"] != nil {
		block.FrontBlockHash = result["prevblockhash"].(string)
	}
	if result["nextblockhash"] != nil {
		block.NextBlockHash = result["nextblockhash"].(string)
	}
	block.Transactions = len(result["tx"].([]interface{}))

	return block
}

func disposs_vin(hash string, highindex int64, txid string, vin map[string]interface{}) (float64, string, *dao.BlockTXVin) {
	var vin_amount float64 = 0
	var watchaddr string = ""
	if vin == nil {
		return vin_amount, watchaddr, nil
	}
	//log.Debug(vin)
	txvin := dao.NewBlockTXVin()
	txvin.Height = highindex
	txvin.Hash = hash
	txvin.Txid = txid

	// prevhash previndex
	if vin["txid"] != nil {
		txvin.Vintxid = vin["txid"].(string)
	}
	if vin["vout"] != nil {
		txvin.VinVoutindex = int(vin["vout"].(float64))
	}

	num, err := txvin.Insert()
	if num <= 0 || err != nil {
		log.Debug(err)
		return vin_amount, watchaddr, nil
	}

	if vin["coinbase"] == nil {
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

			if txvout.AssetId != "" && txvout.AssetName != "" {
				txvin.AssetId = txvout.AssetId
				txvin.AssetName = txvout.AssetName
				txvin.AssetValue = txvout.AssetValue
				txvin.AssetSellTxid = txvout.AssetSellTxid
			}
		}
	}

	return vin_amount, watchaddr, txvin
}

func worker_tx_vin(id int, jobs <-chan interface{}, results chan<- float64, hash string, highindex int64, txid string) {
	//log.Debug(id, len(jobs))
	var vin_amount float64 = 0
	count := len(jobs)
	offset := 0
	for i := 0; i < count; i++ {
		select {
		case vininfo := <-jobs:
			offset += 1
			vin := vininfo.(map[string]interface{})

			// 处理vin
			disposs_vin(hash, highindex, txid, vin)

			//log.Debug(id, i, "vin stop")
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
	results <- vin_amount
}

// 处理vout
func disposs_vout(hash string, highindex int64, txid string, vout map[string]interface{}) (float64, string, *dao.BlockTXVout) {
	var vout_amount float64 = 0
	var watchaddr string = ""
	if vout == nil {
		return vout_amount, watchaddr, nil
	}

	voutn := int(vout["n"].(float64))
	scriptPubKey := vout["scriptPubKey"].(map[string]interface{})

	txvout := dao.NewBlockTXVout()
	txvout.Height = highindex
	txvout.Hash = hash
	txvout.Txid = txid
	txvout.Voutn = voutn
	txvout.Voutvalue = decimal.NewFromFloat(vout["value"].(float64)).Mul(decimal.NewFromFloat(ValuePrecision)).IntPart() //int64(vout["value"].(float64) * ValuePrecision) 这样会丢失精度
	vout_amount += float64(txvout.Voutvalue)

	if scriptPubKey["addresses"] != nil {
		addresses := scriptPubKey["addresses"].([]interface{})
		if len(addresses) == 1 {
			txvout.Voutaddress = addresses[0].(string)
		}
	}

	// token
	if vout["assets"] != nil {
		assets := vout["assets"].([]interface{})
		if len(assets) > 0 {
			asset := assets[0].(map[string]interface{})
			if asset["type"].(string) == "transfer" {
				txvout.AssetName = asset["name"].(string)
				txvout.AssetSellTxid = asset["selltxid"].(string)
				txvout.AssetId = asset["assetref"].(string)
				txvout.AssetValue = int64(asset["raw"].(float64))
			}
		}
	}

	if scriptPubKey["type"].(string) == "nulldata" || (scriptPubKey["type"].(string) == "nonstandard" && scriptPubKey["addresses"] == nil) {
		return vout_amount, watchaddr, nil
	}

	txvout.Invaild = 0
	txvout.Status = common.Confirmed
	num, err := txvout.Insert()
	if num <= 0 || err != nil {
		log.Debug(err)
		return vout_amount, watchaddr, nil
	}

	// 添加或设置用户资产
	if txvout.Voutaddress != "" && txvout.Voutvalue > 0 {
		dao.FindSetBlockAmount(txvout.Voutaddress, txvout.Voutvalue)
	}
	watchaddr = txvout.Voutaddress
	return vout_amount, watchaddr, txvout
}

func worker_tx_vout(id int, jobs <-chan interface{}, results chan<- float64, hash string, highindex int64, txid string) {
	//log.Debug(id, len(jobs))
	var vout_amount float64 = 0
	count := len(jobs)
	offset := 0
	for i := 0; i < count; i++ {
		select {
		case voutinfo := <-jobs:
			//log.Debug(id, i, "vout start")
			vout := voutinfo.(map[string]interface{})

			// 处理vou
			disposs_vout(hash, highindex, txid, vout)

			//log.Debug(id, i, "vout stop")
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
	results <- vout_amount
}

// 解析交易信息到db
func parse_block_tx_todb(id int, hash string, index int64, tx map[string]interface{}, blockInfo *dao.BlockInfo) error {
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
	blocktx.Txid = tx["txid"].(string)
	blocktx.Sysfee = 0
	blocktx.Vincount = len(tx["vin"].([]interface{}))
	blocktx.Voutcount = len(tx["vout"].([]interface{}))

	enablegoroutine := beego.AppConfig.DefaultBool("enablegoroutine", false)
	vinvoutgoroutine := beego.AppConfig.DefaultBool("vinvoutgoroutine", false)
	cpus := runtime.NumCPU()
	var vout_amount float64 = 0.0
	if tx["vout"] != nil {
		//log.Debug(tx["vout"])
		vouts := tx["vout"].([]interface{})
		if len(vouts) > 0 {
			if enablegoroutine && vinvoutgoroutine {
				for j := 0; j < len(vouts); j++ {
					vout := vouts[j].(map[string]interface{})

					// 投递到通道
					chanindex := j % cpus
					JobsTaskList[id].Vouts[chanindex] <- vout
				}

				// 开始执行任务
				for w := 0; w < cpus; w++ {
					go worker_tx_vout(w, JobsTaskList[id].Vouts[w], JobsTaskList[id].VoutResult[w], hash, index, blocktx.Txid)
				}
			} else {
				for j := 0; j < len(vouts); j++ {
					vout := vouts[j].(map[string]interface{})
					tmpamount, watchaddr, blockvout := disposs_vout(hash, index, blocktx.Txid, vout)
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

		}
	}

	var vin_amount float64 = 0.0
	if tx["vin"] != nil {
		//log.Debug(tx["vin"])
		vins := tx["vin"].([]interface{})
		if len(vins) > 0 {
			if enablegoroutine && vinvoutgoroutine {
				for j := 0; j < len(vins); j++ {
					vin := vins[j].(map[string]interface{})

					// 投递到通道
					chanindex := j % cpus
					JobsTaskList[id].Vins[chanindex] <- vin
				}

				// 开始执行任务
				for w := 0; w < cpus; w++ {
					go worker_tx_vin(w, JobsTaskList[id].Vins[w], JobsTaskList[id].VinResult[w], hash, index, blocktx.Txid)
				}
			} else {
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
		}
	}

	if enablegoroutine && vinvoutgoroutine {
		for a := 0; a < cpus; a++ {
			tmpamount := <-JobsTaskList[id].VinResult[a]
			vin_amount += tmpamount
			//log.Debug(a, "vin finish")
		}

		for a := 0; a < cpus; a++ {
			tmpamount := <-JobsTaskList[id].VoutResult[a]
			vout_amount += tmpamount
			//log.Debug(a, "vou finish")
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
		for i := 0; i < len(blockvin_list); i++ {
			// checkout address
			if blockvin_list[i].Address == "" && blockvin_list[i].Amount == 0 {
				for true {
					// 解析原始交易信息
					respdata, err := Request("showrawdeal", []interface{}{blockvin_list[i].Vintxid, 1})
					if err != nil {
						beego.Error(err)
						break
					} else {
						//log.Debug(string(respdata))
					}

					var datas map[string]interface{}
					err = json.Unmarshal(respdata, &datas)
					if err != nil {
						log.Debug(err)
						break
					}
					if datas["result"] == nil {
						log.Debug("showrawdeal not found", blockvin_list[i].Vintxid, "reindex = 1 and txindex = 1 ?")
						break
					}
					tx := datas["result"].(map[string]interface{})
					tmpvouts := tx["vout"].([]interface{})
					if tmpvouts != nil && tmpvouts[blockvin_list[i].VinVoutindex] != nil {
						vout := tmpvouts[blockvin_list[i].VinVoutindex].(map[string]interface{})
						blockvin_list[i].Amount = decimal.NewFromFloat(vout["value"].(float64)).Mul(decimal.NewFromFloat(ValuePrecision)).IntPart()
						scriptPubKey := vout["scriptPubKey"].(map[string]interface{})
						if scriptPubKey["addresses"] != nil {
							addresses := scriptPubKey["addresses"].([]interface{})
							if len(addresses) == 1 {
								blockvin_list[i].Address = addresses[0].(string)
								tmpWatchList[blockvin_list[i].Address] = true
							}
						}
						vin_amount += float64(blockvin_list[i].Amount)

						// token
						if vout["assets"] != nil {
							assets := vout["assets"].([]interface{})
							if len(assets) > 0 {
								asset := assets[0].(map[string]interface{})
								if asset["type"].(string) == "transfer" {
									blockvin_list[i].AssetName = asset["name"].(string)
									blockvin_list[i].AssetSellTxid = asset["selltxid"].(string)
									blockvin_list[i].AssetId = asset["assetref"].(string)
									blockvin_list[i].AssetValue = int64(asset["raw"].(float64))
								}
							}
						}
					}
					break
				}
			}
			var assets *models.AssetsInfo = nil
			if blockvin_list[i].AssetId != "" && blockvin_list[i].AssetName != "" && WatchContractList[blockvin_list[i].AssetId] != nil {
				// token
				amount := decimal.NewFromFloat(float64(blockvin_list[i].AssetValue))
				_amount := amount.Div(decimal.New(1, int32(WatchContractList[blockvin_list[i].AssetId].Decimal))).String()
				assets = new(models.AssetsInfo)
				assets.AssetId = blockvin_list[i].AssetId
				assets.Name = blockvin_list[i].AssetName
				assets.AssetValue = _amount
			}

			value := GetValueStr(float64(blockvin_list[i].Amount))
			pushtx.Vin = append(pushtx.Vin, models.PushTxInput{Txid: blockvin_list[i].Vintxid, Vout: blockvin_list[i].VinVoutindex, Addresse: blockvin_list[i].Address, Value: value, Assets: assets})
		}
		for i := 0; i < len(blockvout_list); i++ {
			var assets *models.AssetsInfo = nil
			if blockvout_list[i].AssetId != "" && blockvout_list[i].AssetName != "" && WatchContractList[blockvout_list[i].AssetId] != nil {
				// token
				amount := decimal.NewFromFloat(float64(blockvout_list[i].AssetValue))
				_amount := amount.Div(decimal.New(1, int32(WatchContractList[blockvout_list[i].AssetId].Decimal))).String()
				assets = new(models.AssetsInfo)
				assets.AssetId = blockvout_list[i].AssetId
				assets.Name = blockvout_list[i].AssetName
				assets.AssetValue = _amount
			}
			value := GetValueStr(float64(blockvout_list[i].Voutvalue))
			pushtx.Vout = append(pushtx.Vout, models.PushTxOutput{Addresse: blockvout_list[i].Voutaddress, Value: value, N: blockvout_list[i].Voutn, Assets: assets})
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
		}
	}

	num, err := blocktx.Insert()
	if num <= 0 || err != nil {
		beego.Error(err)
	}
	return nil
}

func update_confirmations(frontHash string) {
	// 更新确认数
	confirmations := beego.AppConfig.DefaultInt64("confirmations", 6)
	previousblockhash := frontHash
	for i := int64(0); i < confirmations; i++ {
		respdata, err := RequestStr("showblock", []interface{}{previousblockhash, 1})
		if err != nil {
			beego.Error(err.Error())
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
		prevBlockInfo := parse_block(result, false)
		if prevBlockInfo == nil {
			log.Debug("block existern !")
			continue
		}

		// rollback
		{
			block := dao.NewBlockInfo()
			block.GetBlockInfoByIndex(prevBlockInfo.Height)
			if block.Hash != prevBlockInfo.Hash {
				// 通知分叉
				forkdata := map[string]interface{}{
					"type":   models.PushTypeChainFork,
					"coin":   beego.AppConfig.String("coin"),
					"height": prevBlockInfo.Height, // 分叉高度
					"hash":   block.Hash,           // 老区块hash
				}
				pushdata, err := json.Marshal(&forkdata)
				if err == nil {
					AddPushUserTask(prevBlockInfo.Height, pushdata)
				}

				// 删除老数据
				Rollback(prevBlockInfo.Height)

				// 写区块数据
				{
					txs := result["tx"].([]interface{})
					for i := 0; i < len(txs); i++ {
						txid := txs[i].(string)

						// 获取原始交易信息
						log.Debug(txid)
						respdata, err := Request("showrawdeal", []interface{}{txid, 1})
						if err != nil {
							beego.Error(err)
							continue
						} else {
							//log.Debug(string(respdata))
						}

						var datas map[string]interface{}
						err = json.Unmarshal(respdata, &datas)
						if err != nil {
							log.Debug(err)
							continue
						}

						tx := datas["result"].(map[string]interface{})
						err = parse_block_tx_todb(0, prevBlockInfo.Hash, prevBlockInfo.Height, tx, prevBlockInfo)
						if err != nil {
							log.Debug(err)
							continue
						}
					}

					// 写入区块信息
					prevBlockInfo.InsertBlockInfo()
				}
			}
		}

		// update db
		//log.Debug(prevBlockInfo.Height, prevBlockInfo.Confirmations, prevBlockInfo.NextBlockHash)
		dao.UpdateConfirmations(prevBlockInfo.Height, prevBlockInfo.Confirmations, prevBlockInfo.NextBlockHash)

		pushBlockTx := new(models.PushUtxoBlockInfo)
		pushBlockTx.Type = models.PushTypeConfir
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
