package atp

import (
	"atpDataServer/common"
	"atpDataServer/common/conf"
	"atpDataServer/common/log"
	dao "atpDataServer/models/po/cfx"
	"atpDataServer/services"
	rpc "atpDataServer/utils/atp"
	"encoding/json"
	"errors"
	"github.com/shopspring/decimal"
	"sync"
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
	dao.BlockRollBack(height)
	dao.TxRollBack(height)
}

func (s *Scanner) Init() error {
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
	return count - 18, err
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
		//time.Sleep(time.Second * 3)
		goto retryGetBlockByHeight
	}

	task := &ProcTask{
		Irreversible: bestHeight-height >= s.conf.Confirmations,
		BestHeight:   bestHeight,
		Block: &dao.BlockInfo{
			Height:            int64(block.Number),
			Hash:              block.Hash,
			Previousblockhash: block.ParentHash,
			Transactions:      len(block.Transactions),
			Confirmations:     bestHeight - height + 1, //block.Confirmations
			Timestamp:         time.Now(),
			Createtime:        time.Now(),
		},
	}

	//并发处理区块内的交易

	lock := new(sync.Mutex)
	for _, tx := range block.Transactions {
		if tx.To == "" || tx.To == "0x" {
			continue
		}

		if txInfo, err := parseBlockRawTX(s.RpcClient, s.watch, &tx, block.Hash, int64(block.Number)); err != nil {
			//log.Info(err.Error())
		} else if txInfo != nil {
			lock.Lock() //append并发不安全
			task.TxInfos = append(task.TxInfos, txInfo)
			lock.Unlock()
		}
	}
	//ts, _ := json.Marshal(task)
	//log.Info("task", block.Height.ToInt().Int64(), string(ts), "task")
	return task, nil
}
func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}

//解析交易
func parseBlockRawTX(RpcClient *rpc.RpcClient, watch *services.WatchControl, tx *rpc.Transaction, blockhash string, blockheight int64) (*dao.BlockTx, error) {
	//log.Info(tx.Hash)
	//RpcClient.IsUser(tx.To.String())
TransactionReceipt:
	receipt, err := RpcClient.TransactionReceipt(tx.Hash)
	//log.Info(String(receipt))
	if err != nil {
		log.Info(err.Error())
		goto TransactionReceipt
	}
	if receipt.BlockHash != "" && int64(receipt.Status) != 1 {
		log.Info(tx.Hash+" tx.status != success", receipt.Status.String())
		return nil, errors.New("tx.status != success")
	} else {
		//log.Error(tx.Hash+" tx.status != success", receipt.Status.String(), int64(receipt.Status))
		//panic(tx.Hash+" tx.status != success")
	}
	//log.Info(tx.Hash, int64(receipt.GasUsed))
	blocktx := &dao.BlockTx{
		Txid:            tx.Hash,
		CoinName:        conf.Cfg.Sync.Name,
		ContractAddress: "",
		FromAddress:     tx.From,
		ToAddress:       tx.To,
		BlockHeight:     blockheight,
		BlockHash:       blockhash,
		Amount:          decimal.NewFromBigInt(tx.Value.ToInt(), -18).String(),
		Status:          1,
		GasPrice:        int64(tx.GasPrice),
		GasUsed:         int64(receipt.GasUsed),
		Fee:             decimal.NewFromInt(int64(tx.GasPrice)).Mul(decimal.NewFromInt(int64(receipt.GasUsed))).Shift(-18).String(),
		Nonce:           int64(tx.Nonce),
		Input:           tx.Input,
		Logs:            "",
		Timestamp:       time.Now(),
		CreateTime:      time.Now(),
	}
	return blocktx, nil
}
