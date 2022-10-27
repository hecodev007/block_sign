package nyzo

import (
	"stellarsync/common"
	"stellarsync/common/conf"
	"stellarsync/common/log"
	dao "stellarsync/models/po/nyzo"
	"stellarsync/services"
	rpc "stellarsync/utils/stellar"
	"time"
)

type Scanner struct {
	*rpc.RpcClient
	conf conf.SyncConfig
}

func NewScanner(conf conf.Config, node conf.NodeConfig,watch *services.WatchControl) common.Scanner {
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

//var i = int64(1417694)

//获取最高区块高度
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	//i++
	//return i, nil
	count, err := s.GetBlockCount() //获取到的是区块个数
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
retryGetBlockByHeight:
	block, err := s.GetBlockByHeight(height)
	if err != nil {
		log.Warnf("%v height: %v", err.Error(), height)
		time.Sleep(time.Second * 3)
		goto retryGetBlockByHeight
		//return nil, fmt.Errorf("GetBlockByHeight, err : %v", err)
	}


		////log.Infof("GetBlockByHeight : %d, txs : %d ", height, len(block.Txs))
	//if has, err := dao.BlockHashExist(block.Hash); err != nil {
	//	return nil, fmt.Errorf("database err")
	//} else if has {
	//	return nil, fmt.Errorf("already have block height: %v, hash: %s , count : %d", block.Block.Header.Height, block.BlockId.Hash, 1)
	//}

	task := &ProcTask{
		Irreversible: bestHeight-height >= s.conf.Confirmations,
		BestHeight:   bestHeight,
		Block: &dao.BlockInfo{
			Height:            block.Height,
			Hash:              block.Hash,
			Previousblockhash: "",
			Nextblockhash:     "",
			Transactions:      len(block.Transactions),
			Confirmations:     bestHeight - height + 1, //block.Confirmations
			Timestamp:         time.Now(),
			Createtime:        time.Now(),
		},
	}

	//并发处理区块内的交易

	for _, tx := range block.Transactions {
		if block.Height == 11385425{
			log.Info(task.Block.Transactions,tx.From)
		}

		if txInfo, err := parseBlockRawTX(s.RpcClient, tx, block.Hash, height); err != nil {
			//log.Info(err.Error())
		} else if txInfo != nil {
			task.TxInfos = append(task.TxInfos, txInfo)
		}
	}
	return task, nil
}

//解析交易
func parseBlockRawTX(RpcClient *rpc.RpcClient, tx *rpc.Transaction, blockhash string, height int64) (*dao.BlockTx, error) {
	if tx == nil  {
		return nil, nil
	}
	blocktx := &dao.BlockTx{
		BlockHash:   blockhash,
		Txid:        tx.Txid,
		From:        tx.From,
		To:          tx.To,
		Value:      tx.Amount,
		BlockHeight: height,
		Fee:         tx.Fee,
		Timestamp:   time.Now(),
		Createtime:  time.Now(),
		Memo:        tx.Memo,
	}
	return blocktx, nil
}
