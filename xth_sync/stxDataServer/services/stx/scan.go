package stx

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"stxDataServer/common"
	"stxDataServer/common/conf"
	"stxDataServer/common/log"
	dao "stxDataServer/models/po/stx"
	"stxDataServer/services"
	rpc "stxDataServer/utils/stx"
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

var i = int64(2361)

//获取最高区块高度
func (s *Scanner) GetBestBlockHeight() (int64, error) {
	i++
	return i, nil
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
	block, err := s.GetBlockByHeight(bestHeight, height)
	if err != nil {
		log.Infof("%v height:%v", err.Error(), height)
		//time.Sleep(time.Second * 3)
		goto retryGetBlockByHeight
	}

	task := &ProcTask{
		Irreversible: bestHeight-height >= s.conf.Confirmations,
		BestHeight:   bestHeight,
		Block: &dao.BlockInfo{
			Height:            height,
			Blockhash:         block.Hash,
			Previousblockhash: "",
			Transactions:      len(block.Txs),
			Confirmations:     bestHeight - height + 1, //block.Confirmations
			Createtime:        time.Now(),
		},
	}

	for _, txid := range block.Txs {
	GetTransaction:
		tx, err := s.RpcClient.Getransaction(txid)
		if err != nil {
			time.Sleep(time.Second * 5)
			goto GetTransaction
		}
		//log.Info(String(stdtx))
		if txInfo, err := parseBlockRawTX(tx, bestHeight); err != nil {
			//log.Info(err.Error())
		} else if txInfo != nil {
			//lock.Lock() //append并发不安全
			task.TxInfos = append(task.TxInfos, txInfo)
		}
	}
	//ts, _ := json.Marshal(task)
	//log.Info("task", len(block.Txs), string(ts))
	return task, nil
}

//解析交易
func parseBlockRawTX(tx *rpc.Transaction, bestheight int64) (*dao.BlockTx, error) {
	if tx.TxStatus != "success" {
		return nil, errors.New("失败的交易")
	}

	if tx.TxType != "token_transfer" {
		return nil, errors.New("不是转账交易类型")
	}
	blocktx := &dao.BlockTx{

		Txid:          tx.TxId,
		CoinName:      conf.Cfg.Sync.Name,
		FromAddress:   tx.SenderAddress,
		ToAddress:     tx.TokenTransfer.RecipientAddress,
		BlockHeight:   tx.BlockHeight,
		BlockHash:     tx.BlockHash,
		Memo:          "",
		Amount:        tx.TokenTransfer.Amount.Shift(-6).String(),
		Status:        1,
		Fee:           tx.FeeRate.Shift(-6).String(),
		CreateTime:    time.Now(),
		Confirmations: bestheight - tx.BlockHeight + 1,
	}
	memo := tx.TokenTransfer.Memo
	memo = strings.TrimPrefix(memo, "0x")
	memoBytes, _ := hex.DecodeString(memo)
	for i := 0; i < len(memoBytes); i++ {
		if memoBytes[i] == 0 {
			memoBytes = memoBytes[0:i]
		}
	}
	memo = string(memoBytes)
	blocktx.Memo = memo
	return blocktx, nil
}
func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
