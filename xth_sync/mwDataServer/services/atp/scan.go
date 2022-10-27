package atp

import (
	"encoding/json"
	"errors"
	"mwDataServer/common"
	"mwDataServer/common/conf"
	"mwDataServer/common/log"
	dao "mwDataServer/models/po/cfx"
	"mwDataServer/services"
	rpc "mwDataServer/utils/atp"
	"time"
)

type Scanner struct {
	*rpc.RpcClient
	conf  conf.SyncConfig
	watch *services.WatchControl
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	return &Scanner{
		RpcClient: rpc.NewRpcClient(node.Url, node.RPCKey, node.RPCSecret),
		conf:      conf.Sync,
		watch:     watch,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	if _, err := dao.BlockRollBack(height); err != nil {
		panic(err.Error())
	}
	if _, err := dao.TxRollBack(height); err != nil {
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

//var i = int64(781001)

//获取最高区块高度
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i, nil
	count, err := s.GetBlockCount() //获取到的是区块个数
	if count-conf.Cfg.Sync.Delaycount <= 0 {
		return 0, errors.New("GetBestBlockHeight error")
	}
	return count - conf.Cfg.Sync.Delaycount, err
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
	//starttime := time.Now()
retryGetBlockByHeight:
	block, err := s.GetBlockByHeight(height)
	if err != nil {
		log.Info("%v height:%v", err.Error(), height)
		time.Sleep(time.Second * 3)
		goto retryGetBlockByHeight
	}

	task := &ProcTask{
		Irreversible: bestHeight-height >= s.conf.Confirmations,
		BestHeight:   bestHeight,
		Block: &dao.BlockInfo{
			Height:            block.Height,
			Hash:              block.Hash,
			Previousblockhash: block.ParentHash,
			Transactions:      len(block.Transactions),
			Confirmations:     bestHeight - height + 1, //block.Confirmations
			Timestamp:         time.Now(),
			Createtime:        time.Now(),
		},
	}

	//并发处理区块内的交易

	//lock := new(sync.Mutex)
	for _, tx := range block.Transactions {
		if tx.Type != 0 {
			continue
		}

		if txInfo, err := parseBlockRawTX(s.RpcClient, s.watch, tx, block.Hash, int64(block.Height)); err != nil {
			//log.Info(err.Error())
		} else if txInfo != nil {
			//lock.Lock() //append并发不安全
			task.TxInfos = append(task.TxInfos, txInfo)
			//lock.Unlock()
		}
	}
	ts, _ := json.Marshal(task)
	log.Info("task", block.Height, string(ts), "task")
	return task, nil
}

//解析交易
func parseBlockRawTX(RpcClient *rpc.RpcClient, watch *services.WatchControl, tx *rpc.Transaction, blockhash string, blockheight int64) (*dao.BlockTx, error) {

	blocktx := &dao.BlockTx{
		Txid:          tx.Hash,
		CoinName:      conf.Cfg.Sync.Name,
		FromAddress:   tx.From,
		ToAddress:     tx.To,
		BlockHeight:   blockheight,
		BlockHash:     blockhash,
		Amount:        tx.Value.Shift(-8).String(),
		Status:        1,
		Type:          tx.Type,
		Fee:           tx.Fee.Shift(-8).String(),
		Timestamp:     time.Now(),
		CreateTime:    time.Now(),
		Confirmations: tx.Confirmations,
		Deadline:      tx.Deadline,
	}
	return blocktx, nil
}
