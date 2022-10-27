package services

import (
	"brisesign/conf"
	"brisesign/model"
	"brisesign/redis"
	"brisesign/util"
	"brisesign/util2"
	"context"
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
	"github.com/shopspring/decimal"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ChainService struct {
	*BaseService
	client        *util.RpcClient
	client2       *util2.RpcClient
	noncePool     sync.Map
	orderKeeper   *OrderKeeper
	orderCallback *OrderCallback
	contractGas   map[string]int64
	mtx           sync.RWMutex
	addrTimeMap   map[string]int64
}

var (
	GWei = big.NewInt(1000000000)
)

const (
	transferCode = "0xa9059cbb"
	// 执行签名的时间间隔
	processSignInterval = time.Millisecond * 200
)

func minGasPriceLimit() *big.Int {
	confVal := big.NewInt(conf.Config.ChainCfg.MinGasPriceGwei)
	return confVal.Mul(confVal, GWei)
}

func maxGasPriceLimit() *big.Int {
	confVal := big.NewInt(conf.Config.ChainCfg.MaxGasPriceGwei)
	return confVal.Mul(confVal, GWei)
}

/*
初始化币种服务
	注意：
		方法接受者： BaseService
		方法命名： 币种大写 + Service
*/
func (bs *BaseService) ChainService(ctx context.Context) *ChainService {
	cs := new(ChainService)
	cs.BaseService = bs
	// 初始化连接
	client := util.New(conf.Config.ChainCfg.NodeUrl, conf.Config.ChainCfg.User, conf.Config.ChainCfg.Password)
	cs.client = client

	client2 := util2.New(conf.Config.ChainCfg.NodeUrl, conf.Config.ChainCfg.User, conf.Config.ChainCfg.Password)
	cs.client2 = client2

	// 新增nonce维护池
	//cs.orderKeeper = NewOrderKeeper(ctx)
	//cs.orderCallback = NewOrderCallback(ctx)
	cs.contractGas = map[string]int64{}
	cs.addrTimeMap = map[string]int64{}

	if conf.Config.Gas.Special != "" {
		gasArr := strings.Split(conf.Config.Gas.Special, ",")
		for _, s := range gasArr {
			if s == "" {
				continue
			}
			contractGasArr := strings.Split(s, "-")

			gasLimit, err := strconv.ParseInt(contractGasArr[1], 10, 64)
			if err != nil {
				log.Printf("gas.special %s 转换为int64失败 %s", contractGasArr[1], err.Error())
				continue
			}
			cs.contractGas[strings.ToLower(contractGasArr[0])] = gasLimit
		}
	}

	//go cs.loopSign(ctx)
	return cs
}

func (s *ChainService) loopSign(ctx context.Context) {
	log.Println("开始执行 loopSign")
	timer := time.NewTimer(processSignInterval)

loop:
	for {
		select {
		case <-timer.C:
			s.processSign()
			timer.Reset(processSignInterval)
		case <-ctx.Done():
			log.Println("收到Done信号，停止timer，终止循环")
			timer.Stop()
			break loop
		}
	}
	log.Println("退出 loopSign")
}

func (s *ChainService) processSign() {
	// 从待处理列表弹出最早的一笔订单
	order, err := s.orderKeeper.pop()
	if order == nil {
		// 没有待处理订单
		return
	}
	log.Printf("从redis待执行列表获取到订单:%s", order.OuterOrderNo)
	if err != nil {
		log.Printf("从redis待执行列表(缓存)弹出订单失败:%v", err)
		return
	}

	trx, err := s.signAndSendRawTransaction(order)
	if err != nil {
		// 签名失败
		// 将该订单从防重放的缓存中移除
		// 等待blockchains那边再次请求时会继续处理此订单
		s.orderKeeper.removeFromAgainstReplay(order.OuterOrderNo)
	} else {
		// 签名成功
		log.Printf("订单:%s 签名完成", order.OuterOrderNo)
	}

	// 使用异步回调blockchains接口，告知签名的执行结果
	// 当回调响应失败（HTTP.STATUS != 200)时，会将数据保存到redis缓存
	// 然后设置一个`sendTime`表示下一次尝试重新回调的时间
	// 失败会一直重复上述操作直到回调成功
	go s.orderCallback.Send(order.OuterOrderNo, order.OrderHotId, trx, err)
	log.Printf("订单:%s 处理完毕", order.OuterOrderNo)
}

/*
接口创建地址服务
	无需改动
*/
func (s *ChainService) CreateAddressService(req *model.ReqCreateAddressParamsV2) (*model.RespCreateAddressParams, error) {
	if req.Count == 0 {
		req.Count = 1000
	}
	if req.BatchNo == "" {
		req.BatchNo = util.GetTimeNowStr()
	}
	var (
		result *model.RespCreateAddressParams
		err    error
	)
	if conf.Config.IsStartThread {
		result, err = s.BaseService.multiThreadCreateAddress(req.Count, req.CoinCode, req.Mch, req.BatchNo, s.createAddressInfo)
	} else {
		result, err = s.BaseService.createAddress(req, s.createAddressInfo)
	}
	if err == nil {
		log.Printf("CreateAddressService 完成，共生成 %d 个地址，准备重新加载地址", len(result.Address))
		s.InitKeyMap()
		log.Println("重新加载地址完成")
	}
	return result, err

}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (s *ChainService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	_, err := s.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, s.createAddressInfo)
	return err
}

