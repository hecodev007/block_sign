package vet

import (
	"fmt"
	"log"
	"rsksync/common"
	"rsksync/conf"
	dao "rsksync/models/po/vet"
	"rsksync/utils/vet"
	"sync"
	"time"
)

type Scanner struct {
	*vet.VetHttpClient
	lock *sync.Mutex
	conf conf.SyncConfig
}

func NewScanner(conf conf.Config, node conf.NodeConfig) common.Scanner {

	return &Scanner{
		VetHttpClient: vet.NewVetHttpClient(node.Url),
		lock:          &sync.Mutex{},
		conf:          conf.Sync,
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
	return s.GetBestHeight()
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

	block, err := s.GetBlockByHeight(height)
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

	task := &VetProcTask{
		bestHeight: bestHeight,
		block: &dao.BlockInfo{
			Height:         block.Height,
			Hash:           block.Hash,
			FrontBlockHash: block.ParentHash,
			Timestamp:      time.Unix(block.Timestamp, 0),
			Transactions:   len(block.Txs),
			Confirmations:  bestHeight - height + 1,
			CreateTime:     time.Now(),
		},
	}

	if task.block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}

	//处理区块内的交易
	if len(block.Txs) > 0 {
		if s.conf.EnableGoroutine {
			wg := &sync.WaitGroup{}
			wg.Add(len(block.Txs))
			for _, tmp := range block.Txs {
				tx := tmp
				go s.batchParseTx(tx, task, wg)
			}
			wg.Wait()
		} else {
			for _, tx := range block.Txs {
				if txInfo, err := s.parseBlockTX(tx); err == nil {
					task.txInfos = append(task.txInfos, txInfo)
				}
			}
		}
	}

	log.Printf("scanBlock %d ,used time : %f 's", height, time.Since(starttime).Seconds())
	return task, nil
}

//批量解析交易
func (s *Scanner) batchParseTx(txid string, task *VetProcTask, w *sync.WaitGroup) {
	defer w.Done()
	if txInfo, err := s.parseBlockTX(txid); err == nil {
		s.lock.Lock()
		task.txInfos = append(task.txInfos, txInfo)
		s.lock.Unlock()
	}
}

// 解析交易
func (s *Scanner) parseBlockTX(txid string) (*TxInfo, error) {

	if txid == "" {
		return nil, fmt.Errorf("tx is null")
	}

	tx, err := s.GetTransaction(txid)
	if err != nil {
		return nil, err
	}

	txreceipt, err := s.GetTransactionReceipt(txid)
	if err != nil {
		return nil, err
	}

	return createTxDetail(tx, txreceipt)
}

func createTxDetail(tx *vet.Transaction, txreceipt *vet.TransactionReceipt) (txInfo *TxInfo, err error) {
	if tx == nil || txreceipt == nil {
		return nil, fmt.Errorf("tx or txreceipt don't allow nil")
	}

	txInfo = &TxInfo{}
	txInfo.tx = &dao.BlockTX{
		BlockHeight:  int64(tx.Meta.BlockNumber),
		BlockHash:    tx.Meta.BlockID,
		Txid:         tx.Txid,
		Origin:       tx.Origin,
		GasPriceCoef: int(tx.GasPriceCoef),
		Nonce:        tx.Nonce,
		Timestamp:    time.Unix(int64(tx.Meta.BlockTimestamp), 0),
		ChainTag:     int(tx.ChainTag),
		TxCount:      len(tx.Clauses),
		Status:       0,
		CreateTime:   time.Now(),
	}

	if !txreceipt.Reverted {
		txInfo.tx.Status = 1
	}

	//txInfo.tx.TxCount = len(txreceipt.Outputs)
	txInfo.tx.GasUsed = txreceipt.GasUsed
	txInfo.tx.PaidVTHO = txreceipt.Paid
	txInfo.tx.RewardVTHO = txreceipt.Reward

	for _, detail := range txreceipt.Outputs {
		if len(detail.Transfers) > 0 {
			for _, t := range detail.Transfers {
				d := &dao.BlockTxDetail{
					CoinName:    "vet",
					Txid:        tx.Txid,
					BlockHeight: txInfo.tx.BlockHeight,
					BlockHash:   txInfo.tx.BlockHash,
					FromAddress: t.GetSender(),
					ToAddress:   t.GetRecipient(),
					Status:      txInfo.tx.Status,
					Timestamp:   txInfo.tx.Timestamp,
					CreateTime:  time.Now(),
				}
				if d.Amount, err = t.GetAmount(); err != nil {
					log.Printf("%s transfer err : %v", tx.Txid, err)
					continue
				}
				txInfo.details = append(txInfo.details, d)
			}
		}

		if len(detail.Events) > 0 {
			for _, event := range detail.Events {
				if event.Address != vet.VTHOContract {
					//log.Printf("%s don't support contract :%s", tx.Txid, event.Address)
					continue
				}

				d := &dao.BlockTxDetail{
					CoinName:        "vtho",
					Txid:            tx.Txid,
					BlockHeight:     txInfo.tx.BlockHeight,
					BlockHash:       txInfo.tx.BlockHash,
					ContractAddress: event.Address,
					Status:          txInfo.tx.Status,
					Timestamp:       txInfo.tx.Timestamp,
					CreateTime:      time.Now(),
				}
				if d.FromAddress, err = event.GetSender(); err != nil {
					log.Printf("get sender err %v", err)
					continue
				}
				if d.ToAddress, err = event.GetRecipient(); err != nil {
					log.Printf("get receipient err %v", err)
					continue
				}
				if d.Amount, err = event.GetAmount(); err != nil {
					log.Printf("get amount err %v", err)
					continue
				}
				txInfo.details = append(txInfo.details, d)
			}
		}
	}

	return txInfo, nil
}
