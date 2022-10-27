package v1

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/group-coldwallet/bsc-sign/conf"
	"github.com/group-coldwallet/bsc-sign/model"
	"github.com/group-coldwallet/bsc-sign/util"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"math/big"
	"strings"
	"sync"
)

/*
service模板
*/

/*
币种服务结构体
*/
type BscService struct {
	*BaseService
	client              *util.RpcClient
	nonceCtl, noncePool sync.Map
}

const (
	TRANSFER    = "0xa9059cbb"
	GWei        = 1000000000
	minGasPrice = 15 * GWei
	MaxGasPrice = 100 * GWei
)

/*
初始化币种服务
	注意：
		方法接受者： BaseService
		方法命名： 币种大写 + Service
*/
func (bs *BaseService) BSCService() *BscService {
	cs := new(BscService)
	cs.BaseService = bs
	//初始化连接
	client := util.New(conf.Config.BscCfg.NodeUrl, conf.Config.BscCfg.User, conf.Config.BscCfg.Password)
	cs.client = client
	cs.nonceCtl = sync.Map{}
	// 新增nonce维护池
	cs.noncePool = sync.Map{}
	return cs
}

/*
接口创建地址服务
	无需改动
*/
func (cs *BscService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return cs.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, cs.createAddressInfo)
	}
	return cs.BaseService.createAddress(req, cs.createAddressInfo)
}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (cs *BscService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {

	_, err := cs.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, cs.createAddressInfo)
	return err
}

/*
签名服务
*/
func (cs *BscService) SignService(req *model.ReqSignParams) (interface{}, error) {
	reqData, err := json.Marshal(req.Data)
	if err != nil {
		return nil, err
	}
	var tp model.BscSignParams
	if err := json.Unmarshal(reqData, &tp); err != nil {
		return nil, err
	}

	if err := cs.BaseService.parseData(req, &tp); err != nil {
		return nil, err
	}
	if &tp == nil {
		return nil, errors.New("sign params is null")
	}
	if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	}
	if tp.Nonce < 0 {
		return nil, fmt.Errorf("nonce is less 0: nonce=%d", tp.Nonce)
	}
	var gasPrice *big.Int
	if tp.GasPrice <= 0 {
		gasPrice = big.NewInt(conf.Config.BscCfg.GasPrice)
	} else {
		gasPrice = big.NewInt(tp.GasPrice)
	}
	var gasLimit int64
	if tp.GasLimit <= 0 {
		gasLimit = conf.Config.BscCfg.GasLimit
	} else {
		gasLimit = tp.GasLimit
	}
	nonce := tp.Nonce
	toAmount, err := decimal.NewFromString(tp.Amount)
	if err != nil {
		return nil, fmt.Errorf("parse amount error,err=%v", err)
	}
	amount := toAmount.BigInt()
	toAddress := common.HexToAddress(tp.ToAddress)
	logrus.Printf("出账金额为： %d,手续费为： %d,Nonce: %d", amount.Int64(), gasPrice.Int64()*gasLimit, nonce)
	if strings.Compare(util.Del0xToLower(toAddress.String()), util.Del0xToLower(tp.ToAddress[:])) != 0 {
		return nil, fmt.Errorf("to address is not equal,address1=[%s],address2=[%s]", util.Del0xToLower(toAddress.String()),
			util.Del0xToLower(tp.ToAddress[:]))
	}
	from := tp.FromAddress
	hexPrivateKey, err := cs.BaseService.addressOrPublicKeyToPrivate(from)
	if err != nil {
		return nil, fmt.Errorf("get private key error,Err=%v", err)
	}
	privKey, err := crypto.HexToECDSA(hexPrivateKey)
	if privKey == nil || err != nil {
		return nil, fmt.Errorf("private key is null,err=%v", err)
	}
	var (
		signRes  string
		isSignOk bool
	)
	if tp.ContractAddress != "" {
		//合约 转账
		signRes, isSignOk = cs.bscTokenSign(tp.ToAddress, tp.ContractAddress, uint64(nonce), uint64(gasLimit), amount, gasPrice, privKey)
	} else {
		//转账bsc
		signRes, isSignOk = cs.bscSign(tp.ToAddress, uint64(nonce), uint64(gasLimit), amount, gasPrice, privKey)
	}
	if !isSignOk {
		return nil, fmt.Errorf("bsc sign error,Err=[%s]", signRes)
	}
	hexTx := signRes
	if !strings.HasPrefix(hexTx, "0x") {
		hexTx = "0x" + hexTx
	}
	logrus.Printf("rawTx: %s\n", hexTx)
	return hexTx, nil
}

