package biw

import (
	"fmt"
	"lunasync/common"
	"lunasync/common/conf"
	"lunasync/common/log"
	dao "lunasync/models/po/atom"
	"lunasync/services"
	"lunasync/utils"
	"strings"

	"github.com/terra-money/core/app"

	rpc "lunasync/utils/luna"
	"sync"
	"time"

	"github.com/shopspring/decimal"
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
	log.Info(dao.BlockRollBack(height))
	log.Info(dao.TxRollBack(height))
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
	//log.Info("scanBlock", height)
retryGetBlockByHeight:
	block, err := s.GetBlockByHeight(height)
	if err != nil {
		log.Infof("%v height:%v", err.Error(), height)
		time.Sleep(time.Second * 3)
		goto retryGetBlockByHeight
		//return nil, fmt.Errorf("GetBlockByHeight, err : %v", err)
	}

	//log.Infof("GetBlockByHeight : %d, txs : %d ", height, len(block.Block.Data.Txs))
	//if has, err := dao.BlockHashExist(block.BlockId.Hash); err != nil {
	//	return nil, fmt.Errorf("database err")
	//} else if has {
	//	return nil, fmt.Errorf("already have block height: %v, hash: %s , count : %d", block.Block.Header.Height, block.BlockId.Hash, 1)
	//}

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
	workpool := utils.NewWorkPool(1) //一次性发太多请求会让节点窒息
	for index, rawtx := range block.Block.Data.Txs {
		txid := fmt.Sprintf("%X", rawtx.Hash())
		//if txid != "8F3E15617BD5B77D588D4CF4C722B3EC6F6A12F0FCF993DCABCD9128C2FC8EB7" {
		//	continue
		//}
		//log.Info(txid)
		encodingConfig := app.MakeEncodingConfig()
		tx, err := encodingConfig.TxConfig.TxDecoder()(rawtx)
		//codec := app.MakeCodec()
		//tx, err := auth.DefaultTxDecoder(codec)(rawtx)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		msgs := tx.GetMsgs()
		//var hasSend bool
		//for _, v := range msgs {
		//	if _, ok := v.(*btypes.MsgSend); ok {
		//		hasSend = true
		//	}
		//
		//	log.Info(v.String())
		//}
		////log.Info(hasSend)
		//if !hasSend {
		//	continue
		//}
		hasSend := false
		for wachaddr, _ := range s.watch.WatchAddrs {
			for _, v := range msgs {
				msgstr := func() (str string) {
					defer func() {
						err := recover()
						if err != nil {
							//log.Error(txid, err)
							str = ""
						}
					}()
					str = v.String()
					return str
				}()
				//if msgstr == "" {
				//	panic("")
				//	hasSend = true
				//}
				//log.Info(msgstr)
				//log.Info(wachaddr)
				if strings.Contains(strings.ToLower(msgstr), wachaddr) {
					hasSend = true
				}
			}
		}
		if !hasSend {
			continue
		}
		workpool.Add(1)

		go func(txid string, index int) {
			defer workpool.Dec()
			//log.Info("getTransaction" + txid)
			//getTransaction:
			txs, err := s.GetRawTransaction(txid)
			if err != nil {
				//log.Warn(err.Error(), txid)
				//goto getTransaction
				return
			}
			//过滤失败交易
			//if tx.Type != "send" || !tx.Success {
			//	return
			//}
			//log.Info(txid, block.Block.Header.Height, rawtx)
			//log.Info("getrawtransaction success:", txid, index, len(block.Txs))
			if txInfos, err := parseBlockRawTX(s.watch, txs, block.BlockId.Hash, height); err != nil {
				log.Info(err.Error())
			} else if len(txInfos) != 0 {
				lock.Lock() //append并发不安全
				defer lock.Unlock()
				task.TxInfos = append(task.TxInfos, txInfos...)
			}
			return
		}(txid, index)
	}
	workpool.Wait()
	return task, nil
}

//解析交易

func parseBlockRawTX(watch *services.WatchControl, txs []*rpc.Transaction, blockhash string, height int64) (ret []*dao.BlockTx, err error) {
	if len(txs) == 0 {
		return nil, nil
	}

	for _, tx := range txs {
		deciml := 6
		if tx.Token != "uluna" {
			contractinfo, err := watch.GetContract(tx.Token)
			if err != nil {
				continue
			}
			deciml = contractinfo.Decimal
		}
		blocktx := &dao.BlockTx{
			BlockHash:   blockhash,
			Txid:        tx.Hash,
			From:        tx.From,
			To:          tx.To,
			Token:       tx.Token,
			Value:       decimal.NewFromInt(tx.Value).Shift(0 - int32(deciml)).String(),
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
		ret = append(ret, blocktx)
	}
	return ret, nil
}
