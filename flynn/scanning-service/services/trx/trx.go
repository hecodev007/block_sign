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
	"github.com/group-coldwallet/scanning-service/common"
	"github.com/group-coldwallet/scanning-service/conf"
	"github.com/group-coldwallet/scanning-service/log"
	"github.com/group-coldwallet/scanning-service/models/po"
	"github.com/group-coldwallet/scanning-service/utils"
	"github.com/group-coldwallet/scanning-service/utils/dingding"
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
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

//{"Height":0,"Txid":"1d85d2634bdacdf791a92c2161ee4e574d9815e77d2b84f491281512ce051701",
//"IsFakeTx":false,"FromAddr":"","ToAddr":"","Amount":"","Fee":"","Memo":"",
//"ContractAddress":"TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t","IsContainTx":false,"MainDecimal":0}",

func (ts *TrxService) GetTxData(blockData *common.BlockData, txid string, isWatchAddress common.IsWatchAddress, isContractTx common.IsContractTx) (*common.TxData, error) {

	txInfo, err := ts.client.GRPC.GetTransactionByID(txid)
	if err != nil || txInfo == nil {
		if err != nil {
			log.Infof("txId:%s 处理失败,Err:%s", txid, err.Error())
		} else {
			log.Infof("txId:%s 处理失败,txInfo is null", txid)
		}
		return nil, fmt.Errorf("get transaction  error: %v", err)
	} else {
		//log.Infof("txId:%s 已处理 ,txInfo is: %s", txid ,utils.DumpJSON(txInfo))
	}
	td := new(common.TxData)
	if blockData == nil {
		log.Infof("txId:%s  ,blockData is: nil", txid)
		info, err := ts.client.GRPC.GetTransactionInfoByID(txid)
		if err != nil {
			log.Infof("txId:%s 处理失败,Err:%s", txid, err.Error())
			return nil, fmt.Errorf("get transaction info error: %v", err)
		}
		log.Infof("txId:%s  ,blockData 获取结果: %s", txid, utils.DumpJSON(blockData))
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
		log.Infof("txId:%s 处理失败,Err: invalid contracts", txid)
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
		//log.Infof("switch txId:%s. case core.Transaction_Contract_TransferContract",txid)
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
		//log.Infof("switch txId:%s. case core.Transaction_Contract_TransferAssetContract",txid)
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
		//log.Infof("switch txId:%s. case core.Transaction_Contract_TriggerSmartContract",txid)
		var c core.TriggerSmartContract
		if err = ptypes.UnmarshalAny(contract.GetParameter(), &c); err != nil {
			return nil, fmt.Errorf("trc20 inconsistent")
		}
		tv = structs.Map(c)
		//log.Infof("txId:%s. tv is: %s",txid , utils.DumpJSON(tv))
		if v, ok := tv["ContractAddress"]; ok && len(v.([]uint8)) > 0 {
			contractAddress := address.Address(v.([]uint8)).String() //这里获取的是交易的合约地址，并不是TRC20 Transfer里面的合约地址，需要区分
			transactionInfo, err := ts.client.GRPC.GetTransactionInfoByID(txid)
			ci, isExist := isContractTx(contractAddress)
			if isExist {
				td.ContractAddress = contractAddress
				if v, ok := tv["OwnerAddress"]; ok && len(v.([]uint8)) > 0 {
					from = address.Address(v.([]uint8)).String()
					//log.Infof("txId:%s. from is: %s",txid , from)
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
					log.Infof("txId:%s. to is: %s", txid, to)
				}
			} else if err == nil && transactionInfo.Receipt.Result == core.Transaction_Result_SUCCESS {
				//TRC20 Transfer 解析 统一通过Receipt.Result作为交易成功判断
				amount, from, to = ts.contractTxTran(transactionInfo, contractAddress, ci, isExist, isContractTx, td, amount, from, to, isWatchAddress, txid)
				if transactionInfo.InternalTransactions != nil && from == "" && to == "" { //内部转账
					from, to, amount = ts.InternalTransactions(transactionInfo, from, to, amount, isWatchAddress, txid)
				}
			} else {
				return nil, nil
			}
		}
	default:
		log.Infof("txId:%s 处理失败,Err: 不支持的类型", txid)
		// 不支持的type类型
	}
	if isWatchAddress(from) || isWatchAddress(to) {
		log.Infof("txId:%s , 有需要关注的地址, from: %s, to: %s.", txid, from, to)
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
			if err != nil {
				log.Infof("txId:%s 获取手续费失败,Err:%s", txid, err.Error())
			} else {
				log.Infof("txId:%s 获取手续费失败,txInfo is null", txid)
			}
			return nil, fmt.Errorf("get transaction info error: %v", err)
		}

		fee := ti.Receipt.GetEnergyFee() + ti.Receipt.GetNetFee()
		if fee > 0 {
			td.Fee = decimal.NewFromInt(fee).Shift(-trxDecimal).String()
		}
	} else {
		log.Infof("txId:%s , 没有需要关注的地址, from: %s, to: %s.", txid, from, to)
	}
	return td, nil
}

