package wtc

import (
	"encoding/hex"
	"errors"
	"hscsync/common"
	"hscsync/common/conf"
	"hscsync/common/log"
	dao "hscsync/models/po/yotta"
	"hscsync/services"
	"hscsync/utils"
	"hscsync/utils/dingding"
	"hscsync/utils/eth"
	rpc "hscsync/utils/eth"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type Scanner struct {
	*rpc.RpcClient
	lock  *sync.Mutex
	conf  conf.SyncConfig
	Watch *services.WatchControl
	//IrreverseBlock map[int64]common.ProcTask
	TaskCaches sync.Map
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	return &Scanner{
		RpcClient:  rpc.NewRpcClient(node.Url),
		lock:       &sync.Mutex{},
		conf:       conf.Sync,
		Watch:      watch,
		TaskCaches: sync.Map{},
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	_, err := dao.BlockRollBack(height)
	if err != nil {
		panic(err.Error())
	}
	_, err = dao.TxRollBack(height)
	if err != nil {
		panic(err.Error())
	}
}

func (s *Scanner) Init() error {
	if s.conf.EnableRollback {
		s.Rollback(s.conf.RollHeight)
	}
	return nil
}

func (s *Scanner) Clear() {
}

//var i = int64(14717020)

//获取最高区块高度
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i, nil
	count, err := s.BlockNumber() //获取到的是区块个数
	return count, err
}

func (s *Scanner) GetCurrentBlockHeight() (int64, error) {
	return dao.MaxBlockHeight()
}

//扫描一个可逆的区块
func (s *Scanner) ScanReverseBlock(height, bestHeight int64) (t common.ProcTask, err error) {
	//task, ok := s.IrreverseBlock[height]
	taskBest, ok := s.TaskCaches.Load(bestHeight)
	if !ok {
		taskBest, err = s.scanBlock(bestHeight, bestHeight)
		if err != nil {
			return nil, err
		}
		s.TaskCaches.Store(bestHeight, taskBest)
	}
	for h := bestHeight - 1; h >= height; h-- {
		//		log.Info(h)
		taskH, ok := s.TaskCaches.Load(h)
		if !ok {
			//			log.Info(h, bestHeight)
			taskH, err = s.scanBlock(h, bestHeight)
			if err != nil {
				return nil, err
			}
			s.TaskCaches.Store(h, taskH)
		} else {
			taskHer, ok := s.TaskCaches.Load(h + 1)
			if !ok {
				panic("")
			}
			if taskHer.(common.ProcTask).ParentHash() == taskH.(common.ProcTask).GetBlockHash() {
				break
			}
		}
	}
	taskInter, ok := s.TaskCaches.Load(height)
	if !ok {
		panic("")
	}
	taskInter.(common.ProcTask).SetBestHeight(bestHeight, false)
	return taskInter.(common.ProcTask), nil
	//taskInter, ok := s.TaskCaches.Load(height)
	//if ok {
	//	task := taskInter.(common.ProcTask)
	//	block, err := s.GetBlockByNumber(height, false)
	//	if err != nil {
	//		return nil, err
	//	}
	//	if block.Hash == task.GetBlockHash() {
	//		task.SetBestHeight(bestHeight, false)
	//		return task, nil
	//	}
	//}
	//task, err := s.scanBlock(height, bestHeight)
	//if err != nil {
	//	return task, err
	//}
	//s.TaskCaches.Store(height, task)
	//for i := height - 150; i < height-100; i++ {
	//	s.TaskCaches.Delete(i)
	//}
	//return task, nil
}

//扫描一个不可逆的区块
func (s *Scanner) ScanIrreverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	taskInter, ok := s.TaskCaches.Load(height)
	if ok {
		s.TaskCaches.Delete(height)
		task := taskInter.(common.ProcTask)
		block, err := s.GetBlockByNumber(height, false)
		if err != nil {
			return nil, err
		}
		if block.Hash == task.GetBlockHash() {
			task.SetBestHeight(bestHeight, true)
			return task, nil
		}
	}
	task, err := s.scanBlock(height, bestHeight)
	return task, err
}

