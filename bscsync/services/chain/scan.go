package chain

import (
	"dataserver/common"
	"dataserver/conf"
	"dataserver/log"
	"dataserver/models/po"
	"dataserver/services"
	"dataserver/utils"
	"dataserver/utils/dingding"
	"dataserver/utils/eth"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	GDR_Contract = "0xc828dea19b9d68214a140620089853d4a70413bd"
	GDR_Owner    = "0xaa2c88b2ed76408487dd42b0bb0df5cba181affb"
	ZERO         = "0x0000000000000000000000000000000000000000000000000000000000000000"
)

type Scanner struct {
	*eth.RpcClient
	lock    *sync.Mutex
	conf    conf.SyncConfig
	watcher *services.WatchControl
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watcher *services.WatchControl) common.Scanner {
	// 如果启动eth，顺便启动定制加载的合约
	err := InitEthClient(node.Url)
	if err != nil {
		panic(err)
	}
	return &Scanner{
		RpcClient: eth.NewRpcClient(node.Url),
		lock:      &sync.Mutex{},
		conf:      conf.Sync,
		watcher:   watcher,
	}
}

func (s *Scanner) Rollback(height int64) {
	// 删除指定高度之后的数据
	po.DeleteBlockInfo(height)

	po.DeleteBlockTX(height)
}

// 爬数据
func (s *Scanner) Init() error {
	return nil
}

func (s *Scanner) Clear() {
}

func (s *Scanner) GetBestBlockHeight() (int64, error) {
	return s.BlockNumber()
}

func (s *Scanner) GetCurrentBlockHeight() (int64, error) {
	return po.GetMaxBlockIndex()
}

// 批量扫描多个区块
func (s *Scanner) BatchScanIrreverseBlocks(startHeight, endHeight, bestHeight int64) *sync.Map {
	starttime := time.Now()
	count := endHeight - startHeight
	taskmap := &sync.Map{}
	wg := &sync.WaitGroup{}

	wg.Add(int(count))
	for i := int64(0); i < count; i++ {
		height := startHeight + i
		go func(w *sync.WaitGroup) {
			if task, err := s.ScanIrreverseBlock(height, bestHeight); err == nil {
				taskmap.Store(height, task)
			}
			w.Done()
		}(wg)
	}
	wg.Wait()
	log.Debugf("***batchScanBlocks used time : %f 's", time.Since(starttime).Seconds())
	return taskmap
}

func (s *Scanner) ScanReverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock(height, bestHeight)
}

// 扫描一个区块
func (s *Scanner) ScanIrreverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock(height, bestHeight)
}

func (s *Scanner) scanBlock(height, bestHeight int64) (common.ProcTask, error) {
	starttime := time.Now()

	block, err := s.GetBlockByNumber(height, true)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}

	cnt, err := po.GetBlockCountByHash(block.Hash)
	if err != nil {
		return nil, fmt.Errorf("database err")
	}

	if cnt > 0 {
		return nil, fmt.Errorf("already have block , count : %d", cnt)
	}

	task := &EthProcTask{
		bestHeight: bestHeight,
		block: &po.BlockInfo{
			Height:         block.Number,
			Hash:           block.Hash,
			FrontBlockHash: block.ParentHash,
			Timestamp:      time.Unix(block.Timestamp, 0),
			Transactions:   len(block.Transactions),
			Confirmations:  bestHeight - height + 1,
			CreateTime:     time.Now(),
		},
		txInfos: make([][]*po.BlockTX, 0),
	}

	if task.block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}
	// 处理区块内的交易
	if len(block.Transactions) > 0 {
		if s.conf.EnableGoroutine {
			wg := &sync.WaitGroup{}
			wg.Add(len(block.Transactions))
			for _, tx := range block.Transactions {
				go s.batchParseTx(&tx, block.Timestamp, task, wg)
			}
			wg.Wait()
			log.Infof("scanBlock %d ,used time : %f 's", height, time.Since(starttime).Seconds())
		} else {
			ignoreCount := 0
			for _, tx := range block.Transactions {
				blockTx, isIgnore, err := s.parseBlockTX(&tx, block.Timestamp)
				if isIgnore {
					// 表示这一条交易没有我们关心的地址和合约，跳过
					ignoreCount++
				}
				if err != nil {
					log.Errorf("parseBlockTX 出错:%s", err.Error())
					continue
				}
				if len(blockTx) > 0 {
					task.txInfos = append(task.txInfos, blockTx)
				}
			}
			log.Infof("scanBlock %d 共%d条交易，忽略了%d条, 处理了%d条 共耗时: %f 's", height, len(block.Transactions), ignoreCount, len(block.Transactions)-ignoreCount, time.Since(starttime).Seconds())
		}
	} else {
		log.Infof("scanBlock %d 没有任何交易", height)
	}
	return task, nil
}

