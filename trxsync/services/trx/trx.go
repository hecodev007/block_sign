package trx

import (
	"encoding/hex"
	"fmt"
	"github.com/JFJun/trx-sign-go/genkeys"
	"github.com/JFJun/trx-sign-go/grpcs"
	"github.com/fatih/structs"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/golang/protobuf/ptypes"
	"github.com/group-coldwallet/trxsync/common"
	"github.com/group-coldwallet/trxsync/conf"
	"github.com/shopspring/decimal"
	"math/big"
	"strings"
	"sync"
)

type TrxService struct {
	cfg          conf.Config
	nodeCfg      conf.NodeConfig
	client       *grpcs.Client
	latestHeight int64
	url          string
	lock         sync.RWMutex
}

func (ts *TrxService) GetHeightByTxid(txid string) (int64, error) {
	info, err := ts.client.GRPC.GetTransactionInfoByID(txid)
	if err != nil {
		return 0, fmt.Errorf("get transaction info error: %v", err)
	}
	height := info.BlockNumber
	return height, nil
}

func (ts *TrxService) GetTxIsExist(height int64, txid string) bool {
	txInfo, err := ts.client.GRPC.GetTransactionByID(txid)
	if err != nil || txInfo == nil {
		return false
	}
	for _, ret := range txInfo.Ret {
		if ret.GetContractRet() != core.Transaction_Result_SUCCESS {
			return false
		}
	}
	return true
}

func NewScanning(cfg conf.Config, nodeCfg conf.NodeConfig) common.IScanner {
	ts := new(TrxService)
	ts.cfg = cfg
	ts.url = strings.TrimPrefix(nodeCfg.Url, "http://")
	ts.nodeCfg = nodeCfg
	ts.client, _ = grpcs.NewClient(ts.url)
	ts.lock = sync.RWMutex{}
	return ts
}
func (ts *TrxService) getUrl(method string) string {
	url := fmt.Sprintf("%s%s", ts.url, method)
	return url
}

func (ts *TrxService) GetLatestBlockHeight() (int64, error) {
	block, err := ts.client.GRPC.GetNowBlock()
	if err != nil || block == nil {
		return -1, fmt.Errorf("get latest block height error: %v", err)
	}
	ts.lock.Lock()
	ts.latestHeight = block.BlockHeader.RawData.Number
	defer ts.lock.Unlock()
	return block.BlockHeader.RawData.Number, nil
}

func (ts *TrxService) GetBlockByHeight(height int64) (*common.BlockData, error) {
	block, err := ts.client.GRPC.GetBlockByNum(height)
	if err != nil || block == nil {
		return nil, fmt.Errorf("get block by height error: %v,height: %d", err, height)
	}
	bd := new(common.BlockData)
	bd.Height = block.BlockHeader.RawData.Number
	bd.Hash = hex.EncodeToString(block.Blockid)
	bd.PrevHash = hex.EncodeToString(block.BlockHeader.RawData.ParentHash)
	bd.Timestamp = block.BlockHeader.RawData.Timestamp
	bd.Confirmation = ts.latestHeight - height
	//获取交易数量
	txInfo, err := ts.client.GRPC.GetBlockInfoByNum(height)
	if err != nil || txInfo == nil {
		return nil, fmt.Errorf("get block tx info error: %v", err)
	}
	if len(txInfo.TransactionInfo) > 0 {
		bd.TxNums = len(txInfo.TransactionInfo)
	}
	for _, info := range txInfo.TransactionInfo {
		if info.GetResult() == core.TransactionInfo_FAILED {
			//log.Infof("发现一笔假充值============>,%s",hex.EncodeToString(info.Id))
			continue
		}
		bd.TxIds = append(bd.TxIds, hex.EncodeToString(info.Id))
	}

	return bd, nil
}

