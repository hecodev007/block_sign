package heco

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/group-coldwallet/scanning-service/common"
	"github.com/group-coldwallet/scanning-service/common/eth"
	"github.com/group-coldwallet/scanning-service/conf"
	"github.com/group-coldwallet/scanning-service/models/hecoModel"
	"github.com/group-coldwallet/scanning-service/models/po"
	"github.com/shopspring/decimal"
	"math/big"
	"strings"
)

const (
	HecoDecimal = 18
)

type HecoService struct {
	cfg     conf.Config
	nodeCfg conf.NodeConfig
	client  *eth.EthRpcClient
}

func (d *HecoService) GetHeightByTxid(txid string) (int64, error) {
	var resp hecoModel.Transaction
	err := d.client.GetTransactionByHash(txid, &resp)
	if err != nil {
		return 0, err
	}
	return common.ParseInt64(resp.BlockNumber)

}

func (d *HecoService) GetLatestBlockHeight() (int64, error) {
	return d.client.GetLatestBlockHeight()
}

func (d *HecoService) GetBlockByHeight(height int64) (*common.BlockData, error) {
	/*
		heco做一个限制，当高度小于2020-02-26时的高度时，不让补推数据
	*/
	if height < 2468015 {
		return nil, errors.New("补推高度小于2468015，不允许补推")
	}
	var block hecoModel.Block
	err := d.client.GetBlockByNumber(height, &block, false)
	if err != nil {
		return nil, err
	}
	bd := new(common.BlockData)
	bd.Hash = block.Hash
	bd.Height, _ = common.ParseInt64(block.Number)
	bd.PrevHash = block.ParentHash
	bd.Height, _ = common.ParseInt64(block.Number)
	bd.TxNums = len(block.Transactions)
	bd.Confirmation = d.cfg.Sync.Confirmations + 1
	timestamp, _ := common.ParseInt64(block.Timestamp)
	bd.Timestamp = common.Int64ToTime(timestamp).Unix()
	var txs []string
	if len(block.Transactions) > 0 {
		for _, tx := range block.Transactions {
			txid, ok := tx.(string)
			if ok {
				txs = append(txs, txid)
			}
		}
	}
	bd.TxIds = txs
	return bd, err

}