func (s *Scanner) scanBlock(height, bestHeight int64) (common.ProcTask, error) {
	st := time.Now()
	//log.Infof("scanBlock %d ", height)
getblocknumber:
	block, err := s.GetBlockByNumber(height, true)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}
	if block.Number == 0 {
		goto getblocknumber
	}
	task := &ProcTask{
		BestHeight: bestHeight,
		Block: &dao.BlockInfo{
			Height:            block.Number,
			Hash:              block.Hash,
			Previousblockhash: block.ParentHash,
			Timestamp:         time.Unix(int64(block.Timestamp), 0),
			Transactions:      len(block.Transactions),
			Confirmations:     bestHeight - height + 1,
			Createtime:        time.Now(),
		},
	}

	if task.Block.Confirmations >= s.conf.Confirmations {
		task.Irreversible = true
	}
	workpool := utils.NewWorkPool(10) //一次性发太多请求会让节点窒息

	for i, _ := range block.Transactions {
		workpool.Incr()
		go func(tx *rpc.Transaction, task *ProcTask) {

			defer workpool.Dec()
			s.batchParseTx(tx, task)

		}(block.Transactions[i], task)

	}
	workpool.Wait()
	_ = st
	//log.Infof("scanBlock %v ,耗时 : %v", height, time.Since(st))
	return task, nil
}

//批量解析交易
func (s *Scanner) batchParseTx(tx *rpc.Transaction, task *ProcTask) {
	blockTxs, err := s.parseBlockTX(tx, task.Block)
	if blockTxs != nil {
		s.lock.Lock()
		defer s.lock.Unlock()
		task.TxInfos = append(task.TxInfos, blockTxs...)
	}
	if err != nil {
		//log.Info(tx.Hash, err.Error())
		//panic("")
	}
}

// 解析交易
func (s *Scanner) parseBlockTX(tx *rpc.Transaction, block *dao.BlockInfo) (blocktxs []*dao.BlockTx, err error) {

	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}
	//Watch := s.Watch
	//rpc := s.RpcClient
	//var blocktxs = make([]*dao.BlockTx, 0)
