package biw

import (
	"adasync/common"
	"adasync/common/conf"
	"adasync/common/log"
	dao "adasync/models/po/utxo"
	"adasync/services"
	"errors"
	"fmt"
	"strconv"
	"strings"

	//"adasync/utils"
	rpc "adasync/utils/ada"
	"time"

	"github.com/shopspring/decimal"
)

type Scanner struct {
	*rpc.RpcClient
	conf conf.SyncConfig
}

func NewScanner(conf conf.Config, node conf.NodeConfig, WatchControl *services.WatchControl) common.Scanner {
	return &Scanner{
		RpcClient: rpc.NewRpcClient(node.Url, "", ""),
		conf:      conf.Sync,
	}
}

func (s *Scanner) Rollback(height int64) {
	//删除指定高度之后的数据
	dao.BlockRollBack(height)
	dao.TxRollBack(height)
	dao.TxVoutRollBack(height)
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
	//starttime := time.Now()
retryGetBlockByHeight:
	block, err := s.GetBlockByHeight(height)
	if err != nil {
		log.Warnf("%v height:%v", err.Error(), height)
		time.Sleep(time.Second * 3)
		goto retryGetBlockByHeight
	}

	//log.Infof("GetBlockByHeight : %d, txs : %d ", height, len(block.Txs))

	task := &ProcTask{
		Irreversible: bestHeight-height >= s.conf.Confirmations,
		BestHeight:   bestHeight,
		Block: &dao.BlockInfo{
			Height:            block.BlockIdentifier.Index,
			Hash:              block.BlockIdentifier.Hash,
			Previousblockhash: "",
			Nextblockhash:     "",
			Transactions:      len(block.Transactions),
			Confirmations:     bestHeight - height + 1, //block.Confirmations
			Timestamp:         time.Now(),
			Createtime:        time.Now(),
		},
	}

	//并发处理区块内的交易
	//log.Info(block.BlockIdentifier.Index, len(block.Transactions))
	for _, tx := range block.Transactions {
		if txInfo, err := parseBlockRawTX(s.RpcClient, tx, task.Block.Hash, height); err != nil {
			log.Info(err.Error())
		} else if txInfo != nil {
			task.TxInfos = append(task.TxInfos, txInfo)
		}
	}
	//if txjson, err := json.Marshal(task); err == nil {
	//log.Infof("block:%v", xutils.String(task))
	//} else {
	//	log.Warn(err.Error())
	//}
	return task, nil
}

