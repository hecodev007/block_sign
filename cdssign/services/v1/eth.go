package v1

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eth-sign/conf"
	"github.com/eth-sign/model"
	"github.com/eth-sign/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
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
type CdsService struct {
	*BaseService
	client              *util.RpcClient
	nonceCtl, noncePool sync.Map
}

const (
	TRANSFER = "0xa9059cbb"
)

var GWei = big.NewInt(1000000000)

func minGasPriceLimit() *big.Int {
	confVal := big.NewInt(conf.Config.EthCfg.MinGasPriceGwei)
	return confVal.Mul(confVal, GWei)
}

func maxGasPriceLimit() *big.Int {
	confVal := big.NewInt(conf.Config.EthCfg.MaxGasPriceGwei)
	return confVal.Mul(confVal, GWei)
}

/*
初始化币种服务
	注意：
		方法接受者： BaseService
		方法命名： 币种大写 + Service
*/
func (bs *BaseService) CDSService() *CdsService {
	cs := new(CdsService)
	cs.BaseService = bs
	// 初始化连接
	client := util.New(conf.Config.EthCfg.NodeUrl, conf.Config.EthCfg.User, conf.Config.EthCfg.Password)
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
func (cs *CdsService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return cs.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, cs.createAddressInfo)
	}
	return cs.BaseService.createAddress(req, cs.createAddressInfo)
}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (cs *CdsService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {

	_, err := cs.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, cs.createAddressInfo)
	return err
}

/*
签名服务
*/
func (cs *CdsService) SignService(req *model.ReqSignParams) (interface{}, error) {
	reqData, err := json.Marshal(req.Data)
	if err != nil {
		return nil, err
	}
	var tp model.EthSignParams
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
		gasPrice = big.NewInt(conf.Config.EthCfg.GasPrice)
	} else {
		gasPrice = big.NewInt(tp.GasPrice)
	}
	var gasLimit int64
	if tp.GasLimit <= 0 {
		gasLimit = conf.Config.EthCfg.GasLimit
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
		// 合约 转账
		signRes, isSignOk = cs.ethTokenSign(tp.ToAddress, tp.ContractAddress, nonce, uint64(gasLimit), amount, gasPrice, privKey)
	} else {
		// 转账HOO
		signRes, isSignOk = cs.ethSign(tp.ToAddress, nonce, uint64(gasLimit), amount, gasPrice, privKey)
	}
	if !isSignOk {
		return nil, fmt.Errorf("ETH sign error,Err=[%s]", signRes)
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
func (cs *CdsService) TransferService(req interface{}) (interface{}, error) {
	var (
		tp  model.EthTransferParams
		err error
	)
	if err := cs.BaseService.parseData(req, &tp); err != nil {
		return nil, err
	}
	if &tp == nil {
		return nil, errors.New("transfer params is null")
	}

	nonce := uint64(0)
	if tp.Latest {
		// 如果是`latest`需要本地维护
		nonce, err = cs.aegisNonce(tp.FromAddress)
		if err != nil {
			return nil, err
		}
	} else {
		// 如果是`pending`直接获取链上nonce，不需要本地维护
		nonce = cs.getNonce("eth_getTransactionCount", tp.FromAddress, "pending")
	}

	result, err := cs.processTransfer(tp, nonce)
	if err == nil && tp.Latest {
		// 如果交易已广播成功，并且nonce值是使用`latest`获取
		// 那么久需要本地维护nonce
		cs.noncePool.Store(tp.FromAddress, nonce)
	}
	return result, err
}

func (cs *CdsService) TransferWithNonceService(req interface{}) (interface{}, error) {
	var tp model.EthTransferWithNonceParams
	if err := cs.BaseService.parseData(req, &tp); err != nil {
		return nil, err
	}
	if &tp == nil {
		return nil, errors.New("transfer params is null")
	}
	// 指定nonce值
	return cs.processTransfer(tp.EthTransferParams, tp.Nonce)
}

func (cs *CdsService) processTransfer(tp model.EthTransferParams, nonce uint64) (interface{}, error) {
	if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	}
	if cs.client == nil {
		client := util.New(conf.Config.EthCfg.NodeUrl, conf.Config.EthCfg.User, conf.Config.EthCfg.Password)
		cs.client = client
	}

	logrus.Infof("使用的nonce值: %d", nonce)

	var amount *big.Int
	toAmount, err := decimal.NewFromString(tp.Amount)
	if err != nil {
		return nil, fmt.Errorf("parse amount error,err=%v", err)
	}
	gasLimit := conf.Config.EthCfg.GasLimit
	gasPrice := cs.confirmFee()
	logrus.Infof("实际gas price : %d", gasPrice)
	amount = toAmount.BigInt()

	toAddress := common.HexToAddress(tp.ToAddress)
	logrus.Printf("出账金额为: %s, gasPrice: %s, Nonce: %d", amount.String(), gasPrice.String(), nonce)
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
		// 合约转账
		signRes, isSignOk = cs.ethTokenSign(tp.ToAddress, tp.ContractAddress, nonce, uint64(gasLimit), amount, gasPrice, privKey)
	} else {
		// 主链币转账
		signRes, isSignOk = cs.ethSign(tp.ToAddress, nonce, uint64(gasLimit), amount, gasPrice, privKey)
	}
	if !isSignOk {
		return nil, fmt.Errorf("ETH sign error,Err=[%s]", signRes)
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

	logrus.Printf("成功出账一笔交易： txid=%s", string(res))
	return string(res), nil
}