reGetTransactionReceipt:
	txReceipt, err := s.RpcClient.GetTransactionReceipt(tx.Hash)
	if err != nil {
		log.Info(err.Error())
		time.Sleep(time.Second * 2)
		goto reGetTransactionReceipt
	}

	if txReceipt == nil {
		blocktx := dao.BlockTx{
			CoinName:        conf.Cfg.Sync.CoinName,
			Txid:            tx.Hash,
			BlockHeight:     block.Height,
			ContractAddress: "",
			FromAddress:     tx.From,
			ToAddress:       tx.To,
			Amount:          decimal.NewFromBigInt(tx.Value, -18),
			Fee:             decimal.NewFromBigInt(big.NewInt(0).Mul(tx.GasPrice, big.NewInt(21000)), -18),
			BlockHash:       block.Hash,
			Status:          "success",
			Timestamp:       block.Timestamp,
		}
		blocktxs = append(blocktxs, &blocktx)
		return
	}

	//eth 转账交易
	if tx.Value.Cmp(big.NewInt(0)) > 0 && (s.Watch.IsWatchAddressExist(tx.From) || s.Watch.IsWatchAddressExist(tx.To)) {
		blocktx := dao.BlockTx{
			CoinName:        conf.Cfg.Sync.CoinName,
			Txid:            tx.Hash,
			BlockHeight:     block.Height,
			ContractAddress: "",
			FromAddress:     tx.From,
			ToAddress:       tx.To,
			Amount:          decimal.NewFromBigInt(tx.Value, -18),
			Fee:             decimal.NewFromBigInt(big.NewInt(0).Mul(tx.GasPrice, big.NewInt(txReceipt.GasUsed)), -18),
			BlockHash:       block.Hash,
			Status:          "success",
			Timestamp:       block.Timestamp,
		}
		blocktxs = append(blocktxs, &blocktx)
		return
	}
	//失败的提现交易把手续费推送过去
	if txReceipt.Status != "0x1" && s.Watch.IsWatchAddressExist(tx.From) {
		to :=tx.To
		contractAddress:=""
		if s.Watch.IsContractExist(tx.To){
			tokenTo, _, err := rpc.ERC20{}.ParseTransferData(tx.Input)
			if err != nil {
				to=tokenTo
				contractAddress=tx.To
			}
		}

		blocktx := dao.BlockTx{
			CoinName:        conf.Cfg.Sync.CoinName,
			Txid:            tx.Hash,
			BlockHeight:     block.Height,
			ContractAddress: contractAddress,
			FromAddress:     tx.From,
			ToAddress:       to,
			Amount:          decimal.Zero,
			Fee:             decimal.NewFromBigInt(big.NewInt(0).Mul(tx.GasPrice, big.NewInt(txReceipt.GasUsed)), -18),
			BlockHash:       block.Hash,
			Status:          "success",
			Timestamp:       block.Timestamp,
		}
		blocktxs = append(blocktxs, &blocktx)
		return
	}
	//其他失败的交易不处理
	if txReceipt.Status != "0x1" {
		return nil, errors.New("失败的交易")
	}

	if len(txReceipt.Logs) == 0 {
		return nil, errors.New("交易不含监控地址")
	}
	if s.Watch.IsContractExist(tx.To) {
		_, _, erc20err := rpc.ERC20{}.ParseTransferData(tx.Input)
		if erc20err == nil { //erc20 包括erc20销毁币
			for _, txlog := range txReceipt.Logs {
				contractInfo, err := s.Watch.GetContract(txlog.Address)
				if err != nil {
					log.Info(err.Error())
					continue
				}
				if len(txlog.Topics) != 3 || txlog.Topics[0] != "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
					continue
				}
				from := "0x" + txlog.Topics[1][26:66]
				to := "0x" + txlog.Topics[2][26:66]
				data, _ := hex.DecodeString(strings.TrimPrefix(txlog.Data, "0x"))
				amount := decimal.NewFromBigInt(big.NewInt(0).SetBytes(data), 0-int32(contractInfo.Decimal))
				blocktx := &dao.BlockTx{
					BlockHeight:     tx.BlockNumber,
					BlockHash:       tx.BlockHash,
					Txid:            tx.Hash,
					FromAddress:     from,
					CoinName:        conf.Cfg.Sync.CoinName,
					Timestamp:       block.Timestamp,
					Amount:          amount,
					ToAddress:       to,
					Fee:             decimal.NewFromBigInt(big.NewInt(0).Mul(tx.GasPrice, big.NewInt(txReceipt.GasUsed)), -18),
					Status:          "success",
					ContractAddress: txlog.Address,
				}
				blocktxs = append(blocktxs, blocktx)
			}
			return
		}

		if erc20err != nil {
			//return nil, errors.New("不支持的转账交易")
		}
	}
	if conf.Cfg.Sync.EnableInternal {
		internoTx, err := s.InternalTxns(tx, block.Timestamp.Unix())
		if err != nil {
			return nil, err
		}
		if internoTx != nil {
			blocktxs = append(blocktxs, internoTx)
			return blocktxs,nil
		}
	}
	if tx.BlockNumber> conf.Cfg.Sync.InitHeight { //防止以前已经退币的交易
		for _, txlog := range txReceipt.Logs {
			contractInfo, err := s.Watch.GetContract(txlog.Address)
			if err != nil {
				//log.Info(err.Error())
				continue
			}
			if len(txlog.Topics) == 0 || txlog.Topics[0] != "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
				continue
			}
			//log.Info("a")
			from := "0x" + txlog.Topics[1][26:66]
			to := "0x" + txlog.Topics[2][26:66]
			data, _ := hex.DecodeString(strings.TrimPrefix(txlog.Data, "0x"))
			value := decimal.NewFromBigInt(big.NewInt(0).SetBytes(data), 0-int32(contractInfo.Decimal))
			if !s.Watch.IsWatchAddressExist(to){
				continue
			}

			blocktx := &dao.BlockTx{
				BlockHeight:     tx.BlockNumber,
				BlockHash:       tx.BlockHash,
				Txid:            tx.Hash,
				FromAddress:     from,
				CoinName:        conf.Cfg.Sync.CoinName,
				Timestamp:       block.Timestamp,
				Amount:          value,
				ToAddress:       to,
				Fee:             decimal.NewFromBigInt(big.NewInt(0).Mul(tx.GasPrice, big.NewInt(txReceipt.GasUsed)), -18),
				Status:          "success",
				ContractAddress: txlog.Address,
			}
			blocktxs = append(blocktxs, blocktx)
			err = nil
		}
	}
	return
}

