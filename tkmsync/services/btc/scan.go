package btc

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"log"
	"rsksync/common"
	"rsksync/conf"
	dao "rsksync/models/po/btc"
	"rsksync/utils/btc"
	"runtime"
	"sync"
	"time"
)

type Scanner struct {
	*btc.RpcClient

	lock            *sync.Mutex
	taskJobs        []*ScanTask
	jobsNum         int
	EnableGoroutine bool
	conf            conf.SyncConfig
}

func NewScanner(conf conf.Config, node conf.NodeConfig) common.Scanner {
	return &Scanner{
		RpcClient: btc.NewRpcClient(node.Url, node.RPCKey, node.RPCSecret),
		lock:      &sync.Mutex{},
		jobsNum:   runtime.NumCPU(),
		conf:      conf.Sync,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	dao.DeleteBlockInfo(height)

	dao.DeleteBlockTX(height)

	dao.DeleteBlockTXVout(height)
}

func (s *Scanner) Init() error {
	for i := 0; i < s.jobsNum; i++ {
		s.taskJobs = append(s.taskJobs, &ScanTask{
			Txids: make(chan string, 50000),
			Done:  make(chan int, 50000),
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
	return s.GetBlockCount()
}

func (s *Scanner) GetCurrentBlockHeight() (int64, error) {
	return dao.GetMaxBlockIndex()
}

func (s *Scanner) BatchScanIrreverseBlocks(startHeight, endHeight, bestHeight int64) *sync.Map {
	return nil
}

//扫描一个可逆的区块
func (s *Scanner) ScanReverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	if s.conf.Name == "doge" {
		return s.scanBlock1(height, bestHeight)
	}
	return s.scanBlock2(height, bestHeight)
}

//扫描一个不可逆的区块
func (s *Scanner) ScanIrreverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	if s.conf.Name == "doge" {
		return s.scanBlock1(height, bestHeight)
	}
	return s.scanBlock2(height, bestHeight)
}

func (s *Scanner) scanBlock1(height, bestHeight int64) (common.ProcTask, error) {
	starttime := time.Now()

	block, err := s.GetBlockByHeight(height)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByHeight, err : %v", err)
	}

	log.Printf("GetBlockByHeight : %d, txs : %d ", height, len(block.Txs))

	cnt, err := dao.GetBlockCountByHash(block.Hash)
	if err != nil {
		return nil, fmt.Errorf("database err")
	}

	if cnt > 0 {
		return nil, fmt.Errorf("already have block height: %d, hash: %s , count : %d", block.Height, block.Hash, cnt)
	}

	task := &BtcProcTask{
		irreversible: false,
		bestHeight:   bestHeight,
		block: &dao.BlockInfo{
			Height:         block.Height,
			Hash:           block.Hash,
			Version:        block.Version,
			FrontBlockHash: block.PreviousBlockHash,
			NextBlockHash:  block.NextBlockHash,
			Timestamp:      time.Unix(block.Time, 0),
			Transactions:   len(block.Txs),
			Confirmations:  block.Confirmations,
			CreateTime:     time.Now(),
		},
	}

	if task.block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}

	//处理区块内的交易
	if len(block.Txs) > 0 {
		if s.conf.EnableGoroutine {
			//初始化需要并发的任务
			for i, txid := range block.Txs {
				index := i % s.jobsNum
				s.taskJobs[index].Txids <- txid
			}

			//开始执行并发任务
			for i := 0; i < s.jobsNum; i++ {
				index := i
				go s.batchParseRawTx(s.taskJobs[index].Txids, s.taskJobs[index].Done, block.Hash, height, task)
			}

			//等待所有执行结果
			for i := 0; i < s.jobsNum; i++ {
				index := i
				<-s.taskJobs[index].Done
			}
		} else {
			for _, txid := range block.Txs {
				tx, err := s.GetRawTransaction(txid)
				if err == nil {
					txInfo, err := parseBlockRawTX(s.conf.Name, &tx, block.Hash, height)
					if err == nil {
						task.txInfos = append(task.txInfos, txInfo)
					}
				}
			}
		}
	}

	log.Printf("ScanBlock : %d, txs : %d ,used time : %f 's", height, len(block.Txs), time.Since(starttime).Seconds())
	return task, nil
}