/*
签名服务
*/
func (s *ChainService) SignService(req *model.ReqSignParams) (interface{}, error) {
	reqData, err := json.Marshal(req.Data)
	if err != nil {
		return nil, err
	}
	var tp model.SignParams
	if err := json.Unmarshal(reqData, &tp); err != nil {
		return nil, err
	}

	if err := s.BaseService.parseData(req, &tp); err != nil {
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
		gasPrice = big.NewInt(conf.Config.ChainCfg.GasPrice)
	} else {
		gasPrice = big.NewInt(tp.GasPrice)
	}
	var gasLimit int64
	if tp.GasLimit <= 0 {
		gasLimit = conf.Config.ChainCfg.GasLimit
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
	log.Printf("出账金额为：%s,gasPrice：%s,Nonce: %d", amount.String(), gasPrice.String(), nonce)
	if strings.Compare(util.Del0xToLower(toAddress.String()), util.Del0xToLower(tp.ToAddress[:])) != 0 {
		return nil, fmt.Errorf("to address is not equal,address1=[%s],address2=[%s]", util.Del0xToLower(toAddress.String()),
			util.Del0xToLower(tp.ToAddress[:]))
	}
	from := tp.FromAddress
	hexPrivateKey, err := s.BaseService.addressOrPublicKeyToPrivate(from)
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
		signRes, isSignOk = s.chainTokenSign(tp.ToAddress, tp.ContractAddress, uint64(nonce), uint64(gasLimit), amount, gasPrice, privKey)
	} else {
		// 转账HOO
		signRes, isSignOk = s.chainSign(tp.ToAddress, uint64(nonce), uint64(gasLimit), amount, gasPrice, privKey)
	}
	if !isSignOk {
		return nil, fmt.Errorf("sign error,Err=[%s]", signRes)
	}
	hexTx := signRes
	if !strings.HasPrefix(hexTx, "0x") {
		hexTx = "0x" + hexTx
	}
	log.Printf("rawTx: %s\n", hexTx)
	return hexTx, nil
}

/*
热钱包出账服务
*/
func (s *ChainService) TransferService(req interface{}) ([]byte, error) {
	var tp model.TransferParams
	if err := s.BaseService.parseData(req, &tp); err != nil {
		return nil, err
	}
	if &tp == nil {
		return nil, errors.New("transfer params is null")
	}
	if tp.OuterOrderNo == "" {
		return nil, errors.New("outer_order_no require")
	}
	//if tp.Mac == "" {
	//	return errors.New("mac require")
	//}
	if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	}

	// 校验消息认证码是否合法
	//if tp.ToHMAC() != tp.Mac {
	//	return errors.New("invalid MAC")
	//}

	//if s.client == nil {
	//	client := util.New(conf.Config.ChainCfg.NodeUrl, conf.Config.ChainCfg.User, conf.Config.ChainCfg.Password)
	//	s.client = client
	//}

	//return s.orderKeeper.pushIfNotExist(&tp)
	return s.signAndSendRawTransaction(&tp)
}

