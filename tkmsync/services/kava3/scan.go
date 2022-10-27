package kava3

import (
	"fmt"
	"log"
	"rsksync/common"
	"rsksync/conf"
	dao "rsksync/models/po/kava"
	"rsksync/utils/kava3"
	"runtime"
	"sync"
	"time"
)

type ScanTask struct {
	Txids chan string
	Done  chan int
}

type Scanner struct {
	client   *kava3.HttpClient
	lock     *sync.Mutex
	conf     conf.SyncConfig
	taskJobs []*ScanTask
	jobsNum  int
}

func NewScanner(conf conf.Config, node conf.NodeConfig) common.Scanner {
	return &Scanner{
		client:  kava3.NewHttpClient(node.Url),
		lock:    &sync.Mutex{},
		jobsNum: runtime.NumCPU(),
		conf:    conf.Sync,
	}
}
func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	dao.DeleteBlockInfo(height)
	dao.DeleteBlockTX(height)
}

//爬数据
func (s *Scanner) Init() error {
	for i := 0; i < s.jobsNum; i++ {
		s.taskJobs = append(s.taskJobs, &ScanTask{
			Txids: make(chan string, 10000),
			Done:  make(chan int, 10000),
		})
	}
	return nil
}
func (s *Scanner) Clear() {
	for i := 0; i < s.jobsNum; i++ {
		close(s.taskJobs[i].Txids)
		close(s.taskJobs[i].Done)
	}
}
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	return s.client.GetLastBlockHeight()
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
	for i := int64(0); i < count; i++ {
		go func(height int64, w *sync.WaitGroup) {
			w.Add(1)
			defer w.Done()
			task, err := s.ScanIrreverseBlock(height, bestHeight)
			if err == nil {
				taskmap.Store(height, task)
			}
		}(startHeight+i, wg)
	}
	wg.Wait()
	log.Printf("***batchScanBlocks used time : %f 's", time.Since(starttime).Seconds())
	return taskmap
}
func (s *Scanner) ScanReverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock(height, bestHeight)
}
func (s *Scanner) ScanIrreverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock(height, bestHeight)
}

//扫描一个区块
func (s *Scanner) scanBlock(height, bestHeight int64) (common.ProcTask, error) {
	starttime := time.Now()
	block, err := s.client.GetBlockByHeight(height)
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
	task := &ProcTask{
		irreversible: false,
		bestHeight:   bestHeight,
		block: &dao.BlockInfo{
			Height:         block.Height,
			Hash:           block.Hash,
			ChainID:        block.ChainID,
			FrontBlockHash: block.ParentHash,
			Timestamp:      block.Timestamp,
			Transactions:   len(block.Transactions),
			Confirmations:  bestHeight - height + 1,
			CreateTime:     time.Now(),
		},
	}
	if task.block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}
	//处理区块内的交易
	if len(block.Transactions) > 0 {
		if s.conf.EnableGoroutine {
			//初始化需要并发的任务
			for i, txid := range block.Transactions {
				index := i % s.jobsNum
				s.taskJobs[index].Txids <- txid
			}
			//开始执行并发任务
			for i := 0; i < s.jobsNum; i++ {
				index := i
				go s.batchParseTx(s.taskJobs[index].Txids, s.taskJobs[index].Done, block.Hash, block.Timestamp, task)
			}
			//等待所有执行结果
			for i := 0; i < s.jobsNum; i++ {
				index := i
				<-s.taskJobs[index].Done
			}
		} else {
			for _, txid := range block.Transactions {
				if tx, err := s.client.GetTransactionByHash(txid, block.Timestamp); err == nil {
					if txInfo, err := parseBlockTX(tx, block.Hash); err == nil {
						task.txInfos = append(task.txInfos, txInfo)
					}
				}
			}
		}
	}
	log.Printf("ScanBlock : %d, txs : %d ,used time : %f 's", height, len(block.Transactions), time.Since(starttime).Seconds())
	return task, nil
}

//批量解析交易
func (s *Scanner) batchParseTx(jobs <-chan string, results chan<- int, blockhash string, blocktime time.Time, task *ProcTask) {
	count := len(jobs)
	offset := 0
	for i := 0; i < count; i++ {
		select {
		case txid := <-jobs:
			offset += 1
			tx, err := s.client.GetTransactionByHash(txid, blocktime)
			if err != nil {
				log.Printf("GetRawTransaction txid: %s , err: %v", txid, err)
				continue
			}
			txInfo, err := parseBlockTX(tx, blockhash)
			if err != nil {
				log.Printf("parseBlockTX %v", err)
				continue
			}
			s.lock.Lock()
			task.txInfos = append(task.txInfos, txInfo)
			s.lock.Unlock()
			if offset >= count {
				break
			}
		default:
			offset += 1
			if offset >= count {
				break
			}
		}
	}
	results <- 1
}

// 解析交易
func parseBlockTX(tx *kava3.Transaction, blockhash string) (*TxInfo, error) {
	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}
	txInfo := &TxInfo{}
	txInfo.tx = &dao.BlockTX{
		Txid:        tx.Hash,
		BlockHeight: tx.BlockHeight,
		BlockHash:   blockhash,
		Fee:         tx.Fee,
		GasUsed:     tx.GasUsed,
		GasWanted:   tx.GasWanted,
		RawLogs:     tx.RawLogs,
		Type:        tx.Type,
		Memo:        tx.Memo,
		Timestamp:   tx.Timestamp,
		CreateTime:  time.Now(),
		MsgCount:    len(tx.TxMsgs),
	}
	for _, msg := range tx.TxMsgs {
		txmsg := &dao.BlockTXMsg{
			Txid:        tx.Hash,
			Index:       msg.Index,
			BlockHeight: tx.BlockHeight,
			BlockHash:   blockhash,
			FromAddress: msg.From,
			ToAddress:   msg.To,
			Amount:      msg.Amount,
			Log:         msg.Log,
			Type:        msg.Type,
			Timestamp:   txInfo.tx.Timestamp,
			CreateTime:  txInfo.tx.CreateTime,
			UnlockTime:  msg.UnlockTime,
		}
		//if msg.Success {
		//	txmsg.Status = 1
		//}
		txmsg.Status = 1
		txInfo.txmsgs = append(txInfo.txmsgs, txmsg)
	}
	log.Printf("parseBlockTX %s ,tx num : %d : %d", tx.Hash, len(tx.TxMsgs), len(txInfo.txmsgs))
	return txInfo, nil
}