func (s *Scanner) scanBlock2(height, bestHeight int64) (common.ProcTask, error) {
	starttime := time.Now()

	block, err := s.GetBlockByHeight2(height)
	if err != nil {
		log.Println(err.Error())
		return nil, fmt.Errorf("GetBlockByHeight, err : %v", err)
	}

	log.Printf("GetBlockByHeight : %d, txs : %d ", height, len(block.Txs))

	cnt, err := dao.GetBlockCountByHash(block.Hash)
	if err != nil {
		return nil, fmt.Errorf("database err")
	}

	if cnt > 0 {
		return nil, fmt.Errorf("already have block height: %d, hash: %s , count : %d", block.Height, block.Hash, cnt)
	}

	task := &BtcProcTask{
		irreversible: false,
		bestHeight:   bestHeight,
		block: &dao.BlockInfo{
			Height:         block.Height,
			Hash:           block.Hash,
			Version:        block.Version,
			FrontBlockHash: block.PreviousBlockHash,
			NextBlockHash:  block.NextBlockHash,
			Timestamp:      time.Unix(block.Time, 0),
			Transactions:   len(block.Txs),
			Confirmations:  block.Confirmations,
			CreateTime:     time.Now(),
		},
	}

	if task.block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}

	//处理区块内的交易
	if len(block.Txs) > 0 {
		for _, tx := range block.Txs {
			if txInfo, err := parseBlockRawTX(s.conf.Name, tx, block.Hash, height); err == nil {
				task.txInfos = append(task.txInfos, txInfo)
			}
		}
	}

	log.Printf("ScanBlock : %d, txs : %d ,used time : %f 's", height, len(block.Txs), time.Since(starttime).Seconds())
	return task, nil
}

//批量解析交易
func (s *Scanner) batchParseRawTx(jobs <-chan string, results chan<- int, blockhash string, height int64, task *BtcProcTask) {
	count := len(jobs)
	offset := 0
	for i := 0; i < count; i++ {
		select {
		case txid := <-jobs:
			offset += 1

			tx, err := s.GetRawTransaction(txid)
			if err != nil {
				log.Printf("GetRawTransaction txid: %s , err: %v", txid, err)
				continue
			}

			txInfo, err := parseBlockRawTX(s.conf.Name, &tx, blockhash, height)
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

//解析交易
func parseBlockRawTX(coinName string, tx *btc.Transaction, blockhash string, height int64) (*TxInfo, error) {
	var vouts []*dao.BlockTXVout
	var vins []*dao.BlockTXVout

	if tx == nil {
		return nil, fmt.Errorf("txid is null")
	}

	//log.Printf("parse tx : %s",tx.Txid)
	blocktx := &dao.BlockTX{
		Txid:        tx.Txid,
		BlockHeight: height,
		BlockHash:   blockhash,
		Version:     tx.Version,
		Size:        tx.Size,
		VoutCount:   len(tx.Vout),
		VinCount:    len(tx.Vin),
		Timestamp:   time.Unix(tx.Time, 0),
		CreateTime:  time.Now(),
	}

	for _, vout := range tx.Vout {
		blocktxvout := &dao.BlockTXVout{
			Txid:       blocktx.Txid,
			Voutn:      vout.Index,
			BlockHash:  blocktx.BlockHash,
			Value:      decimal.NewFromFloat(vout.Value),
			Status:     0,
			Timestamp:  blocktx.Timestamp,
			CreateTime: time.Now(),
		}
		if address, err := vout.ScriptPubkey.GetAddress(); err == nil {
			blocktxvout.Address = address[0]
		}
		data, _ := json.Marshal(vout.ScriptPubkey)
		blocktxvout.ScriptPubKey = string(data)

		vouts = append(vouts, blocktxvout)
	}

	for _, vin := range tx.Vin {

		if vin.Txid == "" {
			vin.Txid = "coinbase"
			blocktx.Coinbase = 1
			continue
		}

		blocktxvin := &dao.BlockTXVout{
			Txid:      vin.Txid,
			Voutn:     vin.Vout,
			SpendTxid: blocktx.Txid,
			Status:    1,
		}

		vins = append(vins, blocktxvin)
	}

	if blocktx.Coinbase == 1 {
		for _, vout := range vouts {
			vout.Coinbase = 1
		}
	}

	return &TxInfo{
		tx:    blocktx,
		vouts: vouts,
		vins:  vins,
	}, nil
}