/*
热钱包出账服务
*/
func (cs *BscService) TransferService(req interface{}) (interface{}, error) {
	var tp model.BscTransferParams
	if err := cs.BaseService.parseData(req, &tp); err != nil {
		return nil, err
	}
	if &tp == nil {
		return nil, errors.New("transfer params is null")
	}
	if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	}
	if cs.client == nil {
		client := util.New(conf.Config.BscCfg.NodeUrl, conf.Config.BscCfg.User, conf.Config.BscCfg.Password)
		cs.client = client
	}
	//nonce := cs.getNonce("eth_getTransactionCount", tp.FromAddress)
	//if nonce < 0 {
	//	return nil, errors.New("get nonce error")
	//}
	//nonce, err := cs.aegisNonce(tp.FromAddress)
	nonce, err := cs.aegisPendingNonce(tp.FromAddress)
	if err != nil {
		return nil, err
	}
	var amount *big.Int
	toAmount, err := decimal.NewFromString(tp.Amount)
	if err != nil {
		return nil, fmt.Errorf("parse amount error,err=%v", err)
	}
	gasLimit := conf.Config.BscCfg.GasLimit
	var (
		gasPrice *big.Int
	)
	//gp, _ := cs.getBuildTxParams("eth_gasPrice", []interface{}{})
	//if gp < 0 {
	//	gasPrice = big.NewInt(conf.Config.BscCfg.GasPrice)
	//	gp = conf.Config.BscCfg.GasPrice
	//}
	//logrus.Infof("获取gas price： %d",gp)
	//gasPrice = big.NewInt(gp)
	//gasPrice = gasPrice.Mul(gasPrice,big.NewInt(13))
	//gasPrice = gasPrice.Div(gasPrice,big.NewInt(10))
	//
	//if gasPrice.Int64()<210000000000 {
	//	gasPrice = big.NewInt(210000000000)
	//}
	gp := cs.confirmFeeBsc(tp.Fee)
	gasPrice = big.NewInt(gp)
	logrus.Infof("实际gas price : %d", gp)
	amount = toAmount.BigInt()
	toAddress := common.HexToAddress(tp.ToAddress)
	logrus.Printf("出账金额为： %s,手续费为： %d,Nonce: %d", amount.String(), gasPrice.Int64()*gasLimit, nonce)
	if strings.Compare(util.Del0xToLower(toAddress.String()), util.Del0xToLower(tp.ToAddress[:])) != 0 {
		return nil, fmt.Errorf("to address is not equal,address1=[%s],address2=[%s]", util.Del0xToLower(toAddress.String()),
			util.Del0xToLower(tp.ToAddress[:]))
	}
	from := tp.FromAddress
	hexPrivateKey, err3 := cs.BaseService.addressOrPublicKeyToPrivate(from)
	if err3 != nil {
		return nil, fmt.Errorf("get private key error,Err=%v", err3)
	}
	privKey, err2 := crypto.HexToECDSA(hexPrivateKey)
	if privKey == nil || err2 != nil {
		return nil, fmt.Errorf("private key is null,err=%v", err)
	}
	var (
		signRes  string
		isSignOk bool
	)
	if tp.ContractAddress != "" {
		//合约 转账
		//safemoon oks safe
		if strings.ToLower(tp.ContractAddress) == "0x3ad9594151886ce8538c1ff615efa2385a8c3a88" ||
			strings.ToLower(tp.ContractAddress) == "0x8076c74c5e3f5852037f31ff0093eeb8c8add8d3" ||
			strings.ToLower(tp.ContractAddress) == "0x5066c68cae3b9bdacd6a1a37c90f2d1723559d18" ||
			strings.ToLower(tp.ContractAddress) == "0x18acf236eb40c0d4824fb8f2582ebbecd325ef6a" {
			logrus.Info("特殊交易，修改limit")
			gasLimit = 1000000
		}
		signRes, isSignOk = cs.bscTokenSign(tp.ToAddress, tp.ContractAddress, uint64(nonce), uint64(gasLimit), amount, gasPrice, privKey)
	} else {
		//转账bsc
		signRes, isSignOk = cs.bscSign(tp.ToAddress, uint64(nonce), uint64(gasLimit), amount, gasPrice, privKey)
	}
	if !isSignOk {
		return nil, fmt.Errorf("bsc sign error,Err=[%s]", signRes)
	}
	hexTx := signRes
	if !strings.HasPrefix(hexTx, "0x") {
		hexTx = "0x" + hexTx
	}
	logrus.Printf("rawTx: %s", hexTx)
	res, err4 := cs.client.SendRequest("eth_sendRawTransaction", []interface{}{hexTx})
	if err4 != nil {
		return nil, fmt.Errorf("send transaction error,Err=[%v]", err4)
	}
	if res == nil {
		return nil, errors.New("send transaction error,response null")
	}
	//本地维护nonce
	cs.noncePool.Store(tp.FromAddress, nonce) //存储当前使用的nonce

	//nm, ok := cs.nonceCtl.Load(tp.FromAddress)
	//if ok {
	//	nonceMap := nm.(map[string]int64)
	//	nonceMap[string(res)] = nonce + 1
	//	cs.nonceCtl.Store(tp.FromAddress, nonceMap)
	//} else {
	//	nonceMap := make(map[string]int64)
	//	nonceMap[string(res)] = nonce + 1
	//	cs.nonceCtl.Store(tp.FromAddress, nonceMap)
	//}
	logrus.Printf("成功出账一笔交易： txid=%s", string(res))
	return string(res), nil
}

