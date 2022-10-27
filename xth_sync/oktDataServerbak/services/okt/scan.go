package atp

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"oktDataServer/common"
	"oktDataServer/common/conf"
	"oktDataServer/common/log"
	dao "oktDataServer/models/po/cfx"
	"oktDataServer/services"
	rpc "oktDataServer/utils/okt"
	"strings"
	"time"

	"github.com/okex/exchain-go-sdk/types"
	chaintypes "github.com/okex/exchain/x/token/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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
			Height:            int64(height),
			Blockhash:         fmt.Sprintf("%v", block.LastBlockID.Hash),
			Previousblockhash: "",
			Transactions:      len(block.Txs),
			Confirmations:     bestHeight - height + 1, //block.Confirmations
			Timestamp:         time.Now(),
			Createtime:        time.Now(),
		},
	}

	//并发处理区块内的交易
QueryBlockResults:
	blockresult, err := s.Tendermint().QueryBlockResults(height)
	if err != nil {
		log.Info("%v height:%v", err.Error(), height)
		//time.Sleep(time.Second * 3)
		goto QueryBlockResults
	}
	//lock := new(sync.Mutex)
	for index, tx := range block.Txs {
		if blockresult.TxsResults[index].Code != 0 || len(blockresult.TxsResults[index].Events) == 0 {
			log.Info("失败交易:" + strings.ToUpper(hex.EncodeToString(tx.Hash())))
			continue
		}
		stdtx := new(authtypes.StdTx)
		log.Info(strings.ToUpper(hex.EncodeToString(tx.Hash())))
		if err = s.Client.Token().(types.BaseClient).GetCodec().UnmarshalBinaryLengthPrefixed([]byte(tx), stdtx); err != nil {
			log.Error(err.Error())
			panic(err.Error())
			continue

		}

		//log.Info(String(stdtx))
		if txInfo, err := parseBlockRawTX(stdtx, hex.EncodeToString(tx.Hash()), task.Block.Height, task.Block.Blockhash, bestHeight); err != nil {
			//log.Info(err.Error())
		} else if txInfo != nil {
			//lock.Lock() //append并发不安全
			task.TxInfos = append(task.TxInfos, txInfo)

			//if status, err := s.GetTxFromScan(strings.ToUpper(hex.EncodeToString(tx.Hash()))); err != nil || status != "SUCCESS" {
			//	panic(strings.ToUpper(hex.EncodeToString(tx.Hash())) + "  " + status)
			//} else {
			//	//println(status)
			//}
			//lock.Unlock()
		}
	}
	//ts, _ := json.Marshal(task)
	//log.Info("task", len(block.Txs), string(ts))
	return task, nil
}

//解析交易
func parseBlockRawTX(tx *authtypes.StdTx, txhash string, blockheight int64, blockhash string, bestheight int64) (*dao.BlockTx, error) {
	txhash = strings.ToUpper(txhash)
	if tx.GetFee().String() == "" {
		return nil, errors.New("faild tx")
	}

	blocktx := &dao.BlockTx{
		Txid:          txhash,
		CoinName:      conf.Cfg.Sync.Name,
		Contract:      "",
		FromAddress:   "",
		ToAddress:     "",
		BlockHeight:   blockheight,
		BlockHash:     blockhash,
		Memo:          tx.Memo,
		Amount:        "",
		Status:        1,
		Fee:           tx.Fee.Amount[0].Amount.String(),
		Timestamp:     time.Now(),
		CreateTime:    time.Now(),
		Confirmations: bestheight - blockheight + 1,
	}
	for _, msg := range tx.Msgs {
		if msg.Type() != "send" {
			continue
		}
		sendmsg, ok := msg.(chaintypes.MsgSend)
		if !ok {
			panic("逻辑出错，需要处理")
		}
		for _, amount := range sendmsg.Amount {
			if amount.Denom != "okt" {
				continue
			}
			blocktx.FromAddress = sendmsg.FromAddress.String()
			blocktx.ToAddress = sendmsg.ToAddress.String()
			blocktx.Amount = amount.Amount.String()
			return blocktx, nil
		}
	}
	return nil, nil
}
func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