func (s *Scanner) InternalTxns(tx *eth.Transaction, blockTimestamp int64) (*dao.BlockTx, error) {
	debugTraceTransactionInfo, debugTraceTransactionInfoErr := s.GetTraceTransaction(tx.Hash)
	if debugTraceTransactionInfoErr != nil {
		return nil, fmt.Errorf("%s未获取到内部交易,err:%s", tx.Hash, debugTraceTransactionInfoErr.Error())
	}
	if debugTraceTransactionInfo.Calls != nil {
		return s.InternalTxnsRecursion(debugTraceTransactionInfo.Calls, tx, blockTimestamp)
	}
	return nil, fmt.Errorf("%s未获取到内部Calls交易数据,calls为0", tx.Hash)
}

func (s *Scanner) InternalTxnsRecursion(callsArr []eth.TraceTransactionInfoCalls, tx *eth.Transaction, blockTimestamp int64) (*dao.BlockTx, error) {
	for _, v := range callsArr {
		Amount, _ := utils.ParseBigInt(v.Value)
		if Amount.Sign() == 1 && v.Type == "CALL" && v.Input == "0x" && v.Output == "0x" {
			watchLists := make(map[string]bool)
			if s.Watch.IsWatchAddressExist(v.From) {
				watchLists[v.From] = true
			}
			if s.Watch.IsWatchAddressExist(v.To) {
				watchLists[v.To] = true
			}
			if len(watchLists) > 0 {
				//GasUsed, _ := strconv.ParseInt(v.GasUsed, 0, 64)
				blocktx := &dao.BlockTx{
					BlockHeight:     tx.BlockNumber,
					BlockHash:       tx.BlockHash,
					Txid:            tx.Hash,
					CoinName:        conf.Cfg.Sync.CoinName,
					Timestamp:       time.Unix(blockTimestamp, 0),
					ContractAddress: "", //这个没填
				}
				blocktx.FromAddress = v.From
				blocktx.ToAddress = v.To

				blocktx.Amount = decimal.NewFromBigInt(Amount, -18)
				dingding.NotifyError(fmt.Sprintf("内部交易监控到txId: %s\nform：%s\namount:%s\nto:%s", tx.Hash, blocktx.FromAddress, blocktx.Amount.Shift(eth.WEI).String(), blocktx.ToAddress))
				log.Infof("内部交易监控到地址%s form：%s,amount:%s,to:%s", tx.Hash, blocktx.FromAddress, blocktx.Amount.Shift(-eth.WEI).String(), blocktx.ToAddress)
				return blocktx, nil
			}
		}
		Internalblocktx, InternalErr := s.InternalTxnsRecursion(v.Calls, tx, blockTimestamp)
		if InternalErr == nil {
			return Internalblocktx, InternalErr
		}
	}
	return nil, fmt.Errorf("%s未获取到内部交易", tx.Hash)
}

