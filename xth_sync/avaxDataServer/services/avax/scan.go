package avax

import (
	"avaxDataServer/common"
	"avaxDataServer/common/log"
	"avaxDataServer/conf"
	dao "avaxDataServer/models/po/avax"
	"avaxDataServer/utils/avax"
	"runtime"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type Scanner struct {
	*avax.RpcClient

	lock            *sync.Mutex
	taskJobs        []*ScanTask
	jobsNum         int
	EnableGoroutine bool
	conf            conf.SyncConfig
}

func NewScanner(conf conf.Config, node conf.NodeConfig) common.Scanner {
	return &Scanner{
		RpcClient: avax.NewRpcClient(node.Url, node.Node, node.RPCKey, node.RPCSecret),
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
	return s.scanBlock2(height, bestHeight)
}

//扫描一个不可逆的区块
func (s *Scanner) ScanIrreverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock2(height, bestHeight)
}

func (s *Scanner) scanBlock2(height, bestHeight int64) (common.ProcTask, error) {
	starttime := time.Now()
GetTransactionByHeight:
	txs, err := s.RpcClient.TopTransaction(100)
	if err != nil {
		log.Info(err.Error())
		time.Sleep(time.Second)
		goto GetTransactionByHeight
		//return nil, err
	}
	var txinfos []*avax.Transaction
	for i, tx := range txs.Transactions {
		if tx.Timestamp.Unix() > height {
			//过滤锁定的交易
			lockedoutput := false
			for _, output := range tx.Outputs {
				if output.Locktime != 0 {
					lockedoutput = true
				}
			}
			if lockedoutput {
				continue
			}
			txinfos = append(txinfos, txs.Transactions[i])
		}
	}
	task := &AvaxProcTask{
		Irreversible: true,
		Height:       txs.Transactions[0].Timestamp.Unix(),
		BestHeight:   bestHeight,
		Confirms:     bestHeight - height + 13,
		TxInfos:      txinfos,
	}
	//log.Info(task.TxInfos[0].Timestamp)
	for i, _ := range task.TxInfos {
		for k, _ := range task.TxInfos[i].Inputs {
			td, err := decimal.NewFromString(task.TxInfos[i].Inputs[k].Output.Amount)
			if err != nil {
				task.TxInfos[i].Inputs[k].Output.Amount = "0"
				continue
			}
			task.TxInfos[i].Inputs[k].Output.Amount = td.Shift(-9).String()
		}
		for k, _ := range task.TxInfos[i].Outputs {
			td, err := decimal.NewFromString(task.TxInfos[i].Outputs[k].Amount)
			if err != nil {
				task.TxInfos[i].Outputs[k].Amount = "0"
				continue
			}
			task.TxInfos[i].Outputs[k].Amount = td.Shift(-9).String()
		}
	}
	log.Infof("ScanBlock : %d, txs : %d ,used time : %f 's", height, len(task.TxInfos), time.Since(starttime).Seconds())
	return task, nil
}
