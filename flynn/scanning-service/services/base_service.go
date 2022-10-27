package services

import (
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/scanning-service/common"
	"github.com/group-coldwallet/scanning-service/conf"
	"github.com/group-coldwallet/scanning-service/log"
	"github.com/group-coldwallet/scanning-service/models"
	"github.com/group-coldwallet/scanning-service/models/po"
	"github.com/shopspring/decimal"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var c = make(chan os.Signal, 10)

//var MultiJobsTaskList []*MultiTask
//var DispossHeightMap map[int64]int = make(map[int64]int) // key:height,value(0:已分配,1:成功，2：失败)
//var HeightResults chan *HeightResultInfo = make(chan *HeightResultInfo, 10000)
var LatestBlockHeight int64

//type MultiTask struct {
//	WockerID       int
//	HeightList     chan int64 // 高度列表
//	HeightListExit chan int
//}
//type HeightResultInfo struct {
//	Height int64
//	Result int //(1:成功，2：失败)
//}

type BaseService struct {
	scan                common.IScanner      // 扫描服务
	Watcher             *common.WatchControl // 监听地址服务
	Cfg                 *conf.Config
	initHeight          int64
	workNum, maxTaskNum int64
	confirmationPool    *sync.Map //确认推送map

}

func NewBaseService(cfg *conf.Config, scan common.IScanner, watcher *common.WatchControl) *BaseService {
	bs := new(BaseService)
	bs.Cfg = cfg
	bs.scan = scan
	bs.Watcher = watcher
	bs.confirmationPool = &sync.Map{}
	return bs
}

/*
初始化服务
*/
func (bs *BaseService) Init() {
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGTRAP, syscall.SIGHUP, syscall.SIGQUIT)
	//初始化推送服务
	bs.RunPush()
	//if bs.Cfg.Sync.EnableSync {
	//	// 初始化多线程扫描
	//	bs.initMultiThreadService()
	//}else{
	//	//初始化单线程扫描
	//	bs.initHeight = bs.Cfg.Sync.InitHeight
	//	bs.workNum = bs.Cfg.Sync.MultiScanTaskNum
	//	bs.maxTaskNum = bs.Cfg.Sync.MultiScanNum
	//	bs.rollback()
	//}
	//初始化单线程扫描
	bs.initHeight = bs.Cfg.Sync.InitHeight
	bs.workNum = bs.Cfg.Sync.MultiScanTaskNum
	bs.maxTaskNum = bs.Cfg.Sync.MultiScanNum
	bs.rollback()
}

func (bs *BaseService) Start() {
	go func() {
		run := true
		for true {
			select {
			case stop := <-c:
				log.Info("exit", stop)
				run = false
				break
			default:
				break
			}
			if !run {
				break
			}
			result := true
			// 同步区块
			//if bs.Cfg.Sync.EnableSync {
			//	result = bs.multiScanBlock()
			//}else{
			//	result = bs.singleScanBlock()
			//}
			result = bs.singleScanBlock()
			if !result {
				break
			}
			//更新确认数
			//bs.update_confirmations()
		}
		//// 退出多线程
		//if bs.Cfg.Sync.EnableSync {
		//	for i:=0;i<int(bs.workNum);i++ {
		//		MultiJobsTaskList[i].HeightList <- int64(-1)
		//	}
		//	// wait task stop
		//	for i := 0; i<int(bs.workNum); i++ {
		//		<-MultiJobsTaskList[i].HeightListExit
		//	}
		//}
		bs.StopPush()
		//bs.Exit()
		os.Exit(1)
	}()
	go bs.update_confirmations()
}

func (bs *BaseService) Exit() {
	//if bs.Cfg.Sync.EnableSync{
	//	for i := 0; i < int(bs.workNum); i++ {
	//		close(MultiJobsTaskList[i].HeightList)
	//		close(MultiJobsTaskList[i].HeightListExit)
	//	}
	//}
}