func (ts *TrxService) contractTxTran(transactionInfo *core.TransactionInfo, contractAddress string, ci *po.ContractInfo, isExist bool, isContractTx common.IsContractTx, td *common.TxData, amount string, from string, to string, isWatchAddress common.IsWatchAddress, txid string) (string, string, string) {
	if len(transactionInfo.Log) > 0 {
		for _, value := range transactionInfo.Log {
			//contractAddress 是传递后面的值，需要问jan
			contractAddress = genkeys.AddressHexToB58("41" + hex.EncodeToString(value.GetAddress()))
			ci, isExist = isContractTx(contractAddress)
			if isExist {
				td.ContractAddress = contractAddress
				infologTopics := value.GetTopics()
				if hex.EncodeToString(infologTopics[0]) == "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef" { //Transfer(from,to,value)
					amt := new(big.Int)
					//getData := hex.EncodeToString(value.GetData())
					amt.SetString(hex.EncodeToString(value.GetData()), 16)
					amount = decimal.NewFromBigInt(amt, -int32(ci.Decimal)).String()
					from = genkeys.AddressHexToB58("41" + hex.EncodeToString(infologTopics[1])[24:])
					to = genkeys.AddressHexToB58("41" + hex.EncodeToString(infologTopics[2])[24:])
					log.Infof("%s,%s,%s,%s", contractAddress, from, amount, to)
					if isWatchAddress(from) || isWatchAddress(to) {
						//钉钉通知
						var sb strings.Builder
						sb.WriteString("TRX捕捉到一条合约交易：" + txid + "\n")
						sb.WriteString("合约：" + contractAddress + "\n")
						sb.WriteString("from：" + from + "\n")
						sb.WriteString("to：" + to + "\n")
						sb.WriteString("amount：" + amount + "\n")
						dingding.NotifyError(sb.String())
						break
					}
				}
			}
		}
	}
	return amount, from, to
}

func (ts *TrxService) InternalTransactions(transactionInfo *core.TransactionInfo, from string, to string, amount string, isWatchAddress common.IsWatchAddress, txid string) (string, string, string) {
	for _, InternalTransactions := range transactionInfo.InternalTransactions {
		if InternalTransactions.CallValueInfo == nil {
			continue
		}
		for _, CallValueInfoItem := range InternalTransactions.CallValueInfo {
			if CallValueInfoItem.CallValue > int64(0) { //内部转账有0的情况
				from = genkeys.AddressHexToB58(hex.EncodeToString(InternalTransactions.GetCallerAddress()))
				to = genkeys.AddressHexToB58(hex.EncodeToString(InternalTransactions.GetTransferToAddress()))
				amt := new(big.Int)
				CallValue := strconv.FormatInt(CallValueInfoItem.CallValue, 10)
				amt.SetString(CallValue, 10)
				amount = decimal.NewFromBigInt(amt, -trxDecimal).String()
				log.Infof("form:%s,to:%s,amount:%d", from, to, amount)
				if isWatchAddress(from) || isWatchAddress(to) {
					//钉钉通知
					var sb strings.Builder
					sb.WriteString("TRX捕捉到一条合约内部交易：" + txid + "\n")
					sb.WriteString("from：" + from + "\n")
					sb.WriteString("to：" + to + "\n")
					sb.WriteString("amount：" + amount + "\n")
					dingding.NotifyError(sb.String())
					break
				}
			}
		}
	}
	return from, to, amount
}