// 批量解析交易
func (s *Scanner) batchParseTx(tx *eth.Transaction, blockTimestamp int64, task *EthProcTask, w *sync.WaitGroup) {
	defer w.Done()
	blockTx, _, err := s.parseBlockTX(tx, blockTimestamp)
	if err == nil && len(blockTx) > 0 {
		s.lock.Lock()
		task.txInfos = append(task.txInfos, blockTx)
		s.lock.Unlock()
	}
}

// 解析交易
func (s *Scanner) parseBlockTX(tx *eth.Transaction, blockTimestamp int64) ([]*po.BlockTX, bool, error) {
	res := make([]*po.BlockTX, 0)

	if tx == nil {
		return res, true, fmt.Errorf("tx is null")
	}
	txReceipt, err := s.GetTransactionReceipt(tx.Hash)
	if err != nil {
		return res, true, err
	}
	if s.IsMainCoinTx(tx) {
		// 检查是否我们关心的出账和入账地址
		// 由于这是主链币交易，所以出账和入账地址直接从`tx.from`和`tx.to`获取就可以了
		if !s.watcher.IsWatchAddressExist(tx.From) && !s.watcher.IsWatchAddressExist(tx.To) {
			// 都不是关心的地址，不需要再拉取transactionReceipt信息，直接返回
			return res, true, nil
		}
		parseTxStartTime := time.Now()
		errParseTx := s.parseTx(tx, blockTimestamp, &res, txReceipt)
		log.Debugf("hash %s parseTx used time : %f 's", tx.Hash, time.Since(parseTxStartTime).Seconds())

		if errParseTx != nil {
			return res, false, errParseTx
		}
	} else {
		//isCare, err := s.isCareContractTx(tx)
		//if err != nil || !isCare {
		//	return res, true, err
		//}

		parseTxReceiptStartTime := time.Now()
		if errParseTxReceipt := s.parseTxReceipt(tx, blockTimestamp, &res, txReceipt); errParseTxReceipt != nil {
			blocktx, InternalTxnsErr := s.InternalTxns(tx, blockTimestamp, txReceipt)
			if InternalTxnsErr == nil {
				res = append(res, blocktx)
				return res, false, nil //这个false是干嘛的
			}
			return res, false, errParseTxReceipt
		}
		log.Debugf("hash %s parseTxReceipt used time : %f 's", tx.Hash, time.Since(parseTxReceiptStartTime).Seconds())
	}
	if len(res) == 0 {
		blocktx, InternalTxnsErr := s.InternalTxns(tx, blockTimestamp, txReceipt)
		if InternalTxnsErr == nil {
			res = append(res, blocktx)
		}
	}
	return res, false, nil
}