func (cs *BscService) getBuildTxParams(method string, params []interface{}) (int64, error) {
	res, err := cs.client.SendRequest(method, params)
	if err != nil {
		logrus.Errorf("rpc send error,Err=%v", err)
		return -1, err
	}
	if res == nil {
		return -1, nil
	}
	ns := string(res)
	var nonceStr string
	if strings.HasPrefix(ns, "0x") {
		nonceStr = ns[2:]
	} else {
		nonceStr = ns
	}
	nonce := util.HexToDec(nonceStr)
	return nonce, nil
}

func (cs *BscService) getNonce(method, address string) int64 {
	params := []interface{}{address, "latest"}
	nonce, _ := cs.getBuildTxParams(method, params)
	return nonce
}

func (cs *BscService) getPendingNonce(method, address string) int64 {
	params := []interface{}{address, "pending"}
	nonce, _ := cs.getBuildTxParams(method, params)
	return nonce
}

/*
创建地址实体方法
*/
func (cs *BscService) createAddressInfo() (util.AddrInfo, error) {
	privkey, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	var (
		addrInfo util.AddrInfo
		address  string
	)
	//避免priv的len不是32
	if len(privkey.D.Bytes()) != 32 {
		for true {
			privkey, err = crypto.GenerateKey()
			if err != nil {
				//if have some error ,cut this exe
				continue
			}

			if len(privkey.D.Bytes()) == 32 {
				break
			}
		}
	}
	if privkey == nil {
		return addrInfo, errors.New("privKey is nil ptr")
	}
	wif := hex.EncodeToString(privkey.D.Bytes())
	address = strings.ToLower(crypto.PubkeyToAddress(privkey.PublicKey).Hex())
	addrInfo.PrivKey = wif
	addrInfo.Address = address
	return addrInfo, nil
}

