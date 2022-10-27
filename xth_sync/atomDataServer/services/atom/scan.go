package biw

import (
	"atomDataServer/common"
	"atomDataServer/common/conf"
	"atomDataServer/common/log"
	dao "atomDataServer/models/po/atom"
	"atomDataServer/utils"
	rpc "atomDataServer/utils/atom"
	"fmt"
	"github.com/shopspring/decimal"
	"sync"
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
	if s.conf.EnableRollback {
		s.Rollback(s.conf.RollHeight)
	}
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
		log.Warnf("%v height:%v", err.Error(), height)
		time.Sleep(time.Second * 3)
		goto retryGetBlockByHeight
		//return nil, fmt.Errorf("GetBlockByHeight, err : %v", err)
	}

	//log.Infof("GetBlockByHeight : %v, txs : %v", height,xutils.String(block))
	if has, err := dao.BlockHashExist(block.BlockId.Hash); err != nil {
		return nil, fmt.Errorf("database err")
	} else if has {
		return nil, fmt.Errorf("already have block height: %v, hash: %s , count : %d", block.Block.Header.Height, block.BlockId.Hash, 1)
	}

	task := &ProcTask{
		Irreversible: bestHeight-height >= s.conf.Confirmations,
		BestHeight:   bestHeight,
		Block: &dao.BlockInfo{
			Height:            block.Block.Header.Height.IntPart(),
			Hash:              block.BlockId.Hash,
			Previousblockhash: block.Block.Header.LastBlockID.Hash,
			Nextblockhash:     "",
			Transactions:      len(block.Block.Data.Txs),
			Confirmations:     bestHeight - height + 1, //block.Confirmations
			Timestamp:         block.Block.Header.Time,
			Createtime:        time.Now(),
		},
	}

	//并发处理区块内的交易

	lock := new(sync.Mutex)
	workpool := utils.NewWorkPool(10) //一次性发太多请求会让节点窒息
	for index, rawtx := range block.Block.Data.Txs {
		workpool.Add(1)
		txid := fmt.Sprintf("%X", rawtx.Hash())


		go func(txid string, index int) {
			defer workpool.Dec()
			//log.Info(txid)
		getTransaction:
			tx, err := s.GetRawTransaction(txid)
			if err != nil {
				log.Warn(err.Error(), txid)
				goto getTransaction
			}
			//rawtx2,_ := json.Marshal(tx)
			//log.Info(string(rawtx2))
			//过滤失败交易
			if tx.Type != "send" || !tx.Success {
				return
			}
			//log.Info(txid,block.Block.Header.Height,rawtx)
			//log.Info("getrawtransaction success:", txid, index, len(block.Txs))
			if txInfo, err := parseBlockRawTX(s.RpcClient, tx, block.BlockId.Hash, height); err != nil {
				log.Info(err.Error())
			} else if txInfo != nil {
				lock.Lock() //append并发不安全
				defer lock.Unlock()
				task.TxInfos = append(task.TxInfos, txInfo)
			}
			return
		}(txid, index)
	}
	workpool.Wait()
	return task, nil
}

//解析交易
func parseBlockRawTX(RpcClient *rpc.RpcClient, tx *rpc.Transaction, blockhash string, height int64) (*dao.BlockTx, error) {
	if tx == nil {
		return nil, nil
	}

	blocktx := &dao.BlockTx{
		BlockHash:   blockhash,
		Txid:        tx.Hash,
		From:        tx.From,
		To:          tx.To,
		Value:       decimal.NewFromInt(tx.Value).Shift(-6).String(),
		BlockHeight: tx.BlockHeight,
		Type:        tx.Type,
		Gasused:     tx.GasUsed,
		Gaswanted:   tx.GasWanted,
		Fee:         decimal.NewFromInt(tx.Fee).Shift(-6).String(),
		Rawlogs:     tx.RawLogs,
		Timestamp:   tx.Timestamp,
		Createtime:  time.Now(),
		Memo:        tx.Memo,
	}
	return blocktx, nil
}