func (s *Scanner) parseTx(tx *eth.Transaction, blockTimestamp int64, res *[]*po.BlockTX, txReceipt *eth.TransactionReceipt) error {
	// 非合约交易
	// 主币transfer的处理逻辑
	if txReceipt.Status != "0x1" {
		log.Infof("[无效交易报警]parseTx txId=%s 报警", txReceipt.TransactionHash)
		dingding.NotifyErrorForTx(txReceipt.TransactionHash, tx.From, tx.To)
		log.Infof("%s  ", txReceipt.TransactionHash)
		return fmt.Errorf("%s 无效交易", txReceipt.TransactionHash)
	}

	blockTx := &po.BlockTX{
		BlockHeight: tx.BlockNumber,
		BlockHash:   tx.BlockHash,
		Txid:        tx.Hash,
		FromAddress: tx.From,
		Nonce:       tx.Nonce,
		GasUsed:     txReceipt.GasUsed,
		GasPrice:    tx.GasPrice.Int64(),
		Input:       tx.Input,
		CoinName:    s.conf.Name,
		Decimal:     eth.WEI,
		Status:      1, // 主币交易，只有成功才能拿到
		Timestamp:   time.Unix(blockTimestamp, 0),
		Amount:      decimal.NewFromBigInt(tx.Value, 0),
		ToAddress:   tx.To,
		CreateTime:  time.Now(),
	}
	*res = append(*res, blockTx)
	return nil
}

