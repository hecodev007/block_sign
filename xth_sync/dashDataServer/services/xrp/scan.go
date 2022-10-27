package xrp

import (
	"dashDataServer/common"
	"dashDataServer/common/conf"
	"dashDataServer/common/log"
	dao "dashDataServer/models/po/xrp"
	rpc "dashDataServer/utils/xrp"
	"fmt"
	"strconv"
	"time"
)

type Scanner struct {
	*rpc.RpcClient
	conf conf.SyncConfig
}

func NewScanner(conf conf.Config, node conf.NodeConfig) common.Scanner {
	return &Scanner{
		RpcClient: rpc.NewRpcClient(node.Url, node.RPCKey, node.RPCSecret),
		conf:      conf.Sync,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	dao.BlockRollBack(height)
	dao.TxRollBack(height)
	dao.TokenTxRollBack(height)
}

func (s *Scanner) Init() error {
	return nil
}

func (s *Scanner) Clear() {
}

//var i = int64(58044277)

func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i, nil
	return s.BlockHeight()
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
retryGetBlock:
	block, err := s.GetFullBlock(height)
	if err != nil {
		log.Warnf("%v height:%v", err.Error(), height)
		time.Sleep(time.Second * 3)
		goto retryGetBlock
		//return nil, fmt.Errorf("GetBlockByHeight, err : %v", err)
	}

	//log.Infof("GetBlockByHeight : %d, txs : %d ", height, len(block.Txs))
	if has, err := dao.BlockHashExist(block.Hash); err != nil {
		return nil, fmt.Errorf("database err")
	} else if has {
		return nil, fmt.Errorf("already have block height: %d, hash: %s , count : %d", block.Height, block.Hash, 1)
	}

	task := &ProcTask{
		Irreversible: bestHeight-height >= s.conf.Confirmations,
		BestHeight:   bestHeight,
		Block: &dao.BlockInfo{
			Height:            block.Height,
			Hash:              block.Hash,
			Previousblockhash: block.ParentHash,
			Nextblockhash:     "",
			Transactions:      len(block.Block.Transactions),
			Confirmations:     bestHeight - height + 1, //block.Confirmations
			Time:              block.Block.Time,
		},
	}

	//处理区块内的交易
	if len(block.Transacitons) > 0 {
		for _, tx := range block.Transacitons {
			//过滤非转账交易
			if tx.TransactionType != "Payment" {
				continue
			}

			if txInfo, err := s.parseBlockRawTX(tx, block.Hash, height); err != nil {
				log.Info(tx.Hash, "parseBlockRawTX err：", err.Error())
			} else if txInfo != nil {
				task.TxInfos = append(task.TxInfos, txInfo)
			}
		}
	}
	//tsjon, _ := json.Marshal(task)
	//log.Info(string(tsjon))
	//log.Infof("ScanBlock : %d, txs : %d ,used time : %f 's", height, len(block.Txs), time.Since(starttime).Seconds())
	return task, nil
}

//解析交易
func (s *Scanner) parseBlockRawTX(tx *rpc.Transaction, blockhash string, height int64) (txInfo *TxInfo, err error) {

	if tx == nil || tx.TransactionType != "Payment" || tx.Meta.TransactionResult != "tesSUCCESS" || tx.Flags != 2147483648 {
		return nil, nil
	}
	vinCount := 0
	switch v := tx.Meta.DeliveredAmount.(type) {
	case string:
		vinCount = 1
	case []interface{}: //多种代币,可能包含xrp
		vinCount = len(v)
		return nil, fmt.Errorf("muti coin")
	case map[string]interface{}: //一种代币
		return nil, nil
	default:
		return nil, fmt.Errorf("unkonw account type:%v", tx.Hash)
	}
	fee, _ := strconv.ParseInt(tx.Fee, 10, 64)
	blockTx := &dao.BlockTx{
		Txid:     tx.Hash,
		Height:   height,
		Hash:     blockhash,
		Vincount: vinCount,
		Memo:     tx.DestinationTag,
		From:     tx.Account,
		To:       tx.Destination,
		Type:     tx.TransactionType,
		State:    tx.Meta.TransactionResult,
		Fee:      fee,
	}
	txInfo = &TxInfo{Tx: blockTx}
	//获取合约执行状态
	switch v := tx.Meta.DeliveredAmount.(type) {
	case string:
		value, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		tokentx := &dao.TokenTx{
			Contract: "xrp",
			Txid:     tx.Hash,
			Height:   height,
			Hash:     blockhash,
			Vmstate:  tx.Meta.TransactionResult,
			Index:    0,
			From:     tx.Account,
			To:       tx.Destination,
			Value:    value,
			Memo:     tx.DestinationTag,
			Coinname: "xrp",
		}
		txInfo.Contractxs = append(txInfo.Contractxs, tokentx)

	default:
		return nil, nil
	}
	return txInfo, nil
}