func (s *ChainService) signAndSendRawTransaction(order *model.TransferParams) ([]byte, error) {
	log.Println("准备出账")
	log.Println("检查出账余额开始 ")
	if order.ContractAddress != "" {
		// 非主链币需要检查余额是否足够
		// 主链币不需要检查，因为如果主链币余额不足，在广播的时候回直接返回异常
		if err := s.validBalance(order.FromAddress, order.ContractAddress, order.Amount); err != nil {
			log.Printf("检查出账余额不通过：%v", err)
			return nil, err
		}
	}
	log.Println("检查出账余额结束 ")

	log.Println("获取Nonce开始 ")
	nonce, err := s.aegisNonce(order.FromAddress)
	log.Println("获取Nonce结束 ")

	if err != nil {
		return nil, err
	}
	var amount *big.Int
	toAmount, err := decimal.NewFromString(order.Amount)
	if err != nil {
		return nil, fmt.Errorf("parse amount error,err=%v", err)
	}
	gasLimit := conf.Config.ChainCfg.GasLimit
	var (
		gasPrice *big.Int
	)
	log.Println("计算Fee开始 ")
	gasPrice = s.confirmFee()
	log.Println("计算Fee结束 ")

	log.Printf("实际gas price : %d", gasPrice)
	amount = toAmount.BigInt()
	toAddress := common.HexToAddress(order.ToAddress)
	log.Printf("出账金额为： %s,gasPrice： %d,Nonce: %d", amount.String(), gasPrice, nonce)
	if strings.Compare(util.Del0xToLower(toAddress.String()), util.Del0xToLower(order.ToAddress[:])) != 0 {
		return nil, fmt.Errorf("to address is not equal,address1=[%s],address2=[%s]", util.Del0xToLower(toAddress.String()),
			util.Del0xToLower(order.ToAddress[:]))
	}
	from := order.FromAddress

	log.Println("获取公钥开始 ")
	hexPrivateKey, err3 := s.BaseService.addressOrPublicKeyToPrivate(from)
	if err3 != nil {
		return nil, fmt.Errorf("get private key error,Err=%v", err3)
	}
	log.Println("获取公钥结束 ")

	privKey, err2 := crypto.HexToECDSA(hexPrivateKey)
	if privKey == nil || err2 != nil {
		return nil, fmt.Errorf("private key is null,err=%v", err)
	}
	var (
		signRes  string
		isSignOk bool
	)
	log.Println("签名开始 ")
	if order.ContractAddress != "" {

		if v, ok := s.contractGas[strings.ToLower(order.ContractAddress)]; ok {
			log.Printf("特殊合约(%s)的gasLimit 从配置中获取: %d", order.ContractAddress, v)
			gasLimit = v
		}
		// 合约 转账
		// safemoon oks safe
		//if strings.ToLower(order.ContractAddress) == "0x3ad9594151886ce8538c1ff615efa2385a8c3a88" ||
		//	strings.ToLower(order.ContractAddress) == "0x8076c74c5e3f5852037f31ff0093eeb8c8add8d3" ||
		//	strings.ToLower(order.ContractAddress) == "0xf5581dfefd8fb0e4aec526be659cfab1f8c781da" ||
		//	strings.ToLower(order.ContractAddress) == "0x18acf236eb40c0d4824fb8f2582ebbecd325ef6a" {
		//	gasLimit = 1000000
		//	log.Printf("特殊交易，修改limit为: %d", gasLimit)
		//} else if strings.ToLower(order.ContractAddress) == "0x5066c68cae3b9bdacd6a1a37c90f2d1723559d18" {
		//	gasLimit = 1300000
		//	log.Printf("特殊交易，修改limit为: %d", gasLimit)
		//}
		signRes, isSignOk = s.chainTokenSign(order.ToAddress, order.ContractAddress, nonce, uint64(gasLimit), amount, gasPrice, privKey)
	} else {
		signRes, isSignOk = s.chainSign(order.ToAddress, nonce, uint64(gasLimit), amount, gasPrice, privKey)
	}
	log.Println("签名结束 ")

	if !isSignOk {
		return nil, fmt.Errorf("sign error,Err=[%s]", signRes)
	}
	hexTx := signRes
	if !strings.HasPrefix(hexTx, "0x") {
		hexTx = "0x" + hexTx
	}
	if order.OuterOrderNo != "" {
		log.Printf("OuterOrderNo 不为空 = %s", order.OuterOrderNo)
		cache, err := redis.Client.Get(redis.GetBroadcastOuterOrderNoKey(order.OuterOrderNo))
		if err != nil {
			log.Printf("从redis获取广播订单KEY失败:%v", err)
		} else {
			if cache != "" {
				return nil, fmt.Errorf("订单: %s 已被广播，再次广播会造成重复出账", order.OuterOrderNo)
			}
		}
	}
	log.Printf("rawTx: %s", hexTx)
	//log.Println("广播开始 ")
	res, err4 := s.client.SendRequest("eth_sendRawTransaction", []interface{}{hexTx})
	if err4 != nil {
		return nil, fmt.Errorf("send transaction error,Err=[%v]", err4)
	}
	if res == nil {
		return nil, errors.New("send transaction error,response null")
	}
	//log.Println("广播结束 ")

	// 本地维护nonce
	s.noncePool.Store(order.FromAddress, nonce) // 存储当前使用的nonce

	if order.OuterOrderNo != "" {
		if err = redis.Client.Set(redis.GetBroadcastOuterOrderNoKey(order.OuterOrderNo), order.OuterOrderNo, time.Hour*24); err != nil {
			log.Printf("广播订单存入redis失败: %v", err)
		} else {
			log.Printf("广播订单=%s 存入redis成功", order.OuterOrderNo)
		}
	}
	log.Printf("成功出账一笔交易： txId=%s", string(res))
	//log.Println("广播结束 ")
	return res, nil
}

