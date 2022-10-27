package v2

import (
	"encoding/json"
	"errors"
	"fmt"
	substrateConf "github.com/coldwallet-group/substrate-go/config"
	"github.com/coldwallet-group/substrate-go/rpc"
	"github.com/coldwallet-group/substrate-go/sr25519"
	"github.com/coldwallet-group/substrate-go/tx"
	"github.com/shopspring/decimal"
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
type KsmService struct {
	*BaseService
	client   *rpc.Client
	url      string
	nonceCtl map[string]int64
}

/*
初始化币种服务
	注意：
		方法接受者： BaseService
		方法命名： 币种大写 + Service
*/
func (bs *BaseService) KSMService() *KsmService {
	ks := new(KsmService)
	ks.BaseService = bs
	ks.url = conf.Config.KsmCfg.NodeUrl
	ks.nonceCtl = make(map[string]int64)
	return ks
}

/*
接口创建地址服务
	无需改动
*/
func (ks *KsmService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return ks.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, ks.createAddressInfo)
	}
	return ks.BaseService.createAddress(req, ks.createAddressInfo)
}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (ks *KsmService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	fmt.Println("start create Ksm address")
	_, err := ks.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, ks.createAddressInfo)
	return err
}

/*
签名服务
*/
func (ks *KsmService) SignService(req *model.ReqSignParams) (interface{}, error) {
	var tp model.KsmColdParams
	if err := ks.BaseService.parseData(req, &tp); err != nil {

		return nil, err
	}
	if &tp == nil {
		return nil, errors.New("transfer params is null")
	}
	if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	}
	if tp.GenesisHash == "" {
		return nil, fmt.Errorf("params is null,genesisHash=[%s]", tp.GenesisHash)
	}
	amount, err := decimal.NewFromString(tp.Amount)
	if err != nil {
		return nil, err
	}
	originTx := tx.CreateTransaction(tp.FromAddress, tp.ToAddress, uint64(amount.IntPart()), tp.Nonce, uint64(0))
	blockHash := ""
	if tp.BlockHash == "" {
		blockHash = tp.GenesisHash
	} else {
		blockHash = tp.BlockHash
	}
	originTx.SetGenesisHashAndBlockHash(tp.GenesisHash, blockHash, tp.BlockNumber)
	originTx.SetSpecVersionAndCallId(tp.SpecVersion, tp.TransactionVersion, substrateConf.CallIdKusama)
	_, message, err1 := originTx.CreateEmptyTransactionAndMessage()
	if err1 != nil {
		return nil, fmt.Errorf("create raw transaction error,err=%v", err)
	}
	//获取私钥
	privateKey, err2 := ks.addressOrPublicKeyToPrivate(tp.FromAddress)
	if err2 != nil {
		return nil, err2
	}

	sig, err3 := originTx.SignTransaction(privateKey, message)
	if err3 != nil {
		return nil, fmt.Errorf("sign error,Err=[%v]", err3)
	}
	txHex, err4 := originTx.GetSignTransaction(sig)
	if err4 != nil {
		return nil, fmt.Errorf("merge signature and tx error,Err=[%v]", err4)
	}
	return txHex, nil
}

/*
热钱包出账服务
*/
func (ks *KsmService) TransferService(req interface{}) (interface{}, error) {
	if ks.client == nil {
		//初始化客户端
		c, err := rpc.New(ks.url, "", "")
		if err != nil {
			return nil, err
		}
		ks.client = c
		fmt.Println(ks.client.TransactionVersion)
		fmt.Println(ks.client.SpecVersion)
	}
	var tp model.KsmTransferParams
	if err := ks.BaseService.parseData(req, &tp); err != nil {

		return nil, err
	}
	if &tp == nil {
		return nil, errors.New("transfer params is null")
	}
	if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	}

	//判断金额是否足够出账
	// 获取链上余额
	data, err := ks.client.GetAccountInfo(tp.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("get from address amount error,err=%v", err)
	}
	var aInfo model.AccountInfo
	if err := json.Unmarshal(data, &aInfo); err != nil {
		return nil, err
	}
	free := aInfo.Data.Free
	amount, _ := decimal.NewFromString(tp.Amount)
	balance := decimal.NewFromInt(int64(free))
	if balance.LessThanOrEqual(amount) {
		return nil, fmt.Errorf("%s is not enough amount to transfer,transferAmount=%s,actuallyAmount=%d", tp.FromAddress,
			tp.Amount, free)
	}

	originTx := tx.CreateTransaction(tp.FromAddress, tp.ToAddress, uint64(amount.IntPart()), aInfo.Nonce, uint64(0))
	originTx.SetGenesisHashAndBlockHash(ks.client.GetGenesisHash(), ks.client.GetGenesisHash(), 1)
	originTx.SetSpecVersionAndCallId(uint32(ks.client.SpecVersion), uint32(ks.client.TransactionVersion), substrateConf.CallIdKusama)
	_, message, err1 := originTx.CreateEmptyTransactionAndMessage()
	if err1 != nil {
		return nil, fmt.Errorf("create raw transaction error,err=%v", err1)
	}
	//获取私钥
	privateKey, err2 := ks.addressOrPublicKeyToPrivate(tp.FromAddress)
	if err2 != nil {
		return nil, err2
	}

	sig, err3 := originTx.SignTransaction(privateKey, message)
	if err3 != nil {
		return nil, fmt.Errorf("sign error,Err=[%v]", err3)
	}
	txHex, err4 := originTx.GetSignTransaction(sig)
	if err4 != nil {
		return nil, fmt.Errorf("merge signature and tx error,Err=[%v]", err4)
	}
	//提交交易
	txidBytes, err5 := ks.client.Rpc.SendRequest("author_submitExtrinsic", []interface{}{txHex})
	if err5 != nil {
		return nil, fmt.Errorf("submit transaction error,Err=[%v]", err5)
	}
	return string(txidBytes), nil
}

/*
创建地址实体方法
*/
func (ks *KsmService) createAddressInfo() (util.AddrInfo, error) {
	var addrInfo util.AddrInfo
	priv, pub, err := sr25519.GenerateKey()
	if err != nil {
		return addrInfo, err
	}
	secret, err1 := sr25519.PrivateKeyToHex(priv)
	if err1 != nil {
		return addrInfo, err1
	}
	address, err2 := sr25519.CreateAddress(pub, substrateConf.KsmPrefix)
	if err2 != nil {
		return addrInfo, err2
	}
	addrInfo.PrivKey = secret
	addrInfo.Address = address
	return addrInfo, nil
}
