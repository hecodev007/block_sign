package telos

import (
	"waxsync/common"
	"waxsync/common/conf"
	"waxsync/common/log"
	dao "waxsync/models/po/telos"
	"waxsync/services"
	"waxsync/utils/eos"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type Scanner struct {
	*eos.API
	lock  *sync.Mutex
	conf  conf.SyncConfig
	Watch *services.WatchControl
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	return &Scanner{
		API:   eos.NewAPI(node.Url, node.Url2),
		lock:  &sync.Mutex{},
		conf:  conf.Sync,
		Watch: watch,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	if err := dao.DeleteBlockInfo(height); err != nil {
		panic(err.Error())
	}
	if err := dao.DeleteBlockTX(height); err != nil {
		panic(err.Error())
	}
}

func (s *Scanner) Init() error {
	if conf.Cfg.Sync.EnableRollback {
		s.Rollback(conf.Cfg.Sync.RollHeight)
		log.Info("rollback success")
	}
	return nil
}

func (s *Scanner) Clear() {
}

func (s *Scanner) GetBestBlockHeight() (h int64, err error) {
	h, err = s.GetBestHeight()
	return h - conf.Cfg.Sync.Delaycount, err
}

func (s *Scanner) GetCurrentBlockHeight() (int64, error) {
	return dao.MaxBlockHeight()
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
	GetBlockByNum:
	block, err := s.GetBlockByNum(height)
	if err != nil {
		log.Infof("GetBlockByNumber %d  , err : %v", height, err)
		//time.Sleep(5*time.Second)
		goto GetBlockByNum
	}

	cnt, err := dao.GetBlockCountByHash(block.ID.String())
	if err != nil {
		return nil, fmt.Errorf("database err:"+err.Error())
	}

	if cnt > 0 {
		return nil, fmt.Errorf("already have Block , count : %d", cnt)
	}

	task := &ProcTask{
		BestHeight: bestHeight,
		Block: &dao.BlockInfo{
			Height:            int64(block.BlockNum),
			Hash:              block.ID.String(),
			Previousblockhash: block.Previous.String(),
			Timestamp:         block.Timestamp.Time,
			Transactions:      len(block.Transactions),
			Confirmations:     bestHeight - height + 1,
			Createtime:        time.Now(),
		},
	}

	if task.Block.Confirmations >= s.conf.Confirmations {
		task.Irreversible = true
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
				blockTx, err := s.parseBlockTX(&tx, task.Block)
				if err == nil && len(blockTx) > 0 {
					task.TxInfos = append(task.TxInfos, blockTx...)
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

	blockTxs, err := s.parseBlockTX(tx, task.Block)
	if err == nil {
		s.lock.Lock()
		task.TxInfos = append(task.TxInfos, blockTxs...)
		s.lock.Unlock()
	}
}

// 解析交易
func (s *Scanner) parseBlockTX(tx *eos.TransactionReceipt, block *dao.BlockInfo) ([]*dao.BlockTx, error) {

	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}

	if tx.Transaction.Packed == nil {
		return nil, fmt.Errorf("tx.Transaction.Packed is null")
	}

	if tx.Transaction.Packed.Transaction == nil {
		return nil, fmt.Errorf("tx.Transaction.Packed.Transaction is null")
	}
	var blocktxs = make([]*dao.BlockTx, 0)
	blocktx := dao.BlockTx{
		CoinName:    "", //s.conf.Name
		Txid:        tx.Transaction.ID.String(),
		BlockHeight: block.Height,
		BlockHash:   block.Hash,
		Status:      tx.Status,
		Timestamp:   block.Timestamp,
		Createtime:  time.Now(),
	}

	//if blocktx.Status == "delayed" {
	//	if tx, err := s.GetTransactionFromThird(blocktx.Txid); err == nil && tx != nil {
	//		blocktx.Status = tx.Receipt.Status
	//	}
	//}
	data, _ := json.Marshal(tx)
	for _, action := range tx.Transaction.Packed.Transaction.Actions {
		var blocktxTmp dao.BlockTx = blocktx
		if err := parseActionForBlocktx(s.Watch, &blocktxTmp, action); err != nil {
			continue
		} else if blocktxTmp.FromAddress == "" || blocktxTmp.ToAddress == "" {
			continue
		} else {
			if len(data) < 8000 {
				blocktxTmp.Txjson = string(data)
			}
			blocktxs = append(blocktxs, &blocktxTmp)
		}

	}

	//if blocktx.FromAddress == "" || blocktx.ToAddress == "" {
	//	return nil, fmt.Errorf("tx. from : %s , to :%s", blocktx.FromAddress, blocktx.ToAddress)
	//}

	return blocktxs, nil
}

func parseActionForBlocktx(watch *services.WatchControl, blocktx *dao.BlockTx, action *eos.Action) error {
	if err := action.MapToRegisteredAction(); err != nil {
		//log.Infof("action MapToRegistered err : %v", err)
		return fmt.Errorf("action MapToRegistered err : %v", err)
	}

	if action.Name != "transfer" {
		return fmt.Errorf("don't support action name : %v", action.Name)
	}

	blocktx.ContractAddress = string(action.Account)

	if name, _, err := watch.GetContractNameAndDecimal(blocktx.ContractAddress); err == nil {
		//log.Info(blocktx.CoinName, name)
		blocktx.CoinName = name
	}
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
				if blocktx.CoinName == "" {
					blocktx.CoinName = strs[1]
				}
			}
		}
	}

	return nil
}