//解析交易
func parseBlockRawTX(RpcClient *rpc.RpcClient, tx *rpc.Transaction, blockhash string, height int64) (*TxInfo, error) {
	var vouts []*dao.BlockTxVout
	var vins []*dao.BlockTxVin
	if tx == nil {
		return nil, nil
	}

	blocktx := &dao.BlockTx{
		Txid:      tx.TransactionIdentifier.Hash,
		Height:    height,
		Blockhash: blockhash,
		Version:   0,
		//Voutcount:  len(tx.Vout),
		//Vincount:   len(tx.Vin),
		Timestamp:  time.Now(),
		Createtime: time.Now(),
		//Fee:        decimal.New(tx.Fee,-6).String(),
	}
	inAmount := decimal.NewFromInt(0)
	outAmount := decimal.NewFromInt(0)
	for _, op := range tx.Operations {
		if op.Type != "output" {
			continue
		}
		if op.Status != "success" {
			return nil, errors.New("失败的交易")
		}
		value := op.Amount.Value.Shift(0 - op.Amount.Currency.Decimals)
		value = value.Abs()
		if op.Amount.Currency.Symbol == "ADA" {
			outAmount = outAmount.Add(value)
		}
		inputs := strings.Split(op.CoinChange.CoinIdentifier.Identifier, ":")
		//log.Info(op.CoinChange.CoinIdentifier.Identifier,xutils.String(inputs))
		if len(inputs) != 2 {
			inputs = []string{"", "0"}
			//return nil, errors.New("交易解析错误"+op.CoinChange.CoinIdentifier.Identifier)
		}
		VoutN, _ := strconv.Atoi(inputs[1])
		blocktxvout := &dao.BlockTxVout{
			Height:     height,
			Address:    op.Account.Address,
			Txid:       blocktx.Txid,
			VoutN:      VoutN,
			Blockhash:  blocktx.Blockhash,
			Value:      value.String(),
			Timestamp:  blocktx.Timestamp,
			Createtime: time.Now(),
			//AssertId:   op.Amount.Currency.Symbol,
			//AssertName: op.Amount.Currency.Symbol,
		}
		if op.Amount.Currency.Symbol != "ADA" {
			log.Info(tx.TransactionIdentifier.Hash + "交易解析错误")
			panic(tx.TransactionIdentifier.Hash + "交易解析错误")
		}
		if len(op.Metadata.TokenBundle) == 1 && len(op.Metadata.TokenBundle[0].Tokens) == 1 {
			AssertId := fmt.Sprintf("%v-%v-%v", op.Metadata.TokenBundle[0].PolicyId, op.Metadata.TokenBundle[0].Tokens[0].Currency.Symbol, op.Metadata.TokenBundle[0].Tokens[0].Currency.Decials)
			value2 := op.Metadata.TokenBundle[0].Tokens[0].Value.Shift(0 - op.Metadata.TokenBundle[0].Tokens[0].Currency.Decials)
			value2 = value2.Abs()
			blocktxvout.AssertName = ""
			blocktxvout.AssertId = AssertId
			blocktxvout.AssertValue = value2.String()
		}
		vouts = append(vouts, blocktxvout)

	}

	for _, op := range tx.Operations {
		if op.Type != "input" {
			continue
		}
		//log.Info(xutils.String(op))
		if op.Status != "success" {
			return nil, errors.New("失败的交易")
		}
		value := op.Amount.Value.Shift(0 - op.Amount.Currency.Decimals)
		value = value.Abs()
		if op.Amount.Currency.Symbol == "ADA" {
			inAmount = inAmount.Add(value)
		}
		inputs := strings.Split(op.CoinChange.CoinIdentifier.Identifier, ":")
		if len(inputs) != 2 {
			inputs = []string{"", "0"}
			//return nil, errors.New("交易解析错误"+op.CoinChange.CoinIdentifier.Identifier)
		}
		VoutN, _ := strconv.Atoi(inputs[1])
		blocktxvin := &dao.BlockTxVin{
			Blockhash:  "",
			Address:    op.Account.Address,
			Value:      value.String(),
			Timestamp:  time.Now(),
			Createtime: time.Now(),
			Txid:       inputs[0],
			VoutN:      VoutN,
			SpendTxid:  blocktx.Txid,
			//AssertId:   op.Amount.Currency.Symbol,
			//AssertName: op.Amount.Currency.Symbol,
		}
		if op.Amount.Currency.Symbol != "ADA" {
			log.Info(tx.TransactionIdentifier.Hash + "交易解析错误")
			panic(tx.TransactionIdentifier.Hash + "交易解析错误")
		}

		if len(op.Metadata.TokenBundle) == 1 && len(op.Metadata.TokenBundle[0].Tokens) == 1 {

			AssertId := fmt.Sprintf("%v-%v-%v", op.Metadata.TokenBundle[0].PolicyId, op.Metadata.TokenBundle[0].Tokens[0].Currency.Symbol, op.Metadata.TokenBundle[0].Tokens[0].Currency.Decials)
			value2 := op.Metadata.TokenBundle[0].Tokens[0].Value.Shift(0 - op.Metadata.TokenBundle[0].Tokens[0].Currency.Decials)
			value2 = value2.Abs()
			blocktxvin.AssertName = ""
			blocktxvin.AssertId = AssertId
			blocktxvin.AssertValue = value2.String()
			//log.Info(xutils.String(blocktxvin))
		}
		vins = append(vins, blocktxvin)
	}
	blocktx.Fee = inAmount.Sub(outAmount).String()
	//log.Info(xutils.String(vins))
	//log.Info(xutils.String(vouts))
	return &TxInfo{
		Tx:    blocktx,
		Vouts: vouts,
		Vins:  vins,
	}, nil
}
