package telos

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"strings"
	"sync"
	"time"
	"zcashDataServer/common"
	"zcashDataServer/common/log"
	"zcashDataServer/conf"
	dao "zcashDataServer/models/po/telos"
	"zcashDataServer/utils/eos"
)

type Scanner struct {
	*eos.API
	lock *sync.Mutex
	conf conf.SyncConfig
}

func NewScanner(conf conf.Config, node conf.NodeConfig) common.Scanner {
	return &Scanner{
		API:  eos.NewAPI(node.Url, node.RPCKey),
		lock: &sync.Mutex{},
		conf: conf.Sync,
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
	log.Debugf("***batchScanBlocks used time : %f 's", time.Since(starttime).Seconds())
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
	//log.Infof("scanBlock %d ", height)
	block, err := s.GetBlockByNum(height)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}

	cnt, err := dao.GetBlockCountByHash(block.ID.String())
	if err != nil {
		return nil, fmt.Errorf("database err")
	}

	if cnt > 0 {
		return nil, fmt.Errorf("already have block , count : %d", cnt)
	}

	task := &ProcTask{
		bestHeight: bestHeight,
		block: &dao.BlockInfo{
			Height:         int64(block.BlockNum),
			Hash:           block.ID.String(),
			FrontBlockHash: block.Previous.String(),
			Timestamp:      block.Timestamp.Time,
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
			wg := &sync.WaitGroup{}
			wg.Add(len(block.Transactions))
			for _, tmp := range block.Transactions {
				tx := tmp
				go s.batchParseTx(&tx, task, wg)
			}
			wg.Wait()
		} else {
			for _, tx := range block.Transactions {
				blockTx, err := s.parseBlockTX(&tx, task.block)
				if err == nil {
					task.txInfos = append(task.txInfos, blockTx)
				}
			}
		}
	}
	_ = starttime
	//log.Infof("scanBlock %d ,used time : %f 's", height, time.Since(starttime).Seconds())
	return task, nil
}

//批量解析交易
func (s *Scanner) batchParseTx(tx *eos.TransactionReceipt, task *ProcTask, w *sync.WaitGroup) {
	defer w.Done()

	blockTx, err := s.parseBlockTX(tx, task.block)
	if err == nil {
		s.lock.Lock()
		task.txInfos = append(task.txInfos, blockTx)
		s.lock.Unlock()
	}
}

// 解析交易
func (s *Scanner) parseBlockTX(tx *eos.TransactionReceipt, block *dao.BlockInfo) (*dao.BlockTX, error) {

	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}

	if tx.Transaction.Packed == nil {
		return nil, fmt.Errorf("tx.Transaction.Packed is null")
	}

	if tx.Transaction.Packed.Transaction == nil {
		return nil, fmt.Errorf("tx.Transaction.Packed.Transaction is null")
	}

	blocktx := &dao.BlockTX{
		CoinName:    s.conf.Name,
		Txid:        tx.Transaction.ID.String(),
		BlockHeight: block.Height,
		BlockHash:   block.Hash,
		Status:      tx.Status,
		Timestamp:   block.Timestamp,
		CreateTime:  time.Now(),
	}

	//if blocktx.Status == "delayed" {
	//	if tx, err := s.GetTransactionFromThird(blocktx.Txid); err == nil && tx != nil {
	//		blocktx.Status = tx.Receipt.Status
	//	}
	//}

	for _, action := range tx.Transaction.Packed.Transaction.Actions {
		if err := parseActionForBlocktx(blocktx, action); err != nil {

		}
	}

	if blocktx.FromAddress == "" || blocktx.ToAddress == "" {
		return nil, fmt.Errorf("tx. from : %s , to :%s", blocktx.FromAddress, blocktx.ToAddress)
	}

	data, _ := json.Marshal(tx)
	if len(data) < 8000 {
		blocktx.TxJson = string(data)
	}

	return blocktx, nil
}

func parseActionForBlocktx(blocktx *dao.BlockTX, action *eos.Action) error {
	if err := action.MapToRegisteredAction(); err != nil {
		//log.Infof("action MapToRegistered err : %v", err)
		return fmt.Errorf("action MapToRegistered err : %v", err)
	}

	if action.Name != "transfer" {
		return fmt.Errorf("don't support action name : %v", action.Name)
	}

	blocktx.ContractAddress = string(action.Account)

	for k, v := range action.Data.(map[string]interface{}) {
		switch k {
		case "from":
			from, ok := v.(string)
			if !ok {
				return fmt.Errorf("from address type err: %T", v)
			}
			blocktx.FromAddress = from
		case "to":
			to, ok := v.(string)
			if !ok {
				return fmt.Errorf("to address type err: %T", v)
			}
			blocktx.ToAddress = to
		case "memo":
			memo, ok := v.(string)
			if !ok {
				return fmt.Errorf("memo type err: %T", v)
			}
			blocktx.Memo = memo
		case "quantity":
			{
				strs := strings.Split(v.(string), " ")
				blocktx.Amount, _ = decimal.NewFromString(strs[0])
				blocktx.CoinName = strs[1]
			}
		}
	}

	return nil
}
