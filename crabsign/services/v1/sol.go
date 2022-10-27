package v1

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JFJun/solana-go/account"
	"github.com/JFJun/solana-go/rpc"
	"github.com/JFJun/solana-go/transaction"
	"github.com/btcsuite/btcutil/base58"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"wallet-sign/conf"
	"wallet-sign/model"
	"wallet-sign/util"
)

/*
service模板
*/

/*
币种服务结构体
*/
type SolService struct {
	*BaseService
	client *rpc.RpcClient
}

/*
初始化币种服务
	注意：
		方法接受者： BaseService
		方法命名： 币种大写 + Service
*/
func (bs *BaseService) SOLService() *SolService {
	cs := new(SolService)
	cs.BaseService = bs
	cs.client = rpc.New(conf.Config.SolCfg.NodeUrl, conf.Config.SolCfg.User, conf.Config.SolCfg.Password)
	//初始化连接
	return cs
}
func (cs *SolService) reConnect() error {
	if cs.client == nil {
		client := rpc.New(conf.Config.SolCfg.NodeUrl, conf.Config.SolCfg.User, conf.Config.SolCfg.Password)
		if client == nil {
			return errors.New("reconnect error")
		}
		cs.client = client
	}
	return nil
}

/*
接口创建地址服务
	无需改动
*/
func (cs *SolService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return cs.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, cs.createAddressInfo)
	}
	return cs.BaseService.createAddress(req, cs.createAddressInfo)
}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (cs *SolService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	fmt.Println("start create cds address")
	_, err := cs.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, cs.createAddressInfo)
	return err
}
func (cs *SolService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	data, err := cs.client.SendRequest("getBalance", []interface{}{req.Address})
	if err != nil {
		return nil, fmt.Errorf("%s get balance error,err=%v", req.Address, err)
	}
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal balance resopose error,err=%v", err)
	}
	balance := decimal.NewFromFloat(result["value"].(float64))
	return balance.String(), nil
}

/*
签名服务
*/
func (cs *SolService) SignService(req *model.ReqSignParams) (interface{}, error) {
	return nil, nil
}

func (cs *SolService) ValidAddress(address string) error {
	data := base58.Decode(address)
	if len(data) == 0 {
		return errors.New("bas58 decode data error")
	}
	if len(data) != 32 {
		return errors.New("public key length is not equal 32")
	}
	return nil
}

/*
热钱包出账服务
*/
func (cs *SolService) TransferService(req interface{}) (interface{}, error) {
	var tp model.SolTransferParams
	if err := cs.BaseService.parseData(req, &tp); err != nil {
		return nil, err
	}
	if &tp == nil {
		return nil, errors.New("transfer params is null")
	}
	if tp.FromAddress == "" || tp.Amount == "" || tp.ToAddress == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]",
			tp.FromAddress, tp.ToAddress, tp.Amount)
	}

	// todo 做金额限制
	amount, err := decimal.NewFromString(tp.Amount)
	if err != nil {
		return nil, fmt.Errorf("parse amount error,err=%v", err)
	}
	data, err1 := cs.client.SendRequest("getBalance", []interface{}{tp.FromAddress})
	if err1 != nil {
		return nil, fmt.Errorf("%s get balance error,err=%v", tp.FromAddress, err1)
	}
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal balance resopose error,err=%v", err)
	}
	//balance:=decimal.NewFromFloat(result["value"].(float64))
	//if balance.LessThan(amount) {
	//	return nil,fmt.Errorf("%s amount is not engouth,Online amount=[%d],Transfer amount=[%d]",tp.FromAddress,
	//		balance.IntPart(),amount.IntPart())
	//}
	//构建交易参数
	tps := transaction.TransferParams{
		From:   tp.FromAddress,
		To:     tp.ToAddress,
		Amount: amount.BigInt(),
	}
	transfer, err := transaction.NewTransfer(tps)
	if err != nil {
		return nil, fmt.Errorf("build transfer error,err=%v", err)
	}

	err = cs.reConnect()
	if err != nil {
		return nil, err
	}
	data, err = cs.client.SendRequest("getRecentBlockhash", nil)
	if err != nil {
		return nil, fmt.Errorf("rpc get recent block hash error,err=%v", err)
	}

	recentBlockHash, err1 := model.DecodeSolRecentBlockHash(data)
	if err1 != nil {
		return nil, err1
	}

	//创建交易
	tx := transaction.NewTransaction(recentBlockHash)
	tx.SetInstructions(transfer)
	//查找私钥
	hexPriv, err2 := cs.GetKeyByAddress(tp.FromAddress)
	if err2 != nil {
		return nil, fmt.Errorf("get [%s] private key error,err=[%v]", tp.FromAddress, err2)
	}
	seed, err3 := hex.DecodeString(hexPriv)
	if err3 != nil {
		return nil, fmt.Errorf("hex decode private key error,err=%v", err3)
	}
	acc := account.NewAccountBySecret(seed)
	var accounts []*account.Account
	accounts = append(accounts, acc)

	//签名交易
	err4 := tx.Sign(accounts)
	if err4 != nil {
		return nil, fmt.Errorf("sign error,err=%v", err4)
	}
	//序列化交易
	wireTx, err5 := tx.Serialize()
	if err5 != nil {
		return nil, fmt.Errorf("serialize transaction error,err=%v", err5)
	}
	b58Tx := base58.Encode(wireTx)

	//发送交易
	sendData, err6 := cs.client.SendRequest("sendTransaction", []interface{}{b58Tx})
	if err6 != nil {
		logrus.Errorf("发送数据为：%s", b58Tx)
		return nil, fmt.Errorf("send transaction error,err=%v", err6)
	}
	return string(sendData), nil
}

/*
创建地址实体方法
*/
func (cs *SolService) createAddressInfo() (util.AddrInfo, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	var (
		addrInfo util.AddrInfo
	)
	address := base58.Encode(pub)
	wif := hex.EncodeToString(priv.Seed())
	addrInfo.PrivKey = wif
	addrInfo.Address = address
	return addrInfo, nil
}