func (bs *BaseService) parseBlockHeight(height int64) error {
	var (
		blockData *common.BlockData
		txData    *common.TxData
		err       error
	)
	//根据高度获取区块信息
	blockData, err = bs.scan.GetBlockByHeight(height)
	if err != nil {
		return fmt.Errorf("get block data error,err: %v", err)
	}

	blockInfo := &po.BlockInfo{
		Height:         blockData.Height,
		Hash:           blockData.Hash,
		FrontBlockHash: blockData.PrevHash,
		NextBlockHash:  blockData.NextHash,
		Timestamp:      common.Int64ToTime(blockData.Timestamp),
		Transactions:   blockData.TxNums,
		Confirmations:  blockData.Confirmation,
		CreateTime:     time.Now(),
	}

	num, err := po.InsertBlockInfo(blockInfo)

	if num <= 0 || err != nil {
		log.Error(err)
		return err
	}

	if len(blockData.TxIds) > 0 {
		for _, txid := range blockData.TxIds {
			// 处理交易
			log.Infof("开始处理txId:%s[高度:%d].", txid, blockData.Height)
			txData, err = bs.scan.GetTxData(blockData, txid, bs.isWatchAddress, bs.isContractTx)
			if err != nil {
				//log.Infof("get transaction error,txid=%s,err: %v",txid,err)
				log.Infof("txId:%s[高度:%d]处理失败. Err: %s", txid, blockData.Height, err.Error())
				continue
			}
			// 表示没有找到相关的交易
			if txData == nil {
				log.Infof("txId:%s[高度:%d] txData is null", txid, blockData.Height)
				continue
			}
			//log.Infof("txId:%s[高度:%d]处理完成. 结果: %s",txid, blockData.Height,utils.DumpJSON(txData))
			// 包含需要监听的交易才处理
			if txData.IsContainTx {
				//if txData.IsFakeTx {
				//	log.Errorf("发现一笔假充值,txid=%s",txid)
				//	continue
				//}
				//处理入账
				err = bs.processingTx(blockData, txData)
				if err != nil {
					log.Errorf("处理入账失败：%v", err)
				}
			}
		}
		return nil
	}
	//------------------------------------------------------------------------------------------------------------------//
	/*
		如果在这里处理，就得在这里判断是否有坚听的地址和合约了
		目前处理币种： xtz
	*/
	if len(blockData.TxDatas) > 0 {
		for _, td := range blockData.TxDatas {
			var coinDecimal int32
			//第一步，判断是否有监听的合约
			if td.ContractAddress != "" {
				//contractInfo, isHaveContract := bs.isContractTx(td.ContractAddress)
				//if !isHaveContract {
				//	//没这监听这个合约
				//	continue
				//}
				coinDecimal = int32(18)
			} else {
				coinDecimal = td.MainDecimal
			}

			//2. ******更改amount以及fee的精度**********
			amount, _ := decimal.NewFromString(td.Amount)
			td.Amount = amount.Shift(-coinDecimal).String()
			fee, _ := decimal.NewFromString(td.Fee)
			td.Fee = fee.Shift(-td.MainDecimal).String()
			err = bs.processingTx(blockData, txData)
			if err != nil {
				log.Errorf("处理入账失败：%v", err)
			}
		}
		return nil
	}
	//------------------------------------------------------------------------------------------------------------------//

	return nil
}

