package yotta

import (
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"strings"
	"sync"
	"time"
	"chiaDataServer/common"
	"chiaDataServer/common/conf"
	"chiaDataServer/common/log"
	dao "chiaDataServer/models/po/yotta"
	"chiaDataServer/utils"
	rpc "chiaDataServer/utils/eos"
)

type Scanner struct {
	*rpc.RpcClient
	lock *sync.Mutex
	conf conf.SyncConfig
}

func NewScanner(conf conf.Config, node conf.NodeConfig) common.Scanner {
	return &Scanner{
		RpcClient: rpc.NewRpcClient(node.Url, node.RPCKey, node.RPCSecret),
		lock:      &sync.Mutex{},
		conf:      conf.Sync,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	dao.BlockRollBack(height)
	dao.TxRollBack(height)
}

func (s *Scanner) Init() error {
	return nil
}

func (s *Scanner) Clear() {
}

//var i = int64(60612620)

//获取最高区块高度
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i, nil
	count, err := s.GetBestHeight() //获取到的是区块个数
	return count, err
}

func (s *Scanner) GetCurrentBlockHeight() (int64, error) {
	return dao.MaxBlockHeight()
}

//扫描一个可逆的区块
func (s *Scanner) ScanReverseBlock(height, bestHeight int64) (common.ProcTask, error) {
	return s.scanBlock(height, bestHeight)
}

//扫描一个不可逆的区块
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
		task.irreversible = true
	}
	workpool := utils.NewWorkPool(10) //一次性发太多请求会让节点窒息

	for i, _ := range block.Transactions {
		workpool.Incr()
		go func(tx *rpc.TransactionReceipt, task *ProcTask) {
			defer workpool.Dec()
			s.batchParseTx(tx, task)
		}(&block.Transactions[i], task)

	}
	workpool.Wait()
	_ = starttime
	//log.Infof("scanBlock %d ,used time : %f
	return task, nil
}

//批量解析交易
func (s *Scanner) batchParseTx(tx *rpc.TransactionReceipt, task *ProcTask) {

	blockTxs, err := parseBlockTX(tx, task.Block)
	if err == nil {
		s.lock.Lock()
		defer s.lock.Unlock()
		task.TxInfos = append(task.TxInfos, blockTxs...)

	} else {
		log.Info(err.Error())
	}
}

// 解析交易
func parseBlockTX(tx *rpc.TransactionReceipt, block *dao.BlockInfo) ([]*dao.BlockTx, error) {

	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}

	if tx.Transaction.Packed == nil {
		return nil, fmt.Errorf("tx.Transaction.Packed is null")
	}

	if tx.Transaction.Packed.Transaction == nil {
		return nil, fmt.Errorf("tx.Transaction.Packed.Transaction is null")
	}
	if tx.Status!="executed"{
		return nil, fmt.Errorf("err tx.status")
	}
	var blocktxs = make([]*dao.BlockTx, 0)
	blocktx := dao.BlockTx{
		CoinName:    conf.Cfg.Name,
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
	//log.Info(tx.Transaction.ID.String(), string(data))
	for _, action := range tx.Transaction.Packed.Transaction.Actions {
		var blocktxTmp dao.BlockTx = blocktx
		if err := parseActionForBlocktx(&blocktxTmp, action); err != nil {
			//log.Info(err.Error())
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

func parseActionForBlocktx(blocktx *dao.BlockTx, action *rpc.Action) error {
	if err := action.MapToRegisteredAction(); err != nil {
		return fmt.Errorf("action MapToRegistered err : %v", err)
	}

	//if action.Name != "yrctransfer" { //srttransfer sdgtransfer
	//	return fmt.Errorf("don't support action name : %v", action.Name)
	//}
	if !strings.HasSuffix(string(action.Name), "transfer") {
		return fmt.Errorf("don't support action name : %v", action.Name)
	}
	//假充值验证
	if blocktx.Status != "executed" {
		return fmt.Errorf("tx.status != executed,(%v)", blocktx.Status)
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