func (ts *TrxService) GetTxData(blockData *common.BlockData, txid string, isWatchAddress common.IsWatchAddress, isContractTx common.IsContractTx) (*common.TxData, error) {

	txInfo, err := ts.client.GRPC.GetTransactionByID(txid)
	if err != nil || txInfo == nil {
		return nil, fmt.Errorf("get transaction  error: %v", err)
	}
	td := new(common.TxData)
	if blockData == nil {
		info, err := ts.client.GRPC.GetTransactionInfoByID(txid)
		if err != nil {
			return nil, fmt.Errorf("get transaction info error: %v", err)
		}
		td.Height = info.BlockNumber
	}
	td.Txid = txid
	//1. 先判断交易的状态
	td.IsFakeTx = true
	for _, ret := range txInfo.Ret {
		if ret.GetContractRet() == core.Transaction_Result_SUCCESS {
			td.IsFakeTx = false
			// return td,nil
		}
	}
	contracts := txInfo.GetRawData().GetContract()
	if len(contracts) != 1 {
		return nil, fmt.Errorf("invalid contracts")
	}
	contract := contracts[0]
	//2. 判断是什么交易

	var (
		tv               map[string]interface{}
		from, to, amount string
	)

	switch contract.GetType() {
	case core.Transaction_Contract_TransferContract:
		var c core.TransferContract
		if err = ptypes.UnmarshalAny(contract.GetParameter(), &c); err != nil {
			return nil, fmt.Errorf("trx inconsistent")
		}
		tv = structs.Map(c)
		if v, ok := tv["OwnerAddress"]; ok && len(v.([]uint8)) > 0 {
			from = address.Address(v.([]uint8)).String()
		}
		if v, ok := tv["ToAddress"]; ok && len(v.([]uint8)) > 0 {
			to = address.Address(v.([]uint8)).String()
		}
		if v, ok := tv["Amount"]; ok {
			amount = decimal.NewFromInt(v.(int64)).Shift(-trxDecimal).String()
		}

	case core.Transaction_Contract_TransferAssetContract:
		var c core.TransferAssetContract
		if err = ptypes.UnmarshalAny(contract.GetParameter(), &c); err != nil {
			return nil, fmt.Errorf("trc10 inconsistent")
		}
		tv = structs.Map(c)
		if v, ok := tv["AssetName"]; ok && len(v.([]uint8)) > 0 {
			assetName := string(v.([]uint8))
			ci, isExist := isContractTx(assetName)
			if isExist {
				td.ContractAddress = assetName
				if v, ok := tv["OwnerAddress"]; ok && len(v.([]uint8)) > 0 {
					from = address.Address(v.([]uint8)).String()
				}
				if v, ok := tv["ToAddress"]; ok && len(v.([]uint8)) > 0 {
					to = address.Address(v.([]uint8)).String()
				}
				if v, ok := tv["Amount"]; ok {
					amount = decimal.NewFromInt(v.(int64)).Shift(-int32(ci.Decimal)).String()
				}
			}
		}
	case core.Transaction_Contract_TriggerSmartContract:
		var c core.TriggerSmartContract
		if err = ptypes.UnmarshalAny(contract.GetParameter(), &c); err != nil {
			return nil, fmt.Errorf("trc20 inconsistent")
		}
		tv = structs.Map(c)
		if v, ok := tv["ContractAddress"]; ok && len(v.([]uint8)) > 0 {
			contractAddress := address.Address(v.([]uint8)).String()
			ci, isExist := isContractTx(contractAddress)
			if isExist {
				td.ContractAddress = contractAddress
				if v, ok := tv["OwnerAddress"]; ok && len(v.([]uint8)) > 0 {
					from = address.Address(v.([]uint8)).String()
				}
				if v, ok := tv["Data"]; ok && len(v.([]uint8)) > 0 {
					data := hex.EncodeToString(v.([]uint8))
					//解析trc20 data
					if len(data) != 136 { //8 +64 +64
						return nil, fmt.Errorf("trc20 data length is not equal 136,data=[%s]", data)
					}
					if !strings.HasPrefix(data, trc20TransferMethodSignature) {
						return nil, nil
					}
					toAddress := data[len(trc20TransferMethodSignature) : len(trc20TransferMethodSignature)+64]
					amountHex := data[len(data)-64:]
					amt := new(big.Int)
					amt.SetString(amountHex, 16)
					if amt.Sign() < 0 {
						return nil, fmt.Errorf("parse trc20 data amount error: %s", "amount  sign is less 0")
					}
					amount = decimal.NewFromBigInt(amt, -int32(ci.Decimal)).String()
					to = genkeys.AddressHexToB58("41" + toAddress[len(amountHex)-40:])
				}
			} else {
				return nil, nil
			}
		}
	default:
		// 不支持的type类型
	}
	if isWatchAddress(from) || isWatchAddress(to) {
		td.IsContainTx = true
		td.FromAddr = from
		td.ToAddr = to
		if !td.IsFakeTx {
			td.Amount = amount
		} else {
			td.Amount = "0"
		}
		//获取手续费
		ti, err := ts.client.GRPC.GetTransactionInfoByID(txid)
		if err != nil || ti == nil {
			return nil, fmt.Errorf("get transaction info error: %v", err)
		}

		fee := ti.Receipt.GetEnergyFee() + ti.Receipt.GetNetFee()
		if fee > 0 {
			td.Fee = decimal.NewFromInt(fee).Shift(-trxDecimal).String()
		}
	}
	return td, nil
}