func (s *Scanner) parseTxReceipt(tx *eth.Transaction, blockTimestamp int64, res *[]*po.BlockTX, txReceipt *eth.TransactionReceipt) error {
	//txReceipt, err := s.GetTransactionReceipt(tx.Hash)
	//if err != nil {
	//	return err
	//}
	if txReceipt.Status != "0x1" {
		go func(txId, from, to string) {
			// 检查交易的合约地址是否我们关心的地址
			//if _, err = s.watcher.GetContract(tx.To); err != nil {
			//	log.Infof("[无效交易报警]parseTxReceipt txId=%s 不关心合约地址 %s", txReceipt.TransactionHash, tx.To)
			//	return
			//}
			if !s.watcher.IsWatchAddressExist(from) {
				log.Infof("[无效交易报警]parseTxReceipt txId=%s 不关心from地址 %s", txId, from)
				return
			}
			log.Infof("[无效交易报警]parseTxReceipt txId=%s from=%s to=%报警", txId, from, to)
			dingding.NotifyErrorForTx(txId, from, to)
		}(txReceipt.TransactionHash, tx.From, tx.To)

		log.Infof("%s 无效交易", txReceipt.TransactionHash)
		return fmt.Errorf("%s 无效交易", txReceipt.TransactionHash)
	}

	// if txReceipt.Removed {
	// 	return nil, fmt.Errorf("%s 交易已经被删除",txReceipt.TransactionHash)
	// }

	inAmount := decimal.Zero
	for i, lg := range txReceipt.Logs {
		blockTx := &po.BlockTX{
			BlockHeight:     tx.BlockNumber,
			BlockHash:       tx.BlockHash,
			Txid:            tx.Hash,
			Nonce:           tx.Nonce,
			GasUsed:         txReceipt.GasUsed,
			GasPrice:        tx.GasPrice.Int64(),
			Input:           tx.Input,
			CoinName:        s.conf.Name,
			Decimal:         eth.WEI,
			Timestamp:       time.Unix(blockTimestamp, 0),
			ContractAddress: lg.Address,
			CreateTime:      time.Now(),
		}
		sta, staErr := utils.ParseInt(txReceipt.Status)
		if staErr != nil {
			log.Errorf("tx Log[%d] status parse err:%s", i, staErr.Error())
		}
		blockTx.Status = sta

		// 保存Logs对应的索引数据
		btys, jmErr := json.Marshal(txReceipt.Logs[i])
		if jmErr != nil {
			log.Errorf("tx Log[%d] json marshal err:%s", i, jmErr.Error())
		} else {
			blockTx.Logs = string(btys)
		}

		// 没有输出日志数据，认为是非合法的交易
		// status=2 表示失败
		if txReceipt.Logs == nil || len(txReceipt.Logs) == 0 {
			blockTx.Status = 2
		}

		if lg.Data == "" || len(lg.Data) < 3 {
			continue
		}
		tmp, success := new(big.Int).SetString(lg.Data[4:], 16)
		if !success {
			log.Errorf("金额转换出错 %s", lg.Data[4:])
			continue
		}
		blockTx.Amount = decimal.NewFromBigInt(tmp, 0)

		// 内部金额叠加
		inAmount = inAmount.Add(decimal.NewFromBigInt(tmp, 0))

		if len(lg.Topics) < 3 || len(lg.Topics[0]) < 66 || len(lg.Topics[1]) < 66 || len(lg.Topics[2]) < 66 {
			continue
		}
		if lg.Topics[0] == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
			blockTx.FromAddress = "0x" + lg.Topics[1][26:66]
			blockTx.ToAddress = "0x" + lg.Topics[2][26:66]

			if lg.Address == GDR_Contract && blockTx.ToAddress == GDR_Owner && lg.Data == ZERO {
				log.Infof("%s 是GDR币种的特殊交易，不需要推送", tx.Hash)
				continue
			}

		} else {
			continue
		}
		*res = append(*res, blockTx)
	}
	if len(*res) > 0 {
		// log.Println(fmt.Sprintf("【%s】，存在交易，开始解析数据是否销毁", tx.Hash, tx.From, tx.To))
		// 由于内层销毁的金额无法获取，因此需要外层-内层的金额
		outAmount := decimal.Zero
		rawTx := tx
		if len(rawTx.Input) == 0 || len(txReceipt.Logs) == 0 {
			// log.Printf("%s 没有可以解析的交易", txReceipt.TransactionHash)
			return fmt.Errorf("%s 没有可以解析的交易", txReceipt.TransactionHash)
		}
		// 差异金额的标识
		if len(rawTx.Input) == 138 && len(txReceipt.Logs) == 1 && !inAmount.IsZero() {
			if strings.HasPrefix(rawTx.Input, "0xa9059cbb000000000000000000000000") {
				// 有可能是销毁币种
				am, _ := new(big.Int).SetString(rawTx.Input[74:], 16)
				outAmount = decimal.NewFromBigInt(am, 0)
				if outAmount.IsZero() {
					return fmt.Errorf("txid:[%s]解析原始金额错误,可能非交易类型", tx.Hash)
				}
				if outAmount.GreaterThan(inAmount) {
					detroyTx := &po.BlockTX{
						BlockHeight:     tx.BlockNumber,
						BlockHash:       tx.BlockHash,
						Txid:            tx.Hash,
						Nonce:           tx.Nonce,
						GasUsed:         txReceipt.GasUsed,
						GasPrice:        tx.GasPrice.Int64(),
						Input:           tx.Input,
						CoinName:        s.conf.Name,
						Decimal:         eth.WEI, //
						Timestamp:       time.Unix(blockTimestamp, 0),
						ContractAddress: rawTx.To,
						CreateTime:      time.Now(),
						Status:          1,
						Amount:          outAmount.Sub(inAmount),
						FromAddress:     rawTx.From,
						ToAddress:       "0x0000000000000000000000000000000000000000",
					}

					if detroyTx.ContractAddress == "0x18ff245c134d9daa6fed977617654490ba4da526" {
						log.Infof("【%s】销毁币种 MASKDOGE，跳过推送 ", detroyTx.Txid)
					} else {
						*res = append(*res, detroyTx)
						dd, _ := json.Marshal(detroyTx)
						log.Infof("【%s】销毁币种，添加数据 【%s】", detroyTx.Txid, string(dd))
					}
				}
			}
		}
	}

	return nil
}
func (s *Scanner) InternalTxns(tx *eth.Transaction, blockTimestamp int64, txReceipt *eth.TransactionReceipt) (*po.BlockTX, error) {
	if txReceipt.Status != "0x1" {
		log.Infof("%s 无效内部交易", txReceipt.TransactionHash)
		return nil, fmt.Errorf("%s 无效内部交易", txReceipt.TransactionHash)
	}
	debugTraceTransactionInfo, debugTraceTransactionInfoErr := s.GetTraceTransaction(tx.Hash)
	if debugTraceTransactionInfoErr != nil {
		return nil, fmt.Errorf("%s未获取到内部交易,err:%s", tx.Hash, debugTraceTransactionInfoErr.Error())
	}
	if debugTraceTransactionInfo.Calls != nil {
		return s.InternalTxnsRecursion(debugTraceTransactionInfo.Calls, tx, blockTimestamp)
	}
	return nil, fmt.Errorf("%s未获取到内部Calls交易数据,calls为0", tx.Hash)
}
func (s *Scanner) InternalTxnsRecursion(callsArr []eth.TraceTransactionInfoCalls, tx *eth.Transaction, blockTimestamp int64) (*po.BlockTX, error) {
	for _, v := range callsArr {
		Amount, _ := utils.ParseBigInt(v.Value)
		if Amount.Sign() == 1 && v.Type == "CALL" && v.Input == "0x" && v.Output == "0x" {
			watchLists := make(map[string]bool)
			if s.watcher.IsWatchAddressExist(v.From) {
				watchLists[v.From] = true
			}
			if s.watcher.IsWatchAddressExist(v.To) {
				watchLists[v.To] = true
			}
			if len(watchLists) > 0 {
				GasUsed, _ := strconv.ParseInt(v.GasUsed, 0, 64)
				blocktx := &po.BlockTX{
					BlockHeight:     tx.BlockNumber,
					BlockHash:       tx.BlockHash,
					Txid:            tx.Hash,
					Nonce:           tx.Nonce,
					GasUsed:         GasUsed,
					GasPrice:        tx.GasPrice.Int64(),
					Input:           tx.Input,
					CoinName:        s.conf.Name,
					Decimal:         eth.WEI,
					Timestamp:       time.Unix(blockTimestamp, 0),
					ContractAddress: "", //这个没填
					Status:          1,
				}
				blocktx.FromAddress = v.From
				blocktx.ToAddress = v.To
				blocktx.Amount = decimal.NewFromBigInt(Amount, 0)
				dingding.NotifyError(fmt.Sprintf("BSC监测到内部交易：%s\nform：%s\namount:%s\nto:%s", tx.Hash, blocktx.FromAddress, blocktx.Amount.Shift(-eth.WEI).String(), blocktx.ToAddress))
				log.Infof("BSC监测到内部交易%s form：%s,amount:%s,to:%s", tx.Hash, blocktx.FromAddress, blocktx.Amount.Shift(-eth.WEI).String(), blocktx.ToAddress)
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
func (s *Scanner) isCareContractTx(tx *eth.Transaction) (isCare bool, err error) {
	// 检查交易的出账地址是否我们关心的地址
	if s.watcher.IsWatchAddressExist(tx.From) {
		return true, nil
	}
	// 检查交易的合约地址是否我们关心的地址
	if _, err := s.watcher.GetContract(tx.To); err == nil {
		return true, nil
	}

	// 如果无法从单纯的`from`和`to`检查出来
	// 那么就从`input`数据分析

	if len(tx.Input) <= 10 {
		return false, fmt.Errorf("txId:%s 存在异常：非主链币交易，但是input数据异常 input: %s", tx.Hash, tx.Input)
	}

	data := tx.Input[10:] // 去除函数签名值
	count := len(data) / 64
	for i := 0; i < count; i++ {
		startIndex := i * 64
		paramPrefixZero := data[startIndex : startIndex+64]
		paramHexFront := fmt.Sprintf("0x%s", paramPrefixZero[24:]) // 0000000000000000000000000084ce0d0c84999703b64fdda56d1e213ab3c6cc
		paramHexBack := fmt.Sprintf("0x%s", paramPrefixZero[:40])  // 0084ce0d0c84999703b64fdda56d1e213ab3c6cc000000000000000000000000

		// 是否我们关心的出入账地址
		if s.watcher.IsWatchAddressExist(paramHexFront) || s.watcher.IsWatchAddressExist(paramHexBack) {
			return true, nil
		}

		// 是否我们关心的合约地址
		if _, err := s.watcher.GetContract(paramHexFront); err == nil {
			return true, nil
		}
		if _, err := s.watcher.GetContract(paramHexBack); err == nil {
			return true, nil
		}
	}
	return false, nil
}