func (s *ChainService) getBuildTxParams(method string, params []interface{}) (*big.Int, error) {
	res, err := s.client.SendRequest(method, params)
	if err != nil {
		log.Printf("rpc send error,Err=%v", err)
		return big.NewInt(-1), err
	}
	if res == nil {
		return big.NewInt(-1), nil
	}
	value := util.HexToDec(string(res))
	return value, nil
}

func (s *ChainService) getNonce(method, address string) uint64 {
	// params := []interface{}{address, "latest"}
	params := []interface{}{address, "pending"}
	nonce, _ := s.getBuildTxParams(method, params)
	return nonce.Uint64() // nonce值使用64bit来装已经足够
}

/*
创建地址实体方法
*/
func (s *ChainService) createAddressInfo() (util.AddrInfo, error) {
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
func (s *ChainService) makeERC20TransferData(toAddress, transfer_method_id string, amount *big.Int) ([]byte, error) {
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
sign to build transaction and sign
*/
func (s *ChainService) chainSign(to string, nonce, gasLimit uint64, value, gasPrice *big.Int, privKey *ecdsa.PrivateKey) (string, bool) {
	// build transaction
	tx := types.NewTransaction(nonce, common.HexToAddress(to), value, gasLimit,
		gasPrice, nil)
	// to do eth sign
	return s.sign(tx, privKey)

}

/*
erc20 token sign to build transaction and sign
*/
func (s *ChainService) chainTokenSign(to, toContractAddr string, nonoce, gasLimit uint64, value, gasPrice *big.Int, privKey *ecdsa.PrivateKey) (string, bool) {
	// make erc20 tranfer data
	erc20Data, err := s.makeERC20TransferData(to, transferCode, value)
	if err != nil {
		return fmt.Sprintf("Server make ERC20 transfer data error,Err=[%v]", err), false
	}
	// build erc20 transaction
	tx := types.NewTransaction(nonoce, common.HexToAddress(toContractAddr), big.NewInt(0), gasLimit, gasPrice, erc20Data)
	return s.sign(tx, privKey)
}

/*
erc20 token sign method
*/
func (s *ChainService) sign(tx *types.Transaction, privKey *ecdsa.PrivateKey) (string, bool) {

	signTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(conf.Config.ChainCfg.NetWorkId)), privKey)

	if err != nil {
		return fmt.Sprintf("Server types.SignTx error,Err=[%v]", err), false
	}

	b, err := rlp.EncodeToBytes(signTx)
	if err != nil {
		return fmt.Sprintf("Server rlp.EncodeToBytes error,Err=[%v]", err), false
	}
	return "0x" + hex.EncodeToString(b), true
}