/*
smart contract abi to build transfer
*/
func (cs *BscService) makeERC20TransferData(toAddress, transfer_method_id string, amount *big.Int) ([]byte, error) {
	var data []byte
	methodId, err := hexutil.Decode(transfer_method_id)
	if err != nil {
		return methodId, err
	}
	data = append(data, methodId...)
	paddedAddress := common.LeftPadBytes(common.HexToAddress(toAddress).Bytes(), 32)
	data = append(data, paddedAddress...)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	data = append(data, paddedAmount...)
	return data, nil

}

/*
bsc sign to build transaction and sign
*/
func (cs *BscService) bscSign(to string, nonce, gasLimit uint64, value, gasPrice *big.Int, privKey *ecdsa.PrivateKey) (string, bool) {
	//build transaction
	tx := types.NewTransaction(nonce, common.HexToAddress(to), value, gasLimit,
		gasPrice, nil)
	//to do eth sign
	return cs.sign(tx, privKey)

}

/*
erc20 token sign to build transaction and sign
*/
func (cs *BscService) bscTokenSign(to, toContractAddr string, nonoce, gasLimit uint64, value, gasPrice *big.Int, privKey *ecdsa.PrivateKey) (string, bool) {
	//make erc20 tranfer data
	erc20Data, err := cs.makeERC20TransferData(to, TRANSFER, value)
	if err != nil {
		return fmt.Sprintf("Server make ERC20 transfer data error,Err=[%v]", err), false
	}
	//build erc20 transaction
	tx := types.NewTransaction(nonoce, common.HexToAddress(toContractAddr), big.NewInt(0), gasLimit, gasPrice, erc20Data)
	return cs.sign(tx, privKey)
}

/*
bsc and bsc erc20 token sign method
*/
func (cs *BscService) sign(tx *types.Transaction, privKey *ecdsa.PrivateKey) (string, bool) {

	signTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(conf.Config.BscCfg.NetWorkId)), privKey)
	if err != nil {
		return fmt.Sprintf("Server types.SignTx error,Err=[%v]", err), false
	}
	b, err := rlp.EncodeToBytes(signTx)
	if err != nil {
		return fmt.Sprintf("Server rlp.EncodeToBytes error,Err=[%v]", err), false
	}
	return "0x" + hex.EncodeToString(b), true
}

/*
维护地址的nonce值
*/
func (cs *BscService) aegisNonce2(address string) (int64, error) {
	//1. 先获取该地址链上的nonce值
	nonce := cs.getNonce("eth_getTransactionCount", address)
	if nonce < 0 {
		logrus.Errorf("%s 获取链上的nonce值错误", address)
	}
	// 2. 判断内存中是否有这个地址的nonce
	value, ok := cs.nonceCtl.Load(address)
	if !ok {
		// 2.1  如果内存中没有，直接返回链上的nonce
		return nonce, nil
	}
	if value == nil {
		return -1, fmt.Errorf(" %s do not find any value in map", address)
	}
	data, err := json.Marshal(value)
	if err != nil {
		return -1, err
	}
	var bnd map[string]int64
	err = json.Unmarshal(data, &bnd)
	if err != nil {
		return -1, err
	}
	if len(bnd) > 30 {
		return -1, fmt.Errorf("%s address pending tx is big than 30", address)
	}
	var total, unpending int
	total = len(bnd)
	for k, v := range bnd {
		//判断是否处于pending状态
		if !cs.isPendingStatus(k) {
			unpending++
			delete(bnd, k)
			continue
		}
		if v > nonce {
			nonce = v
		}
	}
	logrus.Printf("%s 总共pending数量为：%d，unpending数量为： %d", address, total, unpending)
	cs.nonceCtl.Store(address, bnd)
	return nonce, nil
}

