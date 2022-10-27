package cds

import (
	"fmt"
	"github.com/shopspring/decimal"
	"log"
	"math/big"
	"rsksync/common"
	"rsksync/conf"
	dao "rsksync/models/po/cds"
	"rsksync/utils/cds"
	"sync"
	"time"
)

type Scanner struct {
	*cds.RpcClient
	lock *sync.Mutex
	conf conf.SyncConfig
}

func NewScanner(conf conf.Config, node conf.NodeConfig) common.Scanner {
	return &Scanner{
		RpcClient: cds.NewRpcClient(node.Url),
		lock:      &sync.Mutex{},
		conf:      conf.Sync,
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
	log.Printf("***batchScanBlocks used time : %f 's", time.Since(starttime).Seconds())
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

	task := &CdsProcTask{
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
	log.Printf("scanBlock %d ,used time : %f 's", height, time.Since(starttime).Seconds())
	return task, nil
}

//批量解析交易
func (s *Scanner) batchParseTx(tx *cds.Transaction, bestHeight, blockTimestamp int64, task *CdsProcTask, w *sync.WaitGroup) {
	defer w.Done()
	blockTx, err := s.parseBlockTX(tx, bestHeight, blockTimestamp)
	if err == nil {
		s.lock.Lock()
		task.txInfos = append(task.txInfos, blockTx)
		s.lock.Unlock()
	}
}

// 解析交易
func (s *Scanner) parseBlockTX(tx *cds.Transaction, bestHeight, blockTimestamp int64) ([]*dao.BlockTX, error) {

	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}
	res := make([]*dao.BlockTX, 0)

	if !s.IsContractTx(tx) {
		log.Println("非合约数据解析")
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
			Decimal:     cds.WEI,
			Timestamp:   time.Unix(blockTimestamp, 0),
			Amount:      decimal.NewFromBigInt(tx.Value, 0),
			ToAddress:   tx.To,
		}
		res = append(res, blocktx)
	} else {
		toAddr, amt, err := cds.ERC20{}.ParseTransferData(tx.Input)
		if err == nil {
			log.Println("erc20数据解析")
			//erc2o tx
			blocktx := &dao.BlockTX{
				BlockHeight:     tx.BlockNumber,
				BlockHash:       tx.BlockHash,
				Txid:            tx.Hash,
				FromAddress:     tx.From,
				Nonce:           tx.Nonce,
				GasUsed:         tx.Gas,
				GasPrice:        tx.GasPrice.Int64(),
				Input:           tx.Input,
				CoinName:        s.conf.Name,
				Decimal:         cds.WEI,
				Timestamp:       time.Unix(blockTimestamp, 0),
				Amount:          decimal.NewFromBigInt(amt, 0),
				ToAddress:       toAddr,
				ContractAddress: tx.To,
			}
			res = append(res, blocktx)
		} else {
			log.Println("合约数据解析")
			//contract tx
			txReceipt, err := s.GetTransactionReceipt(tx.Hash)
			if err != nil {
				return nil, err
			}
			for _, lg := range txReceipt.Logs {
				blocktx := &dao.BlockTX{
					BlockHeight:     tx.BlockNumber,
					BlockHash:       tx.BlockHash,
					Txid:            tx.Hash,
					Nonce:           tx.Nonce,
					GasUsed:         txReceipt.GasUsed,
					GasPrice:        tx.GasPrice.Int64(),
					Input:           tx.Input,
					CoinName:        s.conf.Name,
					Decimal:         cds.WEI,
					Timestamp:       time.Unix(blockTimestamp, 0),
					ContractAddress: lg.Address,
				}
				if lg.Data == "" || len(lg.Data) < 3 {
					continue
				}
				tmp, _ := new(big.Int).SetString(lg.Data[4:], 16)
				blocktx.Amount = decimal.NewFromBigInt(tmp, 0)
				if len(lg.Topics) < 3 || len(lg.Topics[0]) < 66 || len(lg.Topics[1]) < 66 || len(lg.Topics[2]) < 66 {
					continue
				}
				if lg.Topics[0] == "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" {
					blocktx.FromAddress = "0x" + lg.Topics[1][26:66]
					blocktx.ToAddress = "0x" + lg.Topics[2][26:66]
				} else {
					continue
				}
				res = append(res, blocktx)
			}
		}
	}
	return res, nil
}