func (d *HecoService) GetTxData(blockData *common.BlockData, txid string, isWatchAddress common.IsWatchAddress, isContractTx common.IsContractTx) (*common.TxData, error) {
	var txReceipt hecoModel.TransactionReceipt
	err := d.client.GetTransactionReceipt(txid, &txReceipt)
	if err != nil {
		return nil, err
	}
	td := new(common.TxData)
	// 默认交易都是假充值
	td.IsFakeTx = true
	// heco的主链状态也是 "0x1"
	if txReceipt.Status == "0x1" {
		td.IsFakeTx = false
	}
	var resp hecoModel.Transaction
	err = d.client.GetTransactionByHash(txid, &resp)
	if err != nil {
		return nil, err
	}

	// 判断是否是合约交易
	var (
		from, to, amount string
		amtInt           *big.Int
		coin             *po.ContractInfo
	)
	from = resp.From

	//表示是合约交易
	if d.IsContract(resp.Input, resp.To) {
		if len(txReceipt.Logs) == 0 {
			td.IsFakeTx = true
		}
		var ok bool
		coin, ok = isContractTx(resp.To)
		if !ok {
			td.IsContainTx = false
			return td, nil
		}
		if coin == nil {
			return nil, fmt.Errorf("get contract(%s) info is null", resp.To)
		}
		td.ContractAddress = resp.To

		toAddr, amt, err := d.ParseTransferData(resp.Input)
		if err != nil {
			return nil, fmt.Errorf("parse txid(%s) input error:%v", txid, err)
		}
		amount = decimal.NewFromBigInt(amt, 0).Shift(-int32(coin.Decimal)).String()
		to = toAddr
	} else {
		to = resp.To
		amtInt, err = common.ParseBigInt(resp.Value)
		if err != nil {
			return nil, fmt.Errorf("txid(%s) parse amount(%s) error: %v", txid, resp.Value, err)
		}

		ad := decimal.NewFromBigInt(amtInt, 0)
		if ad.LessThan(decimal.Zero) {
			return nil, fmt.Errorf("parse amount(%v) error,txid=%s", amtInt, txid)
		}
		amount = ad.Shift(-HecoDecimal).String()
	}
	//if len(txReceipt.Logs)>0 {
	//	for _,l:=range txReceipt.Logs{
	//		if strings.ToLower(l.Address)==strings.ToLower(txReceipt.To) {
	//			//合约交易
	//			coin,isOK = isContractTx(txReceipt.To)
	//			if ! isOK {
	//				//
	//				//log.Infof("未监听该合约交易：%s",txReceipt.ContractAddress)
	//				td.IsContainTx = false
	//				return td, nil
	//			}
	//			if coin==nil {
	//				return nil, fmt.Errorf("get contract(%s) info is null",txReceipt.To)
	//			}
	//			td.ContractAddress  = txReceipt.To
	//		}
	//	}
	//}

	//var (
	//	from,to,amount string
	//	amtInt *big.Int
	//)
	//from =  resp.From
	//if isOK{
	//	toAddr,amt,err:=d.ParseTransferData(resp.Input)
	//	if err != nil {
	//		return nil, fmt.Errorf("parse txid(%s) input error:%v",txid,err)
	//	}
	//	amount = decimal.NewFromBigInt(amt,0).Shift(-int32(coin.Decimal)).String()
	//	to = toAddr
	//}else{
	//	to = resp.To
	//	amtInt,err = common.ParseBigInt(resp.Value)
	//	if err != nil {
	//		return nil, fmt.Errorf("txid(%s) parse amount(%s) error: %v",txid,resp.Value,err)
	//	}
	//
	//	ad:=decimal.NewFromBigInt(amtInt,0)
	//	if ad.LessThan(decimal.Zero){
	//		return nil,fmt.Errorf("parse amount(%v) error,txid=%s",amtInt,txid)
	//	}
	//	amount = ad.Shift(-HecoDecimal).String()
	//}

	// 判断是否是我们的地址，如果不是，就不处理这笔交易

	if isWatchAddress(from) || isWatchAddress(to) {
		td.IsContainTx = true

		gas, _ := common.ParseInt64(txReceipt.GasUsed)
		gasD := decimal.NewFromInt(gas)
		gasPrice, _ := common.ParseInt64(resp.GasPrice)
		gasPriceD := decimal.NewFromInt(gasPrice)

		fee := gasD.Mul(gasPriceD).Shift(-HecoDecimal).String()
		td.FromAddr = from
		td.ToAddr = to

		//避免假充值
		if !td.IsFakeTx {
			td.Amount = amount
		} else {
			td.Amount = "0"
		}
		td.Txid = txid
		//计算手续费
		td.Fee = fee
	}
	return td, nil
}

func (d *HecoService) GetTxIsExist(height int64, txid string) bool {
	var txReceipt hecoModel.TransactionReceipt
	err := d.client.GetTransactionReceipt(txid, &txReceipt)
	if err != nil {
		log.Error(err)
		return false
	}
	return true
}

func NewScanning(cfg conf.Config, nodeCfg conf.NodeConfig) common.IScanner {
	ds := new(HecoService)
	ds.cfg = cfg
	ds.nodeCfg = nodeCfg
	ds.client = eth.NewEthRpcClient(nodeCfg.Url, nodeCfg.RPCKey, nodeCfg.RPCSecret)
	return ds
}

func (d *HecoService) ParseTransferData(input string) (to string, amount *big.Int, err error) {

	//0xa9059cbb0000000000000000000000005237bc08b2fe644487366e246741bd7ec0eb24710000000000000000000000000000000000000000000000000000000005f5e100
	if !strings.HasPrefix(input, "0xa9059cbb") {
		//处理代付模式
		datas := strings.Split(input, "a9059cbb")
		if len(datas) < 2 {
			return to, amount, errors.New("input is not transfer data")
		}
		if len(datas) > 2 {
			return "", nil, fmt.Errorf("无法解析该input数据，该input数据 有多个【a9059cbb】，我们只允许又一个")
		}
		newInput := "0xa9059cbb" + datas[1]
		input = newInput
	}
	if len(input) < 138 {
		return to, amount, fmt.Errorf("input data isn't 138 , size %d ", 138)
	}
	to = "0x" + input[34:74]
	amount = new(big.Int)
	amount.SetString(input[74:138], 16)
	if amount.Sign() < 0 {
		return to, amount, errors.New("bad amount data")
	}
	return to, amount, nil
}

func (d *HecoService) IsContract(input, to string) bool {
	if input == "0x" || input == "0x00" {
		return false
	}
	var code string
	err := d.client.GetCode(to, &code)
	if err != nil {
		return false
	}
	if code == "0x" {
		return false
	}
	return true
}