func (s *ChainService) aegisNonce(address string) (uint64, error) {
	// 获取该地址链上的nonce值
	nonce := s.getNonce("eth_getTransactionCount", address)
	if nonce < 0 {
		log.Printf("%s 获取链上的nonce值错误", address)
		return 0, fmt.Errorf("%s 获取链上的nonce值错误", address)
	}
	return nonce, nil
}

func (s *ChainService) isPendingStatus(txid string) bool {
	data, err := s.client.SendRequest("eth_getTransactionReceipt", []interface{}{txid})
	if err != nil {
		return true
	}
	// pending tx
	if len(data) == 0 {
		return true
	}
	return false
}

func (s *ChainService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	balance, err := s.getBalanceFromChain(req.Address, req.ContractAddress)
	if err != nil {
		return balance, err
	}
	amtStr := decimal.NewFromBigInt(balance, 0).String()
	return amtStr, nil
}

func (s *ChainService) getBalanceFromChain(addr, contractAddr string) (*big.Int, error) {
	var (
		data []byte
		err  error
	)
	if contractAddr == "" {
		// 获取主链
		data, err = s.client.SendRequest("eth_getBalance", []interface{}{addr, "latest"})
	} else {
		// 获取合约地址的金额
		erc20Params := make(map[string]interface{})
		erc20Params["from"] = addr
		erc20Params["to"] = contractAddr
		reqData := fmt.Sprintf("0x70a08231%064s", addr[2:])
		erc20Params["data"] = reqData
		data, err = s.client.SendRequest("eth_call", []interface{}{erc20Params, "latest"})
	}
	if err != nil {
		return nil, fmt.Errorf("get address(%s) contract(%s) balance error: %v", addr, contractAddr, err)
	}
	amtHex := string(data)
	amtInt, err := util.ParseBigInt(amtHex)
	if err != nil {
		return nil, fmt.Errorf("parse amount(%s) to big int error: %v", amtHex, err)
	}
	return amtInt, nil
}

func (s *ChainService) ValidAddress(address string) error {
	if !common.IsHexAddress(address) {
		return errors.New("valid address error")
	}
	return nil
}

func (s *ChainService) DelKey(orderId string) error {
	return s.orderKeeper.delProcessedKey(orderId)
}

func (s *ChainService) validBalance(addr, contractAddr, amountStr string) error {
	balance, err := s.getBalanceFromChain(addr, contractAddr)
	if err != nil {
		return err
	}

	var amount big.Int
	amount.SetString(amountStr, 10)

	if balance.Cmp(&amount) == -1 {
		// 余额不足以支付出账金额
		return fmt.Errorf("[not sufficient funds] from=%s contract=%s balance=%d amount=%d", addr, contractAddr, balance, &amount)
	}
	return nil
}