//
//func (s *Scanner) IsContractTx(tx *rpc.Transaction) bool {
//	from := tx.From
//	to := tx.To
//	input := tx.Input
//
//	hoo_from := s.Watch.IsWatchAddressExist(from)
//	hoo_contract := s.Watch.IsContractExist(to)
//	hoo_to := s.Watch.IsWatchAddressExist(to)
//
//	if hoo_from && len(input) < 10 {
//		return false
//	}
//
//	if hoo_to {
//		return false
//	}
//
//	if hoo_contract {
//		return true
//	}
//	return false
//}
//
//// 解析交易
//func parseBlockTxInternal(Watch *services.WatchControl, rpc *rpc.RpcClient, tx *rpc.Transaction) ([]*dao.BlockTx, error) {
//
//	if tx == nil {
//		return nil, fmt.Errorf("tx is null")
//	}
//	var blocktxs = make([]*dao.BlockTx, 0)
//TransactionReceipt:
//	receipt, err := rpc.TransactionReceipt(tx.Hash)
//	if err != nil {
//		log.Error(err.Error())
//		time.Sleep(time.Second * 10)
//		goto TransactionReceipt
//	}
//
//	if receipt.Status != 1 {
//		return nil, errors.New("失败的交易")
//	}
//	log.Info(String(tx))
//	if tx.Value.ToInt().Cmp(big.NewInt(0)) != 0 {
//		log.Info(tx.To)
//		if Watch.IsWatchAddressExist(tx.From) || Watch.IsWatchAddressExist(tx.To) {
//			blocktx := dao.BlockTx{
//				CoinName:    conf.Cfg.Sync.CoinName,
//				Txid:        tx.Hash,
//				BlockHeight: int64(tx.BlockNumber),
//				BlockHash:   tx.BlockHash,
//				FromAddress: tx.From,
//				ToAddress:   tx.To,
//				Amount:      decimal.NewFromBigInt(tx.Value.ToInt(), -18),
//				Fee:         decimal.NewFromBigInt(big.NewInt(0).Mul(receipt.GasUsed.ToInt(), tx.GasPrice.ToInt()), -18),
//				Status:      "success",
//				Timestamp:   time.Now(),
//			}
//			blocktxs = append(blocktxs, &blocktx)
//		}
//	}
//	log.Info(String(receipt.Logs))
//	for _, tlog := range receipt.Logs {
//		if !Watch.IsContractExist(tlog.Address) {
//			log.Info(tlog.Address)
//			continue
//		}
//		if len(tlog.Topics) != 3 || tlog.Topics[0] != "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
//			continue
//		}
//		from := "0x" + receipt.Logs[0].Topics[1][26:66]
//		to := "0x" + receipt.Logs[0].Topics[2][26:66]
//		//log.Info(receipt.Logs[0].Topics[1], from, to)
//		//log.Info(from, to, tx.To)
//		if !Watch.IsWatchAddressExist(from) && !Watch.IsWatchAddressExist(to) {
//			log.Info(from, to)
//			continue
//		}
//		contract, _ := Watch.GetContract(tx.To)
//		d, err := hex.DecodeString(strings.TrimPrefix(tlog.Data, "0x"))
//		if err != nil {
//			panic(tx.Hash + err.Error())
//		}
//		amount := decimal.NewFromBigInt(big.NewInt(0).SetBytes(d), 0-int32(contract.Decimal))
//		blocktx := dao.BlockTx{
//			CoinName:        contract.Name,
//			Txid:            tx.Hash,
//			BlockHeight:     int64(tx.BlockNumber),
//			ContractAddress: tlog.Address,
//			FromAddress:     from,
//			ToAddress:       to,
//			Amount:          amount,
//			Fee:             decimal.NewFromBigInt(big.NewInt(0).Mul(tx.GasPrice.ToInt(), receipt.GasUsed.ToInt()), -18),
//			BlockHash:       tx.BlockHash,
//			Status:          "success",
//			Timestamp:       time.Now(),
//		}
//		blocktxs = append(blocktxs, &blocktx)
//	}
//	return blocktxs, nil
//}
