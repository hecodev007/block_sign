package bsc

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"log"
	"math/big"
	"rsksync/common"
	"rsksync/conf"
	dao "rsksync/models/po/bsc"
	"rsksync/services"
	"rsksync/utils"
	"rsksync/utils/bsc"
	"rsksync/utils/eth"
	"strings"
	"sync"
	"time"
)

type Scanner struct {
	*bsc.RpcClient
	lock  *sync.Mutex
	conf  conf.SyncConfig
	watch *services.WatchControl
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	//如果启动eth，顺便启动定制加载的合约
	return &Scanner{
		RpcClient: bsc.NewRpcClient(node.Url),
		lock:      &sync.Mutex{},
		conf:      conf.Sync,
		watch:     watch,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	dao.DeleteBlockInfo(height)

	dao.DeleteBlockTX(height)
}

//爬数据
func (s *Scanner) Init() error {
	return nil
}

func (s *Scanner) Clear() {
}

func (s *Scanner) GetBestBlockHeight() (int64, error) {
	return s.BlockNumber()
}

func (s *Scanner) GetCurrentBlockHeight() (int64, error) {
	return dao.GetMaxBlockIndex()
}

//批量扫描多个区块
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
	log.Printf("***batchScanBlocks used time : %f 's \n", time.Since(starttime).Seconds())
	return taskmap
}

func (s *Scanner) ScanReverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock(height, bestHeight)
}

//扫描一个区块
func (s *Scanner) ScanIrreverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock(height, bestHeight)
}

func (s *Scanner) scanBlock(height, bestHeight int64) (common.ProcTask, error) {
	starttime := time.Now()

	block, err := s.GetBlockByNumber(height, true)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}

	cnt, err := dao.GetBlockCountByHash(block.Hash)
	if err != nil {
		return nil, fmt.Errorf("database err")
	}

	if cnt > 0 {
		return nil, fmt.Errorf("already have block , count : %d", cnt)
	}

	task := &BscProcTask{
		bestHeight: bestHeight,
		block: &dao.BlockInfo{
			Height:         block.Number,
			Hash:           block.Hash,
			FrontBlockHash: block.ParentHash,
			Timestamp:      time.Unix(block.Timestamp, 0),
			Transactions:   len(block.Transactions),
			Confirmations:  bestHeight - height + 1,
			CreateTime:     time.Now(),
		},
		txInfos: make([][]*dao.BlockTX, 0),
	}

	if task.block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}
	//处理区块内的交易
	if len(block.Transactions) > 0 {
		if s.conf.EnableGoroutine {
			wg := &sync.WaitGroup{}
			wg.Add(len(block.Transactions))
			for _, tmp := range block.Transactions {

				tx := tmp
				go s.batchParseTx(&tx, bestHeight, block.Timestamp, task, wg)
			}
			wg.Wait()
		} else {
			for _, tx := range block.Transactions {
				blockTx, err := s.parseBlockTX(&tx, bestHeight, block.Timestamp)
				if err == nil {
					task.txInfos = append(task.txInfos, blockTx)
				}
			}
		}
	}
	log.Printf("scanBlock %d ,used time : %f 's \n", height, time.Since(starttime).Seconds())
	return task, nil
}

//批量解析交易
func (s *Scanner) batchParseTx(tx *bsc.Transaction, bestHeight, blockTimestamp int64, task *BscProcTask, w *sync.WaitGroup) {
	defer w.Done()
	blockTx, err := s.parseBlockTX(tx, bestHeight, blockTimestamp)
	if err == nil {
		s.lock.Lock()
		task.txInfos = append(task.txInfos, blockTx)
		s.lock.Unlock()
	}
}

