package atp

import (
	"algoDataServer/common"
	"algoDataServer/common/conf"
	"algoDataServer/common/log"
	dao "algoDataServer/models/po/cfx"
	"algoDataServer/services"
	rpc "algoDataServer/utils/algo"
	"errors"
	"strconv"
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
		//time.Sleep(time.Second * 3)
		goto retryGetBlockByHeight
	}

	task := &ProcTask{
		Irreversible: bestHeight-height >= s.conf.Confirmations,
		BestHeight:   bestHeight,
		Block: &dao.BlockInfo{
			Height:            int64(block.Round),
			Blcokhash:         block.Hash,
			Previousblockhash: "",
			Transactions:      len(block.Transactions.Transactions),
			Confirmations:     bestHeight - height + 1, //block.Confirmations
			Timestamp:         time.Now(),
			Createtime:        time.Now(),
		},
	}

	//并发处理区块内的交易

	//lock := new(sync.Mutex)
	for _, tx := range block.Transactions.Transactions {
		if tx.Type != "pay" && tx.Type != "axfer" {
			continue
		}

		if txInfo, err := parseBlockRawTX(s.RpcClient, s.watch, &tx, block.Hash, bestHeight); err != nil {
			//log.Info(err.Error())
		} else if txInfo != nil {
			//lock.Lock() //append并发不安全
			task.TxInfos = append(task.TxInfos, txInfo)
			//lock.Unlock()
		}
	}
	//ts, _ := json.Marshal(task)
	//log.Info("task", block.Height.ToInt().Int64(), string(ts), "task")
	return task, nil
}

//解析交易
func parseBlockRawTX(RpcClient *rpc.RpcClient, watch *services.WatchControl, tx *rpc.Transaction, blockhash string, bestheight int64) (*dao.BlockTx, error) {
	if tx.Type != "pay" && tx.Type != "axfer" {
		return nil, errors.New("error type")
	}

	if tx.Type == "pay" {
		blocktx := &dao.BlockTx{
			Txid:          tx.TxID,
			CoinName:      conf.Cfg.Sync.Name,
			Contract:      "",
			FromAddress:   tx.From,
			ToAddress:     tx.Payment.To,
			BlockHeight:   int64(tx.ConfirmedRound),
			BlockHash:     blockhash,
			Amount:        decimal.NewFromInt(int64(tx.Payment.Amount)).Shift(-6).String(),
			Status:        1,
			Fee:           decimal.NewFromInt(int64(tx.Fee)).Shift(-6).String(),
			Timestamp:     time.Now(),
			CreateTime:    time.Now(),
			Confirmations: bestheight - int64(tx.ConfirmedRound) + 1,
		}
		return blocktx, nil
	} else {
		contract := strconv.FormatUint(tx.AssetTransfer.AssetID, 10)
		if !watch.IsContractExist(contract) {
			return nil, errors.New("assert not exist")
		}
		coiname, ding, err := watch.GetContractNameAndDecimal(contract)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		blocktx := &dao.BlockTx{
			Txid:          tx.TxID,
			CoinName:      coiname,
			Contract:      strconv.FormatUint(tx.AssetTransfer.AssetID, 10),
			FromAddress:   tx.From,
			ToAddress:     tx.AssetTransfer.Receiver,
			BlockHeight:   int64(tx.ConfirmedRound),
			BlockHash:     blockhash,
			Amount:        decimal.NewFromInt(int64(tx.AssetTransfer.Amount)).Shift(0 - int32(ding)).String(),
			Status:        1,
			Fee:           decimal.NewFromInt(int64(tx.Fee)).Shift(-6).String(),
			Timestamp:     time.Now(),
			CreateTime:    time.Now(),
			Confirmations: bestheight - int64(tx.ConfirmedRound) + 1,
		}
		return blocktx, nil
	}
	return nil, nil
}
