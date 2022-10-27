package dom

import (
	"domsync/common"
	"domsync/common/conf"
	dao "domsync/models/po/dom"
	"domsync/services"
	"domsync/utils"
	"domsync/utils/dom"
	"fmt"
	"github.com/shopspring/decimal"
	"log"
	"math/big"
	"sync"
	"time"
)

type Scanner struct {
	client *dom.RpcClient
	lock   *sync.Mutex
	conf   conf.SyncConfig
	Watch  *services.WatchControl
}

func NewScanner(conf conf.Config, node conf.NodeConfig, watch *services.WatchControl) common.Scanner {
	//如果启动eth，顺便启动定制加载的合约
	//err := InitEthClient(node.Url)
	//if err != nil {
	//	panic(err)
	//}
	return &Scanner{
		client: dom.NewRpcClient(node.Url),
		lock:   &sync.Mutex{},
		conf:   conf.Sync,
		Watch:  watch,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	_, err := dao.BlockRollBack(height)
	if err != nil {
		panic(err.Error())
	}
	_, err = dao.TxRollBack(height)
	if err != nil {
		panic(err.Error())
	}
}

func (s *Scanner) Init() error {
	if s.conf.EnableRollback {
		s.Rollback(s.conf.RollHeight)
	}
	return nil
}

func (s *Scanner) Clear() {
}

//var i = int64(60612620)

//获取最高区块高度
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	return s.client.BlockNumber()
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
	block, err := s.client.GetBlockByHeight(height, true)
	if err != nil {
		return nil, fmt.Errorf("GetBlockByNumber %d  , err : %v", height, err)
	}

	cnt, err := dao.GetBlockCountByHash(block.Items[0].Block.Txhash)
	if err != nil {
		return nil, fmt.Errorf("database err")
	}

	if cnt > 0 {
		return nil, fmt.Errorf("already have block , count : %d", cnt)
	}

	task := &ProcTask{
		BestHeight: bestHeight,
		Block: &dao.BlockInfo{
			Height:            block.Items[0].Block.Height,
			Hash:              block.Items[0].Block.Txhash,
			Previousblockhash: block.Items[0].Block.Parenthash,
			Timestamp:         time.Unix(block.Items[0].Block.Blocktime, 0),
			Transactions:      len(block.Items[0].Block.Txs),
			Confirmations:     bestHeight - height + 1,
			Createtime:        time.Now(),
		},
	}
	workpool := utils.NewWorkPool(10)
	for _, tx := range block.Items[0].Block.Txs {
		workpool.Incr()
		go func(tx dom.Txs, task *ProcTask) {
			defer workpool.Dec()
			blockTx, err := s.parseBlockTX(tx, bestHeight, block.Items[0].Block.Blocktime, block.Items[0].Block.Height)
			if err == nil {
				task.TxInfos = append(task.TxInfos, blockTx...)
			}
		}(tx, task)
	}
	workpool.Wait()
	_ = starttime
	log.Printf("scanBlock %d ,%d,used time : %f 's", height, len(block.Items[0].Block.Txs), time.Since(starttime).Seconds())
	return task, nil
}

////批量解析交易
//func (s *Scanner) batchParseTx(tx *dom.Txs, bestHeight, blockTimestamp int64, task *DomProcTask, w *sync.WaitGroup, BlockHeight int64) {
//	blockTxs, err := s.parseBlockTX(tx, height, blockTimestamp)
//	if err == nil {
//		s.lock.Lock()
//		defer s.lock.Unlock()
//		task.TxInfos = append(task.TxInfos, blockTxs...)
//	} else {
//		//log.Printf(err.Error())
//	}
//}

func (s *Scanner) isValidTransaction(tx dom.Txs) bool {
	//if tx == nil {
	//	return false
	//}
	txre, err := s.client.GetTransactionByHash(tx.Hash)
	if err != nil {
		return false
	}
	_, f64 := common.StrToFloat(tx.Payload.Transfer.Amount, common.Dec_8base)
	if tx.Execer == "coins" && tx.Payload.Ty == 1 && f64 > 0 && txre.Receipt.Ty == 2 {
		return true
	}
	return false
}

// 解析交易
func (s *Scanner) parseBlockTX(tx dom.Txs, bestHeight, blockTimestamp int64, BlockHeight int64) ([]*dao.BlockTx, error) {
	//if tx == nil {
	//	return nil, fmt.Errorf("tx is null")
	//}
	txHash := tx.Hash
	res := make([]*dao.BlockTx, 0)

	if s.isValidTransaction(tx) {
		//主币transfer的处理逻辑
		tx, err := s.client.GetTransactionByHash(tx.Hash)
		if err != nil {
			log.Printf("高度%d,GetTransactionByHash %s, err %v ", BlockHeight, txHash, err)
			return nil, err
		}

		if tx.Receipt.Ty != 2 {
			log.Printf("高度%d ,%s 无效交易", BlockHeight, tx.Tx.Hash)
			return nil, fmt.Errorf("%s 无效交易", tx.Tx.Hash)
		}

		hash, err := s.client.GetBlockHashByHeight(tx.Height)
		if err != nil {
			log.Printf("高度%d ,GetBlockHashByHeight %s, err %v ", BlockHeight, txHash, err)
			return nil, err
		}

		blocktx := &dao.BlockTx{
			BlockHeight: tx.Height,
			BlockHash:   hash,
			Txid:        tx.Tx.Hash,
			FromAddress: tx.Tx.From,
			Nonce:       tx.Tx.Nonce,
			//GasUsed:     tx.Tx.Fee,
			GasPrice:  tx.Tx.Fee,
			Input:     tx.Tx.Rawpayload,
			CoinName:  s.conf.Name,
			Decimal:   dom.WEI,
			Timestamp: time.Unix(blockTimestamp, 0),
			Amount:    decimal.NewFromBigInt(big.NewInt(tx.Tx.Amount), 0),
			ToAddress: tx.Tx.To,
		}
		log.Printf("高度%d ,解析交易 %s", BlockHeight, txHash)
		res = append(res, blocktx)
	}
	log.Printf("高度%d ,交易 %s不是主币交易", BlockHeight, txHash)
	return res, nil
}