// 解析交易
func (s *Scanner) parseBlockTX(tx *bsc.Transaction, bestHeight, blockTimestamp int64) ([]*dao.BlockTX, error) {

	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}
	res := make([]*dao.BlockTX, 0)
	if !s.IsContractTx(tx) {
		//非合约交易
		//主币transfer的处理逻辑
		txReceipt, err := s.GetTransactionReceipt(tx.Hash)
		if err != nil {
			return nil, err
		}
		if txReceipt.Status != "0x1" {
			//log.Printf("%s 无效交易", txReceipt.TransactionHash)
			return nil, fmt.Errorf("%s 无效交易", txReceipt.TransactionHash)
		}

		blocktx := &dao.BlockTX{
			BlockHeight: tx.BlockNumber,
			BlockHash:   tx.BlockHash,
			Txid:        tx.Hash,
			FromAddress: tx.From,
			Nonce:       tx.Nonce,
			GasUsed:     tx.Gas,
			GasPrice:    tx.GasPrice.Int64(),
			Input:       tx.Input,
			CoinName:    s.conf.Name,
			Decimal:     bsc.WEI,
			Timestamp:   time.Unix(blockTimestamp, 0),
			Amount:      decimal.NewFromBigInt(tx.Value, 0),
			ToAddress:   tx.To,
		}
		res = append(res, blocktx)
	} else {
		//toAddr, amt, err := eth.ERC20{}.ParseTransferData(tx.Input)
		if errPTR := s.parseTxReceipt(tx, blockTimestamp, &res); errPTR != nil {
			return nil, errPTR
		}

		//toAddr, amt, err := bsc.ERC20{}.ParseTransferData(tx.Input)
		//
		//if err == nil {
		//	//erc2o tx
		//	blocktx := &dao.BlockTX{
		//		BlockHeight:     tx.BlockNumber,
		//		BlockHash:       tx.BlockHash,
		//		Txid:            tx.Hash,
		//		FromAddress:     tx.From,
		//		Nonce:           tx.Nonce,
		//		GasUsed:         tx.Gas,
		//		GasPrice:        tx.GasPrice.Int64(),
		//		Input:           tx.Input,
		//		CoinName:        s.conf.Name,
		//		Decimal:         bsc.WEI,
		//		Timestamp:       time.Unix(blockTimestamp, 0),
		//		Amount:          decimal.NewFromBigInt(amt, 0),
		//		ToAddress:       toAddr,
		//		ContractAddress: tx.To,
		//	}
		//	res = append(res, blocktx)
		//} else {
		//	//contract tx
		//	//log.Infof("sta ,asko dego:%s",tx.Hash)
		//	txReceipt, err := s.GetTransactionReceipt(tx.Hash)
		//	if err != nil {
		//		return nil, err
		//	}
		//	if txReceipt.Status != "0x1" {
		//		log.Infof("%s 无效交易", txReceipt.TransactionHash)
		//		return nil, fmt.Errorf("%s 无效交易", txReceipt.TransactionHash)
		//	}
		//	//if txReceipt.Removed {
		//	//	return nil, fmt.Errorf("%s 交易已经被删除",txReceipt.TransactionHash)
		//	//}
		//	for _, lg := range txReceipt.Logs {
		//		blocktx := &dao.BlockTX{
		//			BlockHeight:     tx.BlockNumber,
		//			BlockHash:       tx.BlockHash,
		//			Txid:            tx.Hash,
		//			Nonce:           tx.Nonce,
		//			GasUsed:         txReceipt.GasUsed,
		//			GasPrice:        tx.GasPrice.Int64(),
		//			Input:           tx.Input,
		//			CoinName:        s.conf.Name,
		//			Decimal:         bsc.WEI,
		//			Timestamp:       time.Unix(blockTimestamp, 0),
		//			ContractAddress: lg.Address,
		//		}
		//		if lg.Data == "" || len(lg.Data) < 3 {
		//			continue
		//		}
		//		tmp, _ := new(big.Int).SetString(lg.Data[4:], 16)
		//		blocktx.Amount = decimal.NewFromBigInt(tmp, 0)
		//		if len(lg.Topics) < 3 || len(lg.Topics[0]) < 66 || len(lg.Topics[1]) < 66 || len(lg.Topics[2]) < 66 {
		//			continue
		//		}
		//		if lg.Topics[0] == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
		//			blocktx.FromAddress = "0x" + lg.Topics[1][26:66]
		//			blocktx.ToAddress = "0x" + lg.Topics[2][26:66]
		//		} else {
		//			continue
		//		}
		//		res = append(res, blocktx)
		//	}
		//}
	}
	return res, nil
}