func (cs *BscService) aegisNonce(address string) (int64, error) {
	//1. 先获取该地址链上的nonce值
	nonce := cs.getNonce("eth_getTransactionCount", address)
	if nonce < 0 {
		logrus.Errorf("%s 获取链上的nonce值错误", address)
		return 0, fmt.Errorf("%s 获取链上的nonce值错误", address)
	}
	// 判断nonce池中的nonce是否大于当前的nonce
	value, ok := cs.noncePool.Load(address)
	if !ok {
		return nonce, nil
	}
	n := value.(int64) //上一笔使用的nonce
	if n == nonce {    //如果上一笔使用的等于链上的nonce，那么就使用内存的nonce
		nonce = n + 1
	} else if n > nonce {
		if n-nonce > 30 {
			return 0, fmt.Errorf("pending tx is big than 30")
		}
		nonce = n + 1
	}
	return nonce, nil
}

func (cs *BscService) aegisPendingNonce(address string) (int64, error) {
	//1. 先获取该地址链上的nonce值
	nonce := cs.getPendingNonce("eth_getTransactionCount", address)
	if nonce < 0 {
		logrus.Errorf("%s 获取链上的nonce值错误", address)
		return 0, fmt.Errorf("%s 获取链上的nonce值错误", address)
	}
	//// 判断nonce池中的nonce是否大于当前的nonce
	//value, ok := cs.noncePool.Load(address)
	//if !ok {
	//	return nonce, nil
	//}
	//n := value.(int64) //上一笔使用的nonce
	//if n == nonce {    //如果上一笔使用的等于链上的nonce，那么就使用内存的nonce
	//	nonce = n + 1
	//} else if n > nonce {
	//	if n-nonce > 30 {
	//		return 0, fmt.Errorf("pending tx is big than 30")
	//	}
	//	nonce = n + 1
	//}
	return nonce, nil
}

func (cs *BscService) isPendingStatus(txid string) bool {
	data, err := cs.client.SendRequest("eth_getTransactionReceipt", []interface{}{txid})
	if err != nil {
		return true
	}
	//pending tx
	if len(data) == 0 {
		return true
	}
	return false
}

func (cs *BscService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	var (
		data []byte
		err  error
	)
	if req.ContractAddress == "" {
		//获取主链
		data, err = cs.client.SendRequest("eth_getBalance", []interface{}{req.Address, "latest"})
	} else {
		//获取合约地址的金额
		erc20Params := make(map[string]interface{})
		erc20Params["from"] = req.Address
		erc20Params["to"] = req.ContractAddress
		reqData := fmt.Sprintf("0x70a08231%064s", req.Address[2:])
		erc20Params["data"] = reqData
		data, err = cs.client.SendRequest("eth_call", []interface{}{erc20Params, "latest"})
	}
	if err != nil {
		return nil, fmt.Errorf("get %s address(%s) balance error: %v", req.CoinName, req.Address, err)
	}
	amtHex := string(data)
	amtInt, err := util.ParseBigInt(amtHex)
	if err != nil {
		return nil, fmt.Errorf("parse amount(%s) to big int error: %v", amtHex, err)
	}
	amtStr := decimal.NewFromBigInt(amtInt, 0).String()
	return amtStr, nil
}

func (cs *BscService) ValidAddress(address string) error {
	if !common.IsHexAddress(address) {
		return errors.New("valid bsc address error")
	}
	return nil
}

/*
添加动态获取手续费的接口
*/

func (cs *BscService) confirmFeeBsc(feeLevel int64) int64 {
	var basePrice int64
	p, err := cs.getBuildTxParams("eth_gasPrice", []interface{}{})
	if err != nil || p <= 0 {
		return conf.Config.BscCfg.GasPrice
	} else {
		basePrice = p
	}
	var optimalPrice int64
	if feeLevel <= 0 {
		optimalPrice = basePrice * 2
	} else {
		optimalPrice = basePrice + (3 * GWei)
	}
	if optimalPrice < minGasPrice { //最小1Gwei
		optimalPrice = minGasPrice
	}
	if optimalPrice > MaxGasPrice {
		optimalPrice = MaxGasPrice //最大50Gwei
	}
	logrus.Infof("feelevel:%d,basePrice：%d，使用gas:%d", feeLevel, basePrice, optimalPrice)
	return optimalPrice
}
