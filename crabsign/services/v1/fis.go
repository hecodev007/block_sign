package v1

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/JFJun/go-substrate-crypto/crypto"
	"github.com/JFJun/go-substrate-crypto/ss58"
	c "github.com/coldwallet-group/stafi-substrate-go/client"
	"github.com/coldwallet-group/stafi-substrate-go/expand"
	tx "github.com/coldwallet-group/stafi-substrate-go/tx"
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
type FisService struct {
	*BaseService
	client   *c.Client
	url      string
	nonceCtl map[string]int64
	ed       *expand.MetadataExpand
}

/*
初始化币种服务
	注意：
		方法接受者： BaseService
		方法命名： 币种大写 + Service
*/
func (bs *BaseService) FISService() *FisService {
	ks := new(FisService)
	ks.BaseService = bs
	ks.url = conf.Config.FisCfg.NodeUrl
	ks.nonceCtl = make(map[string]int64)

	return ks
}

/*
接口创建地址服务
	无需改动
*/
func (ks *FisService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return ks.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, ks.createAddressInfo)
	}
	return ks.BaseService.createAddress(req, ks.createAddressInfo)
}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (ks *FisService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	fmt.Println("start create Fis address")
	_, err := ks.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, ks.createAddressInfo)
	return err
}

/*
签名服务
*/
func (ks *FisService) SignService(req *model.ReqSignParams) (interface{}, error) {
	return nil, errors.New("unsupport cold sign")
}

func (ks *FisService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	if ks.client == nil {
		//初始化客户端
		client, err := c.New(ks.url)
		if err != nil {
			return nil, err
		}
		client.SetPrefix(ss58.StafiPrefix)
		ks.client = client

		if ks.ed == nil {
			ks.ed, _ = expand.NewMetadataExpand(client.Meta)
		}
	}
	ai, err := ks.client.GetAccountInfo(req.Address)
	if err != nil {
		return nil, fmt.Errorf("get from address amount error,err=%v", err)
	}
	if ai == nil {
		return nil, fmt.Errorf("null")
	}
	return ai.Data.Free.String(), nil
}

func (ks *FisService) ValidAddress(address string) error {
	return ss58.VerityAddress(address, ss58.StafiPrefix)
}

/*
热钱包出账服务
*/
func (ks *FisService) TransferService(req interface{}) (interface{}, error) {
	if ks.client == nil {
		//初始化客户端
		client, err := c.New(ks.url)
		if err != nil {
			return nil, err
		}
		client.SetPrefix(ss58.StafiPrefix)
		ks.client = client

		if ks.ed == nil {
			ks.ed, _ = expand.NewMetadataExpand(client.Meta)
		}
	}
	var tp model.FisTransferParams
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
	ai, err := ks.client.GetAccountInfo(tp.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("get from address amount error,err=%v", err)
	}
	//var aInfo model.AccountInfo
	//if err := json.Unmarshal(data, &aInfo); err != nil {
	//	return nil, err
	//}
	free := ai.Data.Free.String()
	amount, _ := decimal.NewFromString(tp.Amount)
	balance, _ := decimal.NewFromString(free)
	if balance.LessThanOrEqual(amount) {
		return nil, fmt.Errorf("%s is not enough amount to transfer,transferAmount=%s,actuallyAmount=%d", tp.FromAddress,
			tp.Amount, free)
	}
	//types.SetSerDeOptions(types.SerDeOptions{NoPalletIndices: true})
	originTx := tx.CreateTransaction(tp.FromAddress, tp.ToAddress, uint64(amount.IntPart()), uint64(uint32(ai.Nonce)))
	originTx.SetGenesisHashAndBlockHash(ks.client.GetGenesisHash(), ks.client.GetGenesisHash())
	var callIdx string
	callIdx, err = ks.ed.MV.GetCallIndex("Balances", "transfer")
	if err != nil {
		return nil, fmt.Errorf("get balance.transfer call index error: %v", err)
	}
	originTx.SetSpecVersionAndCallId(uint32(ks.client.SpecVersion), uint32(ks.client.TransactionVersion), callIdx)
	//获取私钥
	privateKey, err2 := ks.addressOrPublicKeyToPrivate(tp.FromAddress)
	if err2 != nil {
		return nil, err2
	}
	sig, err3 := originTx.SignTransaction(privateKey, crypto.Sr25519Type)
	if err3 != nil {
		return nil, fmt.Errorf("sign error,Err=[%v]", err3)
	}
	var result interface{}
	err = ks.client.C.Client.Call(&result, "author_submitExtrinsic", sig)
	if err != nil || result == nil {
		return nil, fmt.Errorf("sign error: %v", err)
	}
	txid := result.(string)
	return txid, nil
}

/*
创建地址实体方法
*/
func (ks *FisService) createAddressInfo() (util.AddrInfo, error) {
	var addrInfo util.AddrInfo

	priv, pub, err := crypto.GenerateSubstrateKey(crypto.Sr25519Type)
	if err != nil {
		return addrInfo, err
	}
	if len(priv) != 32 {
		return addrInfo, errors.New("private key length is not equal 32")
	}
	wif := hex.EncodeToString(priv)
	var address string
	address, err = crypto.CreateSubstrateAddress(pub, ss58.StafiPrefix)
	if err != nil {
		return addrInfo, err
	}
	addrInfo.PrivKey = "0x" + wif
	addrInfo.Address = address
	return addrInfo, nil
}