func (cs *CdsService) getBuildTxParams(method string, params []interface{}) (*big.Int, error) {
	res, err := cs.client.SendRequest(method, params)
	if err != nil {
		logrus.Errorf("rpc send error,Err=%v", err)
		return big.NewInt(-1), err
	}
	if res == nil {
		return big.NewInt(-1), nil
	}
	value := util.HexToDec(string(res))
	return value, nil
}

func (cs *CdsService) getNonce(method, address, tag string) uint64 {
	params := []interface{}{address, tag}
	nonce, _ := cs.getBuildTxParams(method, params)
	return nonce.Uint64()
}

/*
创建地址实体方法
*/
func (cs *CdsService) createAddressInfo() (util.AddrInfo, error) {
	privkey, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	var (
		addrInfo util.AddrInfo
		address  string
	)
	// 避免priv的len不是32
	if len(privkey.D.Bytes()) != 32 {
		for true {
			privkey, err = crypto.GenerateKey()
			if err != nil {
				// if have some error ,cut this exe
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
func (cs *CdsService) makeERC20TransferData(toAddress, transfer_method_id string, amount *big.Int) ([]byte, error) {
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
ETH sign to build transaction and sign
*/
func (cs *CdsService) ethSign(to string, nonce, gasLimit uint64, value, gasPrice *big.Int, privKey *ecdsa.PrivateKey) (string, bool) {
	// build transaction
	tx := types.NewTransaction(nonce, common.HexToAddress(to), value, gasLimit, gasPrice, nil)
	// to do eth sign
	return cs.sign(tx, privKey)

}

/*
erc20 token sign to build transaction and sign
*/
func (cs *CdsService) ethTokenSign(to, toContractAddr string, nonoce, gasLimit uint64, value, gasPrice *big.Int, privKey *ecdsa.PrivateKey) (string, bool) {
	// make erc20 tranfer data
	erc20Data, err := cs.makeERC20TransferData(to, TRANSFER, value)
	if err != nil {
		return fmt.Sprintf("Server make ERC20 transfer data error,Err=[%v]", err), false
	}
	// build erc20 transaction
	tx := types.NewTransaction(nonoce, common.HexToAddress(toContractAddr), big.NewInt(0), gasLimit, gasPrice, erc20Data)
	return cs.sign(tx, privKey)
}

/*
ETH and erc20 token sign method
*/
func (cs *CdsService) sign(tx *types.Transaction, privKey *ecdsa.PrivateKey) (string, bool) {
	signTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(conf.Config.EthCfg.NetWorkId)), privKey)
	if err != nil {
		return fmt.Sprintf("Server types.SignTx error,Err=[%v]", err), false
	}
	b, err := rlp.EncodeToBytes(signTx)
	if err != nil {
		return fmt.Sprintf("Server rlp.EncodeToBytes error,Err=[%v]", err), false
	}
	return "0x" + hex.EncodeToString(b), true
}

func (cs *CdsService) sign2(tx *types.Transaction, privKey *ecdsa.PrivateKey) (string, bool) {
	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, privKey)
	b, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return fmt.Sprintf("Server rlp.EncodeToBytes error,Err=[%v]", err), false
	}
	return "0x" + hex.EncodeToString(b), true
}

func (cs *CdsService) aegisNonce(address string) (uint64, error) {
	// 1. 先获取该地址链上的nonce值
	nonce := cs.getNonce("eth_getTransactionCount", address, "latest")
	if nonce < 0 {
		logrus.Errorf("%s 获取链上的nonce值错误", address)
		return 0, fmt.Errorf("%s 获取链上的nonce值错误", address)
	}
	// 判断nonce池中的nonce是否大于当前的nonce
	value, ok := cs.noncePool.Load(address)
	if !ok {
		return nonce, nil
	}
	n := value.(uint64) // 上一笔使用的nonce
	if n == nonce {     // 如果上一笔使用的等于链上的nonce，那么就使用内存的nonce
		nonce = n + 1
	} else if n > nonce {
		if n-nonce > 30 {
			return 0, fmt.Errorf("pending tx is big than 30")
		}
		nonce = n + 1
	}
	return nonce, nil
}
func (cs *CdsService) isPendingStatus(txid string) bool {
	data, err := cs.client.SendRequest("eth_getTransactionReceipt", []interface{}{txid})
	if err != nil {
		return true
	}
	// pending tx
	if len(data) == 0 {
		return true
	}
	return false
}

func (cs *CdsService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	var (
		data []byte
		err  error
	)
	if req.ContractAddress == "" {
		// 获取主链
		data, err = cs.client.SendRequest("eth_getBalance", []interface{}{req.Address, "latest"})
	} else {
		// 获取合约地址的金额
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

func (cs *CdsService) ValidAddress(address string) error {
	if !common.IsHexAddress(address) {
		return errors.New("valid ETH address error")
	}
	return nil
}
