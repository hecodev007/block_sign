package cfx

import (
	"atpDataServer/common"
	"atpDataServer/common/conf"
	"atpDataServer/common/log"
	dao "atpDataServer/models/po/cfx"
	"atpDataServer/services"
	rpc "atpDataServer/utils/cfx"
	"encoding/json"
	"errors"
	"github.com/shopspring/decimal"
	"math/big"
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
		//return nil, fmt.Errorf("GetBlockByHeight, err : %v", err)
	}
	for _, blockHash := range block.RefereeHashes {
	GetBlockByHash:
		tempblock, err := s.GetBlockByHash(string(blockHash))
		if err != nil {
			log.Info(err.Error())
			goto GetBlockByHash
		}
		block.Transactions = append(block.Transactions, tempblock.Transactions...)
	}
	//log.Info(height, bestHeight, len(block.Transactions), block.Hash)

	//log.Infof("GetBlockByHeight : %d, txs : %d ", height, len(block.Transactions))
	//bs, _ := json.Marshal(block)
	//log.Info(string(bs))
	//if has, err := dao.BlockHashExist(string(block.Hash)); err != nil {
	//	return nil, fmt.Errorf("database err")
	//} else if has {
	//	return nil, fmt.Errorf("already have block height: %d, hash: %s , count : %d", block.Height, block.Hash, 1)
	//}

	task := &ProcTask{
		Irreversible: bestHeight-height >= s.conf.Confirmations,
		BestHeight:   bestHeight,
		Block: &dao.BlockInfo{
			Height:            block.Height.ToInt().Int64(),
			Hash:              string(block.Hash),
			Previousblockhash: string(block.ParentHash),
			Nextblockhash:     "",
			Transactions:      len(block.Transactions),
			Confirmations:     bestHeight - height + 1, //block.Confirmations
			Timestamp:         time.Now(),
			Createtime:        time.Now(),
		},
	}

	//并发处理区块内的交易

	lock := new(sync.Mutex)
	tmptxs := make(map[string]bool)
	for _, tx := range block.Transactions {
		if tmptxs[string(tx.Hash)] == true {
			continue
		}
		tmptxs[string(tx.Hash)] = true
		//log.Info(tx.Hash)
		if tx.To == nil {
			continue
		}

		if txInfo, err := parseBlockRawTX(s.RpcClient, s.watch, &tx, string(block.Hash), block.Height.ToInt().Int64()); err != nil {
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

//解析交易
func parseBlockRawTX(RpcClient *rpc.RpcClient, watch *services.WatchControl, tx *rpc.Transaction, blockhash string, blockheight int64) (*dao.BlockTx, error) {
	//log.Info(tx.Hash)
	//RpcClient.IsUser(tx.To.String())
	if RpcClient.IsUser(tx.To.String()) {
		blocktx := &dao.BlockTx{
			Txid:            string(tx.Hash),
			CoinName:        conf.Cfg.Sync.Name,
			ContractAddress: "",
			FromAddress:     tx.From.String(),
			ToAddress:       tx.To.String(),
			BlockHeight:     blockheight,
			BlockHash:       blockhash,
			Amount:          decimal.NewFromBigInt(tx.Value.ToInt(), -18).String(),
			Status:          0,
			GasPrice:        tx.GasPrice.ToInt().Int64(),
			GasUsed:         21000,
			Fee:             decimal.NewFromBigInt(tx.GasPrice.ToInt(), -18).Mul(decimal.NewFromInt(21000)).String(),
			Nonce:           tx.Nonce.ToInt().Int64(),
			Input:           tx.Data,
			Logs:            "",
			Timestamp:       time.Now(),
			CreateTime:      time.Now(),
		}
		if blocktx.Status != 0 {
			return nil, errors.New("tx.status != success")
		}
		return blocktx, nil
	} else if RpcClient.IsContract(tx.To.String()) && RpcClient.IsTransfer(tx.Data) && watch.IsContractExist(tx.To.String()) {

		to, amount, err := RpcClient.ParseTransferData(tx.Data)
		if err != nil {
			return nil, err
		}
		receipt, err := RpcClient.GetTransactionReceipt(string(tx.Hash))
		if err != nil {
			return nil, err
		}
		if len(receipt.Logs) == 0 {
			return nil, errors.New("tx.status != success")
		}
		coinname, decm, err := watch.GetContractNameAndDecimal(tx.To.String())
		if err != nil {
			return nil, err
		}
		logs, _ := json.Marshal(receipt.Logs)
		blocktx := &dao.BlockTx{
			Txid:            string(tx.Hash),
			CoinName:        coinname,
			ContractAddress: tx.To.String(),
			FromAddress:     tx.From.String(),
			ToAddress:       to,
			BlockHeight:     blockheight,
			BlockHash:       blockhash,
			Amount:          decimal.NewFromBigInt(amount, 0-int32(decm)).String(),
			Status:          0,
			GasPrice:        tx.GasPrice.ToInt().Int64(),
			GasUsed:         receipt.GasUsed.ToInt().Int64(),
			Fee:             decimal.NewFromBigInt((*big.Int)(receipt.GasFee), -18).String(),
			Nonce:           tx.Nonce.ToInt().Int64(),
			Input:           tx.Data,
			Logs:            string(logs),
			Timestamp:       time.Now(),
			CreateTime:      time.Now(),
		}
		return blocktx, nil
	}
	if !watch.IsContractExist(tx.To.String()) {
		return nil, errors.New("合约地址不存在")
	}
	return nil, nil

}
