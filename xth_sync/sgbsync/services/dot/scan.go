package dot

import (
	"encoding/json"
	"sgbsync/common"
	"sgbsync/common/conf"
	"sgbsync/common/log"
	dao "sgbsync/models/po/dot"
	"sgbsync/services"
	rpc "sgbsync/utils/dot"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type Scanner struct {
	*rpc.RpcClient
	lock  *sync.Mutex
	conf  conf.SyncConfig
	watch *services.WatchControl
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	return &Scanner{
		RpcClient: rpc.NewRpcClient(node.Url, node.Node, node.RPCSecret),
		lock:      &sync.Mutex{},
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
	if conf.Cfg.Sync.EnableRollback {
		s.Rollback(conf.Cfg.Sync.RollHeight)
	}
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

		task := &ProcTask{
			BestHeight: bestHeight,
			Block: &dao.BlockInfo{
				Height:           height,
				Hash:              "timeout",
				Previousblockhash: "",
				Transactions:      0,
				Confirmations:     bestHeight - height+ 1,
				Time:              time.Now(),
			},
		}
		log.Info(height,err.Error())
		return task,nil
		//log.Info(err.Error())
		//return nil, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}

	task := &ProcTask{
		BestHeight: bestHeight,
		Block: &dao.BlockInfo{
			Height:            block.Number.IntPart(),
			Hash:              block.Hash,
			Previousblockhash: block.ParentHash,
			Transactions:      len(block.Extrinsics),
			Confirmations:     bestHeight - block.Number.IntPart() + 1,
			Time:              time.Now(),
		},
	}

	if task.Block.Confirmations >= s.conf.Confirmations {
		task.irreversible = true
	}
	rawtxs, err := s.RpcClient.GetExtrinsicsByNum(int64(height))
	if err != nil {
		log.Info("GetExtrinsicsByNum", err.Error())
		return nil, err
	}
	s.batchParseTx(block.Extrinsics, rawtxs, task)
	_ = starttime
	//log.Infof("scanBlock %d ,used time : %f
	return task, nil
}

//批量解析交易
func (s *Scanner) batchParseTx(txs []*rpc.Extrinsic, rawtxs [][]byte, task *ProcTask) {

	blockTxs, err := s.parseBlockTX(txs, rawtxs, task.Block)
	if err == nil {
		s.lock.Lock()
		defer s.lock.Unlock()
		task.TxInfos = append(task.TxInfos, blockTxs...)
	} else {
		log.Info(err.Error())
	}
}

// 解析交易
func (s *Scanner) parseBlockTX(txs []*rpc.Extrinsic, rawtxs [][]byte, block *dao.BlockInfo) (btx []*dao.BlockTx, err error) {
	blockTxs := make([]*dao.BlockTx, 0)
	if len(txs) == 0 {
		return nil, nil
	}

	for k, tx := range txs {
		if tx.Success == false {
			continue
		}
		if tx.Method.Pallet == "balances" && (tx.Method.Method == "transferKeepAlive" || tx.Method.Method == "transfer") {
			tmpBlockTx := new(dao.BlockTx)
			tmpBlockTx.Fromaccount = tx.Signature.Signer.Id
			tmpBlockTx.Toaccount = tx.Args.Dest.(map[string]interface{})["id"].(string)
			tmpBlockTx.Amount = tx.Args.Value.Shift(-10).String()
			tmpBlockTx.Hash = block.Hash
			tmpBlockTx.Txid = tx.Hash
			//tmpBlockTx.SysFee = tx.Info.PartialFee.Shift(-10).String()
			//if s.watch.IsWatchAddressExist(tmpBlockTx.Fromaccount) || s.watch.IsWatchAddressExist(tmpBlockTx.Toaccount) {
			tmpBlockTx.SysFee = decimal.NewFromInt(int64(125000000 + len(rawtxs[k]))).Shift(-10).String()
			//} else {
			//	tmpBlockTx.SysFee = "0.0125"
			//}
			tmpBlockTx.Height = block.Height
			//tmpBlockTx.Succuss = 1
			blockTxs = append(blockTxs, tmpBlockTx)
			continue
		}
		//log.Info(tx.Hash,tx.Method.Pallet,tx.Method.Method)

		if tx.Method.Pallet == "Utility" && tx.Method.Method == "batch" {
			//log.Info(tx.Hash)
			for _, call := range tx.Args.Calls {
				if call.Method.Pallet == "balances" && (call.Method.Method == "transferKeepAlive" || call.Method.Method == "transfer") {
					tmpBlockTx := new(dao.BlockTx)
					tmpBlockTx.Fromaccount = tx.Signature.Signer.Id
					tmpBlockTx.Toaccount = call.Args.Dest.Id
					tmpBlockTx.Amount = call.Args.Value.Shift(-10).String()
					tmpBlockTx.Hash = block.Hash
					tmpBlockTx.Txid = tx.Hash
					//tmpBlockTx.SysFee = tx.Info.PartialFee.Shift(-10).String()
					//tmpBlockTx.SysFee = sgb.GetFee(tx.Hash)
					//if s.watch.IsWatchAddressExist(tmpBlockTx.Fromaccount) || s.watch.IsWatchAddressExist(tmpBlockTx.Toaccount) {
					tmpBlockTx.SysFee = decimal.NewFromInt(int64(125000000 + len(rawtxs[k]))).Shift(-10).String()
					//} else {
					//	tmpBlockTx.SysFee = "0.0125"
					//}

					tmpBlockTx.Height = block.Height
					//tmpBlockTx.Succuss = 1
					blockTxs = append(blockTxs, tmpBlockTx)
					continue
				}
			}
		}

	}
	//log.Info(block.Height,String(blockTxs))
	return blockTxs, nil
}

// 解析交易
func parseBlockTX2(txs []*rpc.Extrinsic, rawtxs [][]byte, block *dao.BlockInfo) ([]*dao.BlockTx, error) {
	blockTxs := make([]*dao.BlockTx, 0)
	if len(txs) == 0 {
		return nil, nil
	}

	for k, tx := range txs {
		if tx.Success == false {
			continue
		}
		if tx.Method.Pallet == "balances" && (tx.Method.Method == "transferKeepAlive" || tx.Method.Method == "transfer") {
			tmpBlockTx := new(dao.BlockTx)
			tmpBlockTx.Fromaccount = tx.Signature.Signer.Id
			tmpBlockTx.Toaccount = tx.Args.Dest.(map[string]interface{})["id"].(string)
			tmpBlockTx.Amount = tx.Args.Value.Shift(-10).String()
			tmpBlockTx.Hash = block.Hash
			tmpBlockTx.Txid = tx.Hash
			tmpBlockTx.SysFee = decimal.NewFromInt(int64(125000000 + len(rawtxs[k]))).Shift(-10).String()
			tmpBlockTx.Height = block.Height
			//tmpBlockTx.Succuss = 1
			blockTxs = append(blockTxs, tmpBlockTx)
			continue
		}
		//log.Info(tx.Hash,tx.Method.Pallet,tx.Method.Method)

		if tx.Method.Pallet == "utility" && tx.Method.Method == "batch" {
			//log.Info(tx.Hash)
			for _, call := range tx.Args.Calls {
				if call.Method.Pallet == "balances" && (call.Method.Method == "transferKeepAlive" || call.Method.Method == "transfer") {
					tmpBlockTx := new(dao.BlockTx)
					tmpBlockTx.Fromaccount = tx.Signature.Signer.Id
					tmpBlockTx.Toaccount = call.Args.Dest.Id
					tmpBlockTx.Amount = call.Args.Value.Shift(-10).String()
					tmpBlockTx.Hash = block.Hash
					tmpBlockTx.Txid = tx.Hash
					tmpBlockTx.SysFee = decimal.NewFromInt(int64(125000000 + len(rawtxs[k]))).Shift(-10).String()
					tmpBlockTx.Height = block.Height
					//tmpBlockTx.Succuss = 1
					blockTxs = append(blockTxs, tmpBlockTx)
					continue
				}
			}
		}

	}
	//log.Info(block.Height,String(blockTxs))
	return blockTxs, nil
}

func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