func (s *ChainService) TransferCollectService(req interface{}) (interface{}, error) {
	defer s.mtx.Unlock()
	s.mtx.Lock()
	log.Println("准备归集")
	var tp model.BscTransferParams
	if err := s.BaseService.parseData(req, &tp); err != nil {
		return nil, err
	}
	if &tp == nil {
		return nil, errors.New("transfer params is null")
	}
	if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	}
	if s.client == nil {
		client := util.New(conf.Config.ChainCfg.NodeUrl, conf.Config.ChainCfg.User, conf.Config.ChainCfg.Password)
		s.client = client
	}

	nonce := s.getNonce("eth_getTransactionCount", tp.FromAddress)

	var amount *big.Int
	toAmount, err := decimal.NewFromString(tp.Amount)
	if err != nil {
		return nil, fmt.Errorf("parse amount error,err=%v", err)
	}
	gasLimit := conf.Config.ChainCfg.GasLimit
	var (
		gasPrice *big.Int
	)

	if tp.GasLimit > 0 {
		gasLimit = tp.GasLimit
		log.Printf("使用入参传过来的gasLimit %d", gasLimit)
	} else {
		if tp.ContractAddress != "" {
			if v, ok := s.addrTimeMap[strings.ToLower(tp.FromAddress)]; ok {
				if time.Now().Unix() < v {
					log.Printf("地址 %s 归集完不到30秒，本次不能归集", tp.FromAddress)
					return nil, fmt.Errorf("address %s collect too busy", tp.FromAddress)
				}
			}
			s.addrTimeMap[strings.ToLower(tp.FromAddress)] = time.Now().Unix() + 30

			if v, ok := s.contractGas[strings.ToLower(tp.ContractAddress)]; ok {
				log.Printf("特殊合约(%s)的gasLimit 从配置中获取: %d", tp.ContractAddress, v)
				gasLimit = v
			}
		}
	}

	if tp.GasPrice > 0 {
		gasPrice = big.NewInt(tp.GasPrice)
		log.Printf("使用入参的gasPrice %d", tp.GasPrice)
	} else {
		gasPrice = big.NewInt(s.confirmFeeBsc(tp.Fee))
		log.Printf("最终 from=%s  to=%s 使用手续费为 %d", tp.FromAddress, tp.ToAddress, gasPrice.Int64())
	}

	log.Printf("实际gas price : %s", gasPrice.String())
	amount = toAmount.BigInt()
	toAddress := common.HexToAddress(tp.ToAddress)
	log.Printf("出账金额为： %s,手续费为： %d,Nonce: %d", amount.String(), gasPrice.Int64()*gasLimit, nonce)
	if strings.Compare(util.Del0xToLower(toAddress.String()), util.Del0xToLower(tp.ToAddress[:])) != 0 {
		return nil, fmt.Errorf("to address is not equal,address1=[%s],address2=[%s]", util.Del0xToLower(toAddress.String()),
			util.Del0xToLower(tp.ToAddress[:]))
	}
	from := tp.FromAddress
	hexPrivateKey, err3 := s.BaseService.addressOrPublicKeyToPrivate(from)
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
		// safemoon oks safe
		//if strings.ToLower(tp.ContractAddress) == "0x3ad9594151886ce8538c1ff615efa2385a8c3a88" ||
		//	strings.ToLower(tp.ContractAddress) == "0x8076c74c5e3f5852037f31ff0093eeb8c8add8d3" ||
		//	strings.ToLower(tp.ContractAddress) == "0x18acf236eb40c0d4824fb8f2582ebbecd325ef6a" {
		//	gasLimit = 750000
		//	log.Printf("特殊交易，修改limit为: %d", gasLimit)
		//} else if strings.ToLower(tp.ContractAddress) == "0x5066c68cae3b9bdacd6a1a37c90f2d1723559d18" {
		//	gasLimit = 1300000
		//	log.Printf("特殊交易，修改limit为: %d", gasLimit)
		//}

		transferData, _ := ERC20Transfer(tp.ToAddress, amount)
		egr := &EstimateGasRequest{
			From:     from,
			To:       tp.ToAddress,
			GasPrice: gasPrice.String(),
			Data:     transferData,
		}

		estimateGas, err := s.EstimateGas(*egr)
		log.Printf("EstimateGas估算的gas为 %d 实际使用的gas为 %d", estimateGas, gasLimit)
		if err != nil {
			return nil, err
		}
		if gasLimit < estimateGas {
			gasLimit = estimateGas
			log.Printf("使用EstimateGas估算的值 %d", gasLimit)
		}

		signRes, isSignOk = s.bscTokenSign(tp.ToAddress, tp.ContractAddress, uint64(nonce), uint64(gasLimit), amount, gasPrice, privKey)
	} else {
		//转账bsc
		signRes, isSignOk = s.bscSign(tp.ToAddress, uint64(nonce), uint64(gasLimit), amount, gasPrice, privKey)
	}
	if !isSignOk {
		return nil, fmt.Errorf("bsc sign error,Err=[%s]", signRes)
	}
	hexTx := signRes
	if !strings.HasPrefix(hexTx, "0x") {
		hexTx = "0x" + hexTx
	}
	log.Printf("rawTx: %s", hexTx)
	res, err4 := s.client2.SendRequest("eth_sendRawTransaction", []interface{}{hexTx})
	if err4 != nil {
		return nil, fmt.Errorf("send transaction error,Err=[%v]", err4)
	}
	if res == nil {
		return nil, errors.New("send transaction error,response null")
	}

	log.Printf("成功出账一笔交易： txid=%s", string(res))
	return string(res), nil
}

