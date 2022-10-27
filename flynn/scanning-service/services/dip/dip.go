package dip

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Dipper-Labs/go-sdk/constants"
	"github.com/group-coldwallet/scanning-service/common"
	"github.com/group-coldwallet/scanning-service/conf"
	"github.com/group-coldwallet/scanning-service/models/dip"
	"github.com/shopspring/decimal"
)

const (
	DipDecimal = 12
)

type DipService struct {
	cfg     conf.Config
	nodeCfg conf.NodeConfig
}

func (d *DipService) GetHeightByTxid(txid string) (int64, error) {
	return 0, errors.New("unsupport it")
}

func (d *DipService) GetLatestBlockHeight() (int64, error) {
	data, err := d.httpGet("/blocks/latest")
	if err != nil {
		return 0, fmt.Errorf("get latest block error: %v", err)
	}
	result := new(dip.ResponseBlock)
	err = json.Unmarshal(data, result)
	if err != nil {
		return 0, fmt.Errorf("json unmarshal latest block error: %v", err)
	}
	return common.ParseInt64(result.BlockMeta.Header.Height)
}

func (d *DipService) GetBlockByHeight(height int64) (*common.BlockData, error) {
	path := fmt.Sprintf("/blocks/%d", height)
	data, err := d.httpGet(path)
	if err != nil {
		return nil, fmt.Errorf("get block %d error: %v", height, err)
	}
	result := new(dip.ResponseBlock)
	err = json.Unmarshal(data, result)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal block %d error: %v", height, err)
	}
	bd := new(common.BlockData)
	bd.Hash = result.BlockMeta.BlockID.Hash
	bd.PrevHash = result.Block.Herder.LastBlockID.Hash
	bd.Timestamp = result.Block.Herder.Time.Unix()
	bd.Height, _ = common.ParseInt64(result.Block.Herder.Height)
	bd.Confirmation = d.cfg.Sync.Confirmations + 1
	bd.TxNums = len(result.Block.Data.Txs)
	for _, tx := range result.Block.Data.Txs {
		bd.TxIds = append(bd.TxIds, fmt.Sprintf("%X", tx.Hash()))
	}
	return bd, nil
}

func (d *DipService) GetTxData(blockData *common.BlockData, txid string, isWatchAddress common.IsWatchAddress, isContractTx common.IsContractTx) (*common.TxData, error) {
	path := fmt.Sprintf("/txs/%s", txid)
	data, err := d.httpGet(path)
	if err != nil {
		return nil, fmt.Errorf("get tx data error: %v,txid=%s", err, txid)
	}
	proxy := new(dip.ResponseTx)
	err = json.Unmarshal(data, proxy)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal tx data error: %v,txid=%s", err, txid)
	}
	if proxy.Error != "" {
		return nil, fmt.Errorf("tx data has error: %v", proxy.Error)
	}

	td := new(common.TxData)
	td.Txid = proxy.TxHash
	if proxy.Tx.Type != "dip/StdTx" {
		td.IsContainTx = false
		return td, nil
	}
	//判断是否是假充值
	for _, log := range proxy.Logs {
		if !log.Success {
			td.IsFakeTx = true
			return td, nil
		}
	}
	td.Memo = proxy.Tx.Value.Memo
	gasWant, err := decimal.NewFromString(proxy.GasWanted)
	if err != nil || gasWant == decimal.Zero {
		return nil, fmt.Errorf("get gas want error")
	}
	gasUsed, err := decimal.NewFromString(proxy.GasUsed)
	if err != nil {
		return nil, fmt.Errorf("get gas used error")
	}
	ratio := gasUsed.Div(gasWant)
	gasMax := decimal.NewFromInt(proxy.Tx.Value.Fee.Amount.AmountOf(constants.TxDefaultDenom).Int64())
	td.Fee = gasMax.Mul(ratio).Shift(-DipDecimal).String()
	msgs := proxy.Tx.Value.Msgs
	for _, tmp := range msgs {
		switch tmp.Type {
		case "dip/MsgSend":
			valuebyte, _ := json.Marshal(tmp.Value)
			var msgSend dip.MsgSend
			err := json.Unmarshal(valuebyte, &msgSend)
			if err != nil {
				return nil, fmt.Errorf("json. unmarshal msg send error: %v", err)
			}
			if len(msgSend.Amount) != 1 {
				return nil, fmt.Errorf("这笔交易的amount数量不等于1，请检查：%s", txid)
			}
			amount := msgSend.Amount[0]
			if amount.Denom != constants.TxDefaultDenom {
				return nil, fmt.Errorf("这笔交易不是dip，有可能是代币转账,denom=%s", amount.Denom)
			}
			td.Amount = decimal.NewFromInt(amount.Amount.Int64()).Shift(-DipDecimal).String()
			td.FromAddr = msgSend.FromAddress
			td.ToAddr = msgSend.ToAddress
			//判断是否是监听的地址
			if isWatchAddress(td.FromAddr) || isWatchAddress(td.ToAddr) {
				td.IsContainTx = true
			}
		default:
			break
		}
	}
	return td, nil
}

func (d *DipService) GetTxIsExist(height int64, txid string) bool {
	path := fmt.Sprintf("/txs/%s", txid)
	data, err := d.httpGet(path)
	if err != nil {
		return false
	}
	proxy := new(dip.ResponseTx)
	err = json.Unmarshal(data, proxy)
	if err != nil {
		return false
	}
	if proxy.Error != "" {
		return false
	}
	//判断是否是假充值
	for _, log := range proxy.Logs {
		if !log.Success {
			return false
		}
	}
	return true
}

func NewScanning(cfg conf.Config, nodeCfg conf.NodeConfig) common.IScanner {
	ds := new(DipService)
	ds.cfg = cfg
	ds.nodeCfg = nodeCfg

	return ds
}
