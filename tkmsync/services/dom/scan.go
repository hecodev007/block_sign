package dom

import (
	"fmt"
	"github.com/shopspring/decimal"
	"log"
	"math/big"
	"rsksync/common"
	"rsksync/conf"
	dao "rsksync/models/po/dom"
	"rsksync/services"
	"rsksync/utils/dom"
	"sync"
	"time"
)

type Scanner struct {
	*dom.RpcClient
	lock  *sync.Mutex
	conf  conf.SyncConfig
	watch *services.WatchControl
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	return &Scanner{
		RpcClient: dom.NewRpcClient(node.Url),
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

	block, err := s.GetBlockByHeight(height, true)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}

	cnt, err := dao.GetBlockCountByHash(block.Items[0].Block.Txhash)
	if err != nil {
		return nil, fmt.Errorf("database err")
	}

	if cnt > 0 {
		return nil, fmt.Errorf("already have block , count : %d", cnt)
	}

	task := &DomProcTask{
		bestHeight: bestHeight,
		block: &dao.BlockInfo{
			Height:         block.Items[0].Block.Height,
			Hash:           block.Items[0].Block.Txhash,
			FrontBlockHash: block.Items[0].Block.Parenthash,
			Timestamp:      time.Unix(block.Items[0].Block.Blocktime, 0),
			Transactions:   len(block.Items[0].Block.Txs),
			Confirmations:  bestHeight - height + 1,
			CreateTime:     time.Now(),
		},
		txInfos: make([][]*dao.BlockTX, 0),
	}

	if task.block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}
	//处理区块内的交易
	if len(block.Items[0].Block.Txs) > 0 {
		if s.conf.EnableGoroutine {
			wg := &sync.WaitGroup{}
			wg.Add(len(block.Items[0].Block.Txs))
			for _, tmp := range block.Items[0].Block.Txs {
				tx := tmp
				go s.batchParseTx(&tx, bestHeight, block.Items[0].Block.Blocktime, task, wg)
			}
			wg.Wait()
		} else {
			for _, tx := range block.Items[0].Block.Txs {
				blockTx, err := s.parseBlockTX(&tx, bestHeight, block.Items[0].Block.Blocktime)
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
func (s *Scanner) batchParseTx(tx *dom.Txs, bestHeight, blockTimestamp int64, task *DomProcTask, w *sync.WaitGroup) {
	defer w.Done()
	blockTx, err := s.parseBlockTX(tx, bestHeight, blockTimestamp)
	if err == nil {
		s.lock.Lock()
		task.txInfos = append(task.txInfos, blockTx)
		s.lock.Unlock()
	}
}

func (s *Scanner) isValidTransaction(tx *dom.Txs) bool {
	if tx == nil {
		return false
	}
	txre, err := s.GetTransactionByHash(tx.Hash)
	if err != nil {
		return false
	}
	_, f64 := common.StrToFloat(tx.Payload.Transfer.Amount, common.Dec_8base)
	if tx.Execer == "coins" && tx.Payload.Ty == 1 && f64 > 0 && txre.Receipt.Ty == 2 {
		return true
	}
	return false
}

// 解析交易
func (s *Scanner) parseBlockTX(tx *dom.Txs, bestHeight, blockTimestamp int64) ([]*dao.BlockTX, error) {
	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}
	res := make([]*dao.BlockTX, 0)

	if s.isValidTransaction(tx) {
		//主币transfer的处理逻辑
		tx, err := s.GetTransactionByHash(tx.Hash)
		if err != nil {
			return nil, err
		}

		if tx.Receipt.Ty != 2 {
			log.Printf("%s 无效交易", tx.Tx.Hash)
			return nil, fmt.Errorf("%s 无效交易", tx.Tx.Hash)
		}

		hash, err := s.GetBlockHashByHeight(tx.Height)
		if err != nil {
			return nil, err
		}

		blocktx := &dao.BlockTX{
			BlockHeight: tx.Height,
			BlockHash:   hash,
			Txid:        tx.Tx.Hash,
			FromAddress: tx.Tx.From,
			Nonce:       tx.Tx.Nonce,
			//GasUsed:     tx.Tx.Fee,
			GasPrice:  tx.Tx.Fee,
			Input:     tx.Tx.Rawpayload,
			CoinName:  s.conf.Name,
			Decimal:   dom.WEI,
			Timestamp: time.Unix(blockTimestamp, 0),
			Amount:    decimal.NewFromBigInt(big.NewInt(tx.Tx.Amount), 0),
			ToAddress: tx.Tx.To,
		}
		res = append(res, blocktx)
	}
	return res, nil
}