func (s *ChainService) confirmFeeBsc(feeLevel int64) int64 {
	optimalPrice := big.NewInt(5 * GWei.Int64())

	if optimalPrice.Int64() < minGasPriceLimit().Int64() { //最小1Gwei
		optimalPrice = big.NewInt(minGasPriceLimit().Int64())
	}
	if optimalPrice.Int64() > maxGasPriceLimit().Int64() {
		optimalPrice = big.NewInt(maxGasPriceLimit().Int64()) //最大50Gwei
	}
	return optimalPrice.Int64()
}

func ERC20Transfer(to string, amount *big.Int) (data string, err error) {
	if !isAddress(to) {
		return data, errors.New("to isn't address format")
	}
	data = fmt.Sprintf("0xa9059cbb%064s%064x", to[2:], amount)

	return data, err
}

func isAddress(address string) bool {
	bigInt := new(big.Int)
	_, ok := bigInt.SetString(address, 0)

	if !ok || len(address) != 42 {
		return false
	} else {
		return true
	}
}

type EstimateGasRequest struct {
	From     string `json:"from"`
	To       string `json:"to"`
	GasPrice string `json:"gasPrice"`
	Value    string `json:"value"`
	Data     string `json:"data"`
}

type T struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Gas      string
	GasPrice string
	Value    string
	Data     string `json:"data"`
	Nonce    int    `json:"nonce"`
}

func (s *ChainService) EstimateGas(req EstimateGasRequest) (int64, error) {
	var response string
	gp, _ := decimal.NewFromString(req.GasPrice)
	t := &T{From: req.From, To: req.To, GasPrice: fmt.Sprintf("0x%x", gp.Coefficient())}
	if req.Data == "" {
		v, _ := decimal.NewFromString(req.Value)
		t.Value = fmt.Sprintf("0x%x", v.Coefficient())
	} else {
		t.Data = req.Data
	}

	err := s.client2.CallNoAuth("eth_estimateGas", &response, t)
	if err != nil {
		return 0, err
	}
	if err != nil {
		return 0, err
	}

	return ParseInt64(response)
}

func ParseInt64(value string) (int64, error) {
	i, err := strconv.ParseInt(value, 0, 64)
	if err != nil {
		return 0, err
	}

	return int64(i), nil
}

/*
erc20 token sign to build transaction and sign
*/
func (s *ChainService) bscTokenSign(to, toContractAddr string, nonoce, gasLimit uint64, value, gasPrice *big.Int, privKey *ecdsa.PrivateKey) (string, bool) {
	//make erc20 tranfer data
	erc20Data, err := s.makeERC20TransferData(to, "0xa9059cbb", value)
	if err != nil {
		return fmt.Sprintf("Server make ERC20 transfer data error,Err=[%v]", err), false
	}
	//build erc20 transaction
	tx := types.NewTransaction(nonoce, common.HexToAddress(toContractAddr), big.NewInt(0), gasLimit, gasPrice, erc20Data)
	return s.sign(tx, privKey)
}

/*
bsc sign to build transaction and sign
*/
func (s *ChainService) bscSign(to string, nonce, gasLimit uint64, value, gasPrice *big.Int, privKey *ecdsa.PrivateKey) (string, bool) {
	//build transaction
	tx := types.NewTransaction(nonce, common.HexToAddress(to), value, gasLimit,
		gasPrice, nil)
	//to do eth sign
	return s.sign(tx, privKey)

}