func (bs *BaseService) processingTx(blockData *common.BlockData, txData *common.TxData) error {
	var (
		pushTxs      []models.PushAccountTx
		tmpWatchList map[string]bool = make(map[string]bool)
	)
	// 假充值金额设置为0，但是如果是自己出现了交易失败，要把手续费发送上去
	isFakeTx := false //默认是正常交易
	if txData.IsFakeTx || txData.Amount == "" || txData.Amount == "0" || txData.Amount == "null" {
		isFakeTx = true
	}
	// 保证充值不是假充值
	if txData.ToAddr != "" && bs.Watcher.IsWatchAddressExist(txData.ToAddr) && !isFakeTx {
		tmpWatchList[txData.ToAddr] = true
	}
	if txData.FromAddr != "" && bs.Watcher.IsWatchAddressExist(txData.FromAddr) {
		// 假充值的时候需要把from地址的手续费推送上去，但是如果手续费为0的话，就不推送了
		if isFakeTx {
			if txData.Fee != "" {
				tmpWatchList[txData.FromAddr] = true //有手续费的时候才推送
			}
		} else {
			tmpWatchList[txData.FromAddr] = true
		}
	}
	if tmpWatchList[txData.FromAddr] || tmpWatchList[txData.ToAddr] {
		blockTx := &po.BlockTX{
			Height:          blockData.Height,
			Hash:            blockData.Hash,
			Txid:            txData.Txid,
			From:            txData.FromAddr,
			To:              txData.ToAddr,
			Amount:          txData.Amount,
			SysFee:          txData.Fee,
			Memo:            txData.Memo,
			ContractAddress: txData.ContractAddress,
		}
		var pushtx models.PushAccountTx
		pushtx.From = blockTx.From
		pushtx.To = blockTx.To
		pushtx.Amount = blockTx.Amount
		pushtx.Fee = blockTx.SysFee
		pushtx.Txid = blockTx.Txid
		pushtx.Memo = blockTx.Memo
		pushtx.Contract = blockTx.ContractAddress
		pushTxs = append(pushTxs, pushtx)
		//插入到表中
		_, err := po.InsertBlockTX(blockTx)
		if err != nil {
			log.Error(err)

		}
	}
	if len(tmpWatchList) > 0 {
		pushBlockTx := new(models.PushAccountBlockInfo)
		pushBlockTx.Type = models.PushTypeAccountTX
		pushBlockTx.CoinName = bs.Cfg.Sync.Name
		pushBlockTx.Height = blockData.Height
		pushBlockTx.Hash = blockData.Hash
		confrimation := blockData.Confirmation
		isNeedPushAgain := false
		if confrimation >= bs.Cfg.Sync.Confirmations {
			confrimation = bs.Cfg.Sync.Confirmations + 1
		} else {
			//如果小于推送需要的确认数，那么从1开始推送确认数
			confrimation = 1
			isNeedPushAgain = true
		}
		pushBlockTx.Confirmations = confrimation
		pushBlockTx.Time = common.Int64ToTime(blockData.Timestamp).Unix()
		pushBlockTx.Txs = pushTxs
		pushdata, err := json.Marshal(&pushBlockTx)
		if err != nil {
			log.Error(err)
		}
		if isNeedPushAgain {
			//存储到map中
			bs.confirmationPool.Store(txData.Txid, pushdata)
		}
		log.Infof("PushData: %s", string(pushdata))
		// 添加推送
		bs.AddPushTask(blockData.Height, txData.Txid, tmpWatchList, pushdata)
	}
	return nil
}

/*
判断是否是合约交易
*/
func (bs *BaseService) isContractTx(contractAddress string) (*po.ContractInfo, bool) {
	isExist := bs.Watcher.IsContractExist(contractAddress)
	if !isExist {
		return nil, false
	}
	ci, err := bs.Watcher.GetContract(contractAddress)
	if err != nil {
		log.Error(err)
		return nil, false
	}
	return ci, true
}

/*
判断是否是监听地址
*/
func (bs *BaseService) isWatchAddress(address string) bool {
	return bs.Watcher.IsWatchAddressExist(address)
}

