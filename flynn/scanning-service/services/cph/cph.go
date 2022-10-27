package cph

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/group-coldwallet/scanning-service/common"
	"github.com/group-coldwallet/scanning-service/conf"
	"github.com/group-coldwallet/scanning-service/models/cph"
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
	"strings"
)

type CphService struct {
	cfg     conf.Config
	nodeCfg conf.NodeConfig
	client  *common.Client
}

func (cs *CphService) GetHeightByTxid(txid string) (int64, error) {
	return 0, errors.New("unsupport it")
}

func (cs *CphService) GetTxIsExist(height int64, txid string) bool {
	panic("implement me")
}

var CPHDecimal int32 = 18

func NewScanning(cfg conf.Config, nodeCfg conf.NodeConfig) common.IScanner {
	cs := new(CphService)
	cs.cfg = cfg
	cs.nodeCfg = nodeCfg
	cs.client, _ = common.NewRpcClient(nodeCfg.Url, nodeCfg.RPCKey, nodeCfg.RPCSecret)
	return cs
}

func (cs *CphService) GetLatestBlockHeight() (int64, error) {
	var err error
	if cs.client == nil {
		cs.client, err = common.NewRpcClient(cs.nodeCfg.Url, cs.nodeCfg.RPCKey, cs.nodeCfg.RPCSecret)
		if err != nil {
			return -1, fmt.Errorf("reconnect rpc client error: %v", err)
		}
	}
	var resp interface{}
	err = cs.client.Post("cph_txBlockNumber", &resp, []interface{}{})
	if err != nil {
		return -1, fmt.Errorf("cph get latest block number error: %v", err)
	}

	//var data []byte
	//data,err = json.Marshal(resp)
	//if err != nil {
	//	return -1, err
	//}
	lb := resp.(string)
	latestBlockHeight := cs.parseHexToInt64(lb)

	if latestBlockHeight < 0 {
		return latestBlockHeight, fmt.Errorf("get latest block height error,is less than zero")
	}

	return latestBlockHeight, nil
}

func (cs *CphService) GetBlockByHeight(height int64) (*common.BlockData, error) {
	var err error
	if cs.client == nil {
		cs.client, err = common.NewRpcClient(cs.nodeCfg.Url, cs.nodeCfg.RPCKey, cs.nodeCfg.RPCSecret)
		if err != nil {
			return nil, fmt.Errorf("reconnect rpc client error: %v", err)
		}
	}
	h := cs.encodeIntToHex(height)
	var resp cph.CphBlockStruct
	err = cs.client.Post("cph_getTxBlockByNumber", &resp, []interface{}{h, true, false})
	if err != nil {
		return nil, fmt.Errorf("get block by number: %d error: %v", height, err)
	}
	bd := new(common.BlockData)
	bd.Hash = resp.Hash
	bd.TxIds = resp.Transactions
	bd.Height = height
	bd.Timestamp = cs.parseHexToInt64(resp.Timestamp)
	bd.TxNums = len(resp.Transactions)
	bd.NextHash = ""
	bd.PrevHash = resp.ParentHash
	bd.Confirmation = cs.cfg.Sync.Confirmations + 1

	return bd, nil
}

func (cs *CphService) GetTxData(blockData *common.BlockData, txid string, isWatchAddress common.IsWatchAddress, isContractTx common.IsContractTx) (*common.TxData, error) {
	var err error
	if cs.client == nil {
		cs.client, err = common.NewRpcClient(cs.nodeCfg.Url, cs.nodeCfg.RPCKey, cs.nodeCfg.RPCSecret)
		if err != nil {
			return nil, fmt.Errorf("reconnect rpc client error: %v", err)
		}
	}

	var txReceipt cph.CphTransactionReceipt
	err = cs.client.Post("cph_getTransactionReceipt", &txReceipt, []interface{}{txid})
	if err != nil {
		return nil, fmt.Errorf("get tx receipt error: %v,txid=%s", err, txid)
	}
	td := new(common.TxData)
	td.IsFakeTx = false
	// todo 处理合约交易
	if txReceipt.ContractAddress != "" && txReceipt.ContractAddress != "null" {
		_, ok := isContractTx(txReceipt.ContractAddress)
		//合约交易
		if !ok {
			return nil, fmt.Errorf("未监听该合约交易：%s", txReceipt.ContractAddress)
		}

		td.ContractAddress = txReceipt.ContractAddress
		if txReceipt.Status != "0x1" {
			//假充值
			td.IsFakeTx = true
			return td, nil
		}
	}
	//根据txid 获取交易详情
	var resp cph.CphTransactionStruct
	err = cs.client.Post("cph_getTransactionByHash", &resp, []interface{}{txid})
	if err != nil {
		return nil, fmt.Errorf("get transaction by hash: %s error: %v", txid, err)
	}
	var (
		from, to, amount string
	)
	from = resp.From

	to = resp.To
	amt := cs.parseHexToInt64(resp.Value)
	if amt < 0 {
		return nil, fmt.Errorf("parse amount error,txid=%s", txid)
	}
	ad := decimal.NewFromInt(amt)

	amount = ad.Shift(-CPHDecimal).String()

	gas := cs.parseHexToInt64(resp.Gas)
	gasD := decimal.NewFromInt(gas)
	gasPrice := cs.parseHexToInt64(resp.GasPrice)
	gasPriceD := decimal.NewFromInt(gasPrice)

	fee := gasD.Mul(gasPriceD).Shift(-CPHDecimal).String()
	td.FromAddr = from
	td.ToAddr = to
	td.Amount = amount
	td.Txid = txid
	//计算手续费
	td.Fee = fee
	return td, nil
}

func (cs *CphService) parseHexToInt64(hexStr string) int64 {
	bi := cs.parseHexToBigInt(hexStr)
	if bi == nil {
		return -1
	}
	return bi.Int64()
}
func (cs *CphService) parseHexToBigInt(hexStr string) *big.Int {
	if hexStr == "" {

		return nil
	}
	if strings.HasPrefix(hexStr, "0x") {
		hexStr = hexStr[2:]
	}
	bi := new(big.Int)
	bi.SetString(hexStr, 16)
	if bi.Sign() < 0 {
		log.Errorf("parse hex to big int,error,sign is less than 0")
		return nil
	}
	return bi
}
func (cs *CphService) encodeIntToHex(i int64) string {
	hexStr := strconv.FormatInt(i, 16)
	hexStr = strings.TrimPrefix(hexStr, "0x")
	return "0x" + hexStr
}
