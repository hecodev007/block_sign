package crust

import (
	"crustDataServer/common"
	"crustDataServer/common/conf"
	"crustDataServer/common/log"
	dao "crustDataServer/models/po/crust"
	rpc "crustDataServer/utils/crust"
	"fmt"
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
}

func (s *Scanner) Init() error {
	return nil
}

func (s *Scanner) Clear() {
}

//var i = int64(3395436)

//获取最高区块高度
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i, nil
	return s.BlockHeight() //获取到的是区块个数

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
		log.Warnf("%v height:%v", err.Error(), height)
		time.Sleep(time.Second * 3)
		goto retryGetBlockByHeight
		//return nil, fmt.Errorf("GetBlockByHeight, err : %v", err)
	}

	//log.Infof("GetBlockByHeight : %d, txs : %d ", height, len(block.Block.Extrinsics))
	//log.Infof("%+v", block.Block)
	if has, err := dao.BlockHashExist(block.Hash); err != nil {
		return nil, fmt.Errorf("database err")
	} else if has {
		return nil, fmt.Errorf("already have block height: %d, hash: %s , count : %d", block.Block.Header.Height, block.Hash, 1)
	}

	task := &ProcTask{
		Irreversible: bestHeight-height >= s.conf.Confirmations,
		BestHeight:   bestHeight,
		Block: &dao.BlockInfo{
			Height:            block.Block.Header.Height,
			Hash:              block.Hash,
			Previousblockhash: block.Block.Header.ParentHash,
			Nextblockhash:     "",
			Transactions:      len(block.Block.Extrinsics),
			Confirmations:     bestHeight - height + 1, //block.Confirmations
			Time:              time.Now(),
		},
	}
	meta, err := s.RpcClient.GetMetadata()
	if err != nil {
		log.Warn(err.Error())
		return nil, err
	}
	for i, rawtx := range block.Block.Extrinsics {
		tx, err := rpc.HexToTransaction(&meta.Metadata, rawtx)
		if err != nil {
			log.Warn("block:", height, "tx.index:", i, "err:", err.Error())
			continue
		}
		if tx.Function != "transfer" {
			continue
		}
		tmptx, _ := parseBlockRawTX(s.RpcClient, tx, block.Hash, block.Block.Header.Height)
		task.TxInfos = append(task.TxInfos, tmptx)
	}

	//if txjson, err := json.Marshal(task); err == nil {
	//	log.Infof("block:%v", string(txjson))
	//} else {
	//	log.Warn(err.Error())
	//}
	return task, nil
}

//解析交易
func parseBlockRawTX(RpcClient *rpc.RpcClient, tx *rpc.Transaction, blockhash string, height int64) (*dao.BlockTx, error) {
	blocktx := &dao.BlockTx{
		Txid:        tx.Txid,
		Height:      height,
		Hash:        blockhash,
		Fee:         tx.Fee.String(),
		Fromaccount: tx.From,
		Toaccount:   tx.To,
		Amount:      tx.Value.String(),
	}
	return blocktx, nil
}