func (bs *BaseService) singleScanBlock() bool {
	starttime := common.GetMillTime()
	endtime := starttime + 20*1000*bs.Cfg.Sync.SleepTime
	var (
		tmpCount, latestHeight, dbblockcount int64
		err                                  error
	)
	/*
		获取最新的高度
	*/
	latestHeight, err = bs.scan.GetLatestBlockHeight()
	if err != nil {
		log.Errorf("get latest block height error, err: %v", err)
		time.Sleep(time.Second * 3)
		return true
	}
	LatestBlockHeight = latestHeight
	// 获取db区块高度
	dbblockcount, err = po.GetMaxBlockIndex()
	if err != nil {
		log.Errorf("get db block height error, err: %v", err)
		return true
	}
	/*
		在某一个高度停止服务
	*/
	if bs.Cfg.Sync.EnableStop {
		if dbblockcount >= bs.Cfg.Sync.StopHeight {
			log.Infof("扫描完成：终止高度为： %d", dbblockcount)
			os.Exit(1)
		}
	}
	//跳块处理
	if bs.initHeight > dbblockcount {
		dbblockcount = bs.initHeight
	}
	//log.Info("latestHeight: ",latestHeight, "dbBlockHeight: ",dbblockcount)
	if dbblockcount >= (latestHeight - bs.Cfg.Sync.DelayHeight) {
		time.Sleep(time.Second * time.Duration(bs.Cfg.Sync.SleepTime))
		return true
	}
	tmpCount = dbblockcount
	// 多线程处理
	if bs.Cfg.Sync.EnableSync {
		count := latestHeight - bs.Cfg.Sync.DelayHeight - dbblockcount
		if count > bs.maxTaskNum {
			tmpval := tmpCount + 1
			//多线程扫描
			return bs.BatchScanIrreverseBlocks(tmpval, latestHeight-bs.Cfg.Sync.DelayHeight)
		}
	}
	failNum := 0
	for i := 0; i < int(bs.maxTaskNum); i++ {
		tmpval := tmpCount + 1
		log.Infof("============> start process %d height", tmpval)
		err = bs.parseBlockHeight(tmpval)
		if err != nil {
			log.Errorf("处理区块%d错误，err：%v", tmpval, err)
			//重试三次
			failNum++
			if failNum < 3 {
				time.Sleep(time.Second * time.Duration(bs.Cfg.Sync.SleepTime))
				continue
			} else {
				tmpCount++
				failNum = 0
				continue
			}
		}
		log.Infof("end process %d height <===============", tmpval)
		if tmpval > latestHeight-1-bs.Cfg.Sync.DelayHeight {
			break
		}
		tmpCount++
		currtime := common.GetMillTime()
		if currtime >= endtime {
			break
		}
	}
	currtime := common.GetMillTime()
	if (currtime + 20*bs.Cfg.Sync.SleepTime) < endtime {
		time.Sleep(time.Second * time.Duration(bs.Cfg.Sync.SleepTime))
	}
	return true
}

