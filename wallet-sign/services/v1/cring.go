package v1

import "C"
import (
	"errors"
	"fmt"
	"github.com/JFJun/go-substrate-crypto/crypto"
	"github.com/JFJun/go-substrate-crypto/ss58"
	substrateConf "github.com/coldwallet-group/substrate-go/config"
	"github.com/coldwallet-group/substrate-go/sr25519"
	"github.com/coldwallet-group/bifrost-go/tx"
	"github.com/coldwallet-group/bifrost-go/client"
	"github.com/coldwallet-group/bifrost-go/expand"

	"github.com/group-coldwallet/wallet-sign/conf"
	"github.com/group-coldwallet/wallet-sign/model"
	"github.com/group-coldwallet/wallet-sign/util"
	"github.com/shopspring/decimal"
)

/*
service模板
*/

/*
币种服务结构体
*/

type CringService struct {
	*BaseService
	client   *client.Client
	url      string
	nonceCtl map[string]int64
}

/*
初始化币种服务
	注意：
		方法接受者： BaseService
		方法命名： 币种大写 + Service
*/
func (bs *BaseService) CRINGService() *CringService {
	ks := new(CringService)
	ks.BaseService = bs
	ks.url = conf.Config.CringCfg.NodeUrl
	ks.nonceCtl = make(map[string]int64)
	return ks
}

/*
接口创建地址服务
	无需改动
*/
func (ks *CringService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return ks.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, ks.createAddressInfo)
	}
	return ks.BaseService.createAddress(req, ks.createAddressInfo)
}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (ks *CringService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	fmt.Println("start create Cring address")
	_, err := ks.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, ks.createAddressInfo)
	return err
}

func (ks *CringService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	aInfo, err := ks.client.GetAccountInfo(req.Address)
	if err != nil {
		return nil, fmt.Errorf("get cring address amount error,err=%v", err)
	}

	free := aInfo.Data.Free
	balance := decimal.NewFromInt(free.Int64())
	return balance.String(), nil
}

func (ks *CringService) ValidAddress(address string) error {
	return ss58.VerityAddress(address, ss58.SubstratePrefix)
}

/*
签名服务
*/
func (ks *CringService) SignService(req *model.ReqSignParams) (interface{}, error) {
	var tp model.CringColdParams
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
	transaction := tx.NewSubstrateTransaction(tp.FromAddress, tp.Nonce)
	//5. 初始化metadata的扩张结构
	ed, err := expand.NewMetadataExpand(ks.client.Meta)
	if err != nil {
		return nil,err
	}
	//6. 初始化Balances.transfer的call方法
	call, err := ed.BalanceTransferCall(tp.ToAddress, amount.BigInt())
	if err != nil {
		return nil,err
	}
	/*
		//Balances.transfer_keep_alive  call方法
		btkac,err:=ed.BalanceTransferKeepAliveCall(to,amount)
	*/

	/*
		toAmount:=make(map[string]uint64)
		toAmount[to] = amount
		//...
		//true: user Balances.transfer_keep_alive  false: Balances.transfer
		ubtc,err:=ed.UtilityBatchTxCall(toAmount,false)
	*/

	//7. 设置交易的必要参数
	transaction.SetGenesisHashAndBlockHash(ks.client.GetGenesisHash(), ks.client.GetGenesisHash()).
		SetSpecAndTxVersion(uint32(ks.client.SpecVersion), uint32(ks.client.TransactionVersion)).
		SetCall(call) //设置call
	privateKey, err := ks.addressOrPublicKeyToPrivate(tp.FromAddress)
	if err != nil {
		return nil, err
	}
	//8. 签名交易
	sig, err := transaction.SignTransaction(privateKey, crypto.Sr25519Type)
	if err != nil {
		return nil,err
	}
	return sig, nil
}

/*
热钱包出账服务
*/
func (ks *CringService) TransferService(req interface{}) (interface{}, error) {
	if ks.client == nil {
		//初始化客户端
		c, err := client.New(ks.url)
		if err != nil {
			return nil, err
		}
		ks.client = c
	}
	var tp model.CringTransferParams
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
	aInfo, err := ks.client.GetAccountInfo(tp.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("get from address amount error,err=%v", err)
	}

	free := aInfo.Data.Free
	amount, _ := decimal.NewFromString(tp.Amount)
	balance := decimal.NewFromInt(free.Int64())
	if balance.LessThanOrEqual(amount) {
		return nil, fmt.Errorf("%s is not enough amount to transfer,transferAmount=%s,actuallyAmount=%d", tp.FromAddress,
			tp.Amount, free)
	}

	transaction := tx.NewSubstrateTransaction(tp.FromAddress, uint64(aInfo.Nonce))
	//5. 初始化metadata的扩张结构
	ed, err := expand.NewMetadataExpand(ks.client.Meta)
	if err != nil {
		return nil,err
	}
	//6. 初始化Balances.transfer的call方法
	call, err := ed.BalanceTransferCall(tp.ToAddress, amount.BigInt())
	if err != nil {
		return nil,err
	}
	/*
		//Balances.transfer_keep_alive  call方法
		btkac,err:=ed.BalanceTransferKeepAliveCall(to,amount)
	*/

	/*
		toAmount:=make(map[string]uint64)
		toAmount[to] = amount
		//...
		//true: user Balances.transfer_keep_alive  false: Balances.transfer
		ubtc,err:=ed.UtilityBatchTxCall(toAmount,false)
	*/

	//7. 设置交易的必要参数
	transaction.SetGenesisHashAndBlockHash(ks.client.GetGenesisHash(), ks.client.GetGenesisHash()).
		SetSpecAndTxVersion(uint32(ks.client.SpecVersion), uint32(ks.client.TransactionVersion)).
		SetCall(call) //设置call
	privateKey, err := ks.addressOrPublicKeyToPrivate(tp.FromAddress)
	if err != nil {
		return nil, err
	}
	//8. 签名交易
	sig, err := transaction.SignTransaction(privateKey, crypto.Sr25519Type)
	if err != nil {
		return nil,err
	}
	//9. 提交交易
	var result interface{}
	err = ks.client.C.Client.Call(&result, "author_submitExtrinsic", sig)
	if err != nil {
		return nil,err
	}
	//10. txid
	txid := result.(string)
	return txid, nil
}

/*
创建地址实体方法
*/
func (ks *CringService) createAddressInfo() (util.AddrInfo, error) {
	var addrInfo util.AddrInfo
	priv, pub, err := sr25519.GenerateKey()
	if err != nil {
		return addrInfo, err
	}
	secret, err1 := sr25519.PrivateKeyToHex(priv)
	if err1 != nil {
		return addrInfo, err1
	}
	address, err2 := sr25519.CreateAddress(pub, substrateConf.SubstratePrefix)
	if err2 != nil {
		return addrInfo, err2
	}
	addrInfo.PrivKey = secret
	addrInfo.Address = address
	return addrInfo, nil
}