func (s *Scanner) parseTxReceipt(tx *bsc.Transaction, blockTimestamp int64, res *[]*dao.BlockTX) error {

	txReceipt, err := s.GetTransactionReceipt(tx.Hash)
	if err != nil {
		return err
	}
	if txReceipt.Status != "0x1" {
		//log.Printf("%s 无效交易", txReceipt.TransactionHash)
		return fmt.Errorf("%s 无效交易", txReceipt.TransactionHash)
	}
	//if txReceipt.Removed {
	//	return nil, fmt.Errorf("%s 交易已经被删除",txReceipt.TransactionHash)
	//}
	inAmount := decimal.Zero
	for i, lg := range txReceipt.Logs {
		blockTx := &dao.BlockTX{
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
			ContractAddress: lg.Address,
			CreateTime:      time.Now(),
		}
		sta, staErr := utils.ParseInt(txReceipt.Status)
		if staErr != nil {
			log.Printf("tx Log[%d] status parse err:%s", i, staErr.Error())
		}
		blockTx.Status = sta

		// 保存Logs对应的索引数据
		btys, jmErr := json.Marshal(txReceipt.Logs[i])
		if jmErr != nil {
			log.Printf("tx Log[%d] json marshal err:%s", i, jmErr.Error())
		} else {
			blockTx.Logs = string(btys)
		}

		//没有输出日志数据，认为是非合法的交易
		//status=2 表示失败
		if txReceipt.Logs == nil || len(txReceipt.Logs) == 0 {
			blockTx.Status = 2
		}

		if lg.Data == "" || len(lg.Data) < 3 {
			continue
		}
		tmp, _ := new(big.Int).SetString(lg.Data[4:], 16)
		blockTx.Amount = decimal.NewFromBigInt(tmp, 0)

		//内部金额叠加
		inAmount = inAmount.Add(decimal.NewFromBigInt(tmp, 0))
		if len(lg.Topics) < 3 || len(lg.Topics[0]) < 66 || len(lg.Topics[1]) < 66 || len(lg.Topics[2]) < 66 {
			continue
		}
		if lg.Topics[0] == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
			blockTx.FromAddress = "0x" + lg.Topics[1][26:66]
			blockTx.ToAddress = "0x" + lg.Topics[2][26:66]

			hasFrom := false
			hasTo := false

			tmpWatchList := make(map[string]bool)
			if s.watch.IsWatchAddressExist(blockTx.FromAddress) {
				tmpWatchList[blockTx.FromAddress] = true
				hasFrom = true
			}
			if s.watch.IsWatchAddressExist(blockTx.ToAddress) {
				tmpWatchList[blockTx.ToAddress] = true
				hasTo = true
			}
			if !hasFrom && !hasTo {
				//if tx.Hash == "0x8053a5242d9ecfcc288a513f34a1010d2939922f0393b75f15ef12c19f39e4b7" {
				//	log.Println(fmt.Sprintf("[%s],没有，from[%s],to[%s]", tx.Hash, tx.From, tx.To))
				//	//没有关注的交易直接忽略
				//	continue
				//}
				continue
			}
		} else {
			continue
		}
		*res = append(*res, blockTx)
	}

	if len(*res) > 0 {
		//log.Println(fmt.Sprintf("【%s】，存在交易，开始解析数据是否销毁", tx.Hash, tx.From, tx.To))
		//由于内层销毁的金额无法获取，因此需要外层-内层的金额
		outAmount := decimal.Zero
		rawTx, err := s.GetTransactionByHash(tx.Hash)
		if err != nil {
			return err
		}
		if len(rawTx.Input) == 0 || len(txReceipt.Logs) == 0 {
			//log.Printf("%s 没有可以解析的交易", txReceipt.TransactionHash)
			return fmt.Errorf("%s 没有可以解析的交易", txReceipt.TransactionHash)
		}
		//差异金额的标识
		if len(rawTx.Input) == 138 && len(txReceipt.Logs) == 1 && !inAmount.IsZero() {
			if strings.HasPrefix(rawTx.Input, "0xa9059cbb000000000000000000000000") {
				//有可能是销毁币种
				am, _ := new(big.Int).SetString(rawTx.Input[74:], 16)
				outAmount = decimal.NewFromBigInt(am, 0)
				if outAmount.IsZero() {
					return fmt.Errorf("txid:[%s]解析原始金额错误,可能非交易类型", tx.Hash)
				}
				if outAmount.GreaterThan(inAmount) {
					detroyTx := &dao.BlockTX{
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
					*res = append(*res, detroyTx)
					dd, _ := json.Marshal(detroyTx)
					log.Printf("【%s】销毁币种，添加数据 【%s】", string(detroyTx.Txid), string(dd))
				}
			}
		}
	}
	return nil
}