func (bs *BaseService) update_confirmations() {
	run := true
	for true {
		select {
		case <-c:
			run = false
			break
		default:
			break
		}
		if !run {
			break
		}
		//1. 判断内存中是否有需要更新的交易
		bs.confirmationPool.Range(func(key, value interface{}) bool {
			data := value.([]byte)
			if data == nil {
				return false
			}
			var pushTx models.PushAccountBlockInfo
			err := json.Unmarshal(data, &pushTx)
			if err != nil {
				log.Errorf("json unmarshal pushTx error: %v", err)
				return false
			}
			height, confirms := pushTx.Height, pushTx.Confirmations
			newConfirms := LatestBlockHeight - height
			newPushTx := new(models.PushAccountBlockInfo)
			newPushTx.Height = pushTx.Height
			newPushTx.Hash = pushTx.Hash
			newPushTx.CoinName = pushTx.CoinName
			newPushTx.Type = models.PushTypeAccountConfir
			if newConfirms > confirms && newConfirms < bs.Cfg.Sync.Confirmations {
				for i := confirms; i < newConfirms; i++ {
					newPushTx.Confirmations = i + 1
					newPushTx.Time = time.Now().Unix()
					pushData, _ := json.Marshal(newPushTx)
					height++
					bs.AddPushUserTask(height, pushData)
				}
				//再次存入map中
				pushTx.Confirmations = newConfirms
				pd, _ := json.Marshal(pushTx)
				bs.confirmationPool.Store(key, pd)
			}
			if newConfirms >= bs.Cfg.Sync.Confirmations {
				for i := confirms; i < bs.Cfg.Sync.Confirmations; i++ {
					newPushTx.Confirmations = i + 1
					newPushTx.Time = time.Now().Unix()
					pushData, _ := json.Marshal(newPushTx)
					height++
					bs.AddPushUserTask(height, pushData)
					// 最后一个确认数要判断交易是否在链上，避免假充值
					if i == bs.Cfg.Sync.Confirmations-1 {
						//如果存在链上，就推送，不存在就不推送
						if bs.scan.GetTxIsExist(pushTx.Height, key.(string)) {
							bs.AddPushUserTask(height, pushData)
						}
					}
				}
				//将这笔交易从内存中移除
				bs.confirmationPool.Delete(key)
			}
			return true
		})
		time.Sleep(time.Second * time.Duration(bs.Cfg.Sync.SleepTime*30))
	}
	confirmation := bs.Cfg.Sync.Confirmations
	bis, err := po.GetUnconfirmBlockInfos(confirmation)
	if err == nil && len(bis) > 0 {
		var ids []int64
		for _, bi := range bis {
			ids = append(ids, bi.Id)
			pushBlockTx := new(models.PushAccountBlockInfo)
			pushBlockTx.Type = models.PushTypeAccountConfir
			pushBlockTx.Height = bi.Height
			pushBlockTx.Hash = bi.Hash
			pushBlockTx.CoinName = bs.Cfg.Sync.Name
			pushBlockTx.Confirmations = bi.Confirmations + 1
			pushBlockTx.Time = bi.Timestamp.Unix()
			pusdata, err := json.Marshal(&pushBlockTx)
			if err == nil {
				bs.AddPushUserTask(bi.Height, pusdata)
			}
		}
		err = po.BatchUpdateConfirmations(ids, 1)
		if err != nil {
			log.Errorf("update confirmation error")
		}
	}
}

/*
回滚处理
*/
func (bs *BaseService) rollback() {
	if bs.Cfg.Sync.EnableRollback {
		rollHeight := bs.Cfg.Sync.RollHeight
		if rollHeight <= 0 {
			dbBlockCount, err := po.GetMaxBlockIndex()
			if err != nil || dbBlockCount <= 0 {
				log.Errorf("回滚失败： err：%v", err)
				return
			}
			rollHeight = dbBlockCount
		}
		log.Infof("开始回滚区块数据： 回滚到： %d", rollHeight)
		err := po.DeleteBlockInfo(rollHeight)
		if err != nil {
			log.Errorf("回滚失败： %v", err)
			return
		}
		err = po.DeleteBlockTX(rollHeight)
		if err != nil {
			log.Errorf("回滚失败： %v", err)
			return
		}
		bs.initHeight = rollHeight
	}
	bs.initHeight = bs.Cfg.Sync.InitHeight
}

func (bs *BaseService) BatchScanIrreverseBlocks(startHeight, endHeight int64) bool {
	starttime := time.Now()
	count := endHeight - startHeight
	if count > 100 {
		count = 100 //限制每次最大只处理100个
	}
	wg := &sync.WaitGroup{}
	wg.Add(int(count))
	for i := int64(0); i < count; i++ {
		height := startHeight + i
		go func(w *sync.WaitGroup) {
			if err := bs.parseBlockHeight(height); err != nil {
				log.Errorf("多线程处理区块%d错误： %v", height, err)
			}
			w.Done()
		}(wg)
	}
	wg.Wait()
	log.Infof("***batchScanBlocks used time : %f 's", time.Since(starttime).Seconds())
	return true
}
