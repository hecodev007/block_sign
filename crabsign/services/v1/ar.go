package v1

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JFJun/arweave-go/api"
	arTx "github.com/JFJun/arweave-go/tx"
	"github.com/JFJun/arweave-go/utils"
	"github.com/JFJun/arweave-go/wallet"
	"github.com/mendsley/gojwk"
	"github.com/onethefour/common/log"
	"github.com/shopspring/decimal"
	"wallet-sign/conf"
	"wallet-sign/model"
	"wallet-sign/util"
)

const (
	KeyLength = 4096
)

type ArService struct {
	*BaseService
	client *api.Client
}

func (bs *BaseService) ARService() *ArService {
	ar := new(ArService)
	ar.BaseService = bs
	//初始化连接
	client, _ := api.Dial(conf.Config.ARCfg.NodeUrl)
	ar.client = client
	return ar
}
func (ar *ArService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return ar.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, ar.createAddressInfo)
	}
	return ar.BaseService.createAddress(req, ar.createAddressInfo)
}

func (ar *ArService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	_, err := ar.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, ar.createAddressInfo)
	return err
}
func (ar *ArService) SignService(req *model.ReqSignParams) (interface{}, error) {
	var (
		arSignparams model.ARSignParams
		err          error
	)
	err = ar.BaseService.parseData(req.Data, &arSignparams)
	if err != nil {
		return nil, fmt.Errorf("parse ar sign data error,err=[%v]", err)
	}
	if &arSignparams == nil {
		return nil, errors.New("ar sign data is null")
	}
	//判断参数是否为空
	if arSignparams.Amount == "" || arSignparams.FromAddress == "" || arSignparams.ToAddress == "" || arSignparams.Fee == "" {
		return nil, fmt.Errorf("sign params have null,from=[%s],to=[%s],amount=[%s],fee=[%s]", arSignparams.FromAddress,
			arSignparams.ToAddress, arSignparams.Amount, arSignparams.Fee)
	}
	var jwkPriv string
	//获取jwk格式私钥数据
	jwkPriv, err = ar.BaseService.addressOrPublicKeyToPrivate(arSignparams.FromAddress)
	if err != nil {
		return nil, err
	}
	w := wallet.NewWallet()
	err = w.LoadKey([]byte(jwkPriv))
	if err != nil {
		return nil, err
	}
	//构建交易
	txBuilder := arTx.NewTransactionV2(arSignparams.LastTx, w.PubKeyModulus(), arSignparams.Amount, arSignparams.ToAddress, nil, arSignparams.Fee)
	//签名交易
	var tx *arTx.TransactionV2
	tx, err = txBuilder.Sign(w)

	if err != nil {
		return nil, err
	}
	var data []byte
	data, err = json.Marshal(tx)
	if err != nil {
		return nil, err
	}
	//返回16进制交易数据，用于广播交易
	return hex.EncodeToString(data), nil
}

func (ar *ArService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	onlineAmount, err := ar.client.GetBalance(context.TODO(), req.Address)
	if err != nil {
		return nil, fmt.Errorf("get  address [%s] online amount error,Err=[%v]", req.Address, err)
	}
	return onlineAmount, nil
}

func (ar *ArService) ValidAddress(address string) error {
	data, err := base64.RawURLEncoding.DecodeString(address)
	if len(data) != 32 {
		return errors.New("invalid address ,address length is not correct")
	}
	if err != nil {
		return fmt.Errorf("base64 decode address error,Err=[%v]", err)
	}
	return nil
}

func (ar *ArService) TransferService(req interface{}) (interface{}, error) {
	var (
		tp  model.ARTransferParams
		err error
	)
	if err = ar.BaseService.parseData(req, &tp); err != nil {
		return nil, err
	}
	if &tp == nil {
		return nil, errors.New("transfer params is null")
	}
	if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	}
	log.Info(String(tp))
	if ar.client == nil {
		//重联节点
		client, err := api.Dial(conf.Config.ARCfg.NodeUrl)
		if err != nil {
			return nil, fmt.Errorf("reconnect node error,err=%v", err)
		}
		ar.client = client
	}
	//获取私钥
	//获取jwk格式私钥数据
	var jwkPriv string
	jwkPriv, err = ar.BaseService.addressOrPublicKeyToPrivate(tp.FromAddress)

	if err != nil {
		return nil, err
	}
	w := wallet.NewWallet()
	err = w.LoadKey([]byte(jwkPriv))
	if err != nil {
		log.Info(err.Error())
		return nil, err
	}

	var (
		lastTx string
		price  string
	)
	lastTx, err = ar.client.GetTransactionAnchor(context.TODO())
	if err != nil {
		log.Info(err.Error())
		return nil, err
	}
	price, err = ar.client.GetRewardV2(context.TODO(), nil, tp.ToAddress)
	if err != nil {
		log.Info(err.Error())
		return nil, err
	}
	//fee,_:=decimal.NewFromString(price)
	toAmount, _ := decimal.NewFromString(tp.Amount)
	if toAmount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("amount less or equal than 0,Amount=[%s]", tp.Amount)
	}
	onlineAmount, errA := ar.client.GetBalance(context.TODO(), tp.FromAddress)
	if errA != nil {
		return nil, fmt.Errorf("get from address [%s] online amount error,Err=[%v]", tp.FromAddress, errA)
	}
	if onlineAmount == "0" || onlineAmount == "" {
		return nil, fmt.Errorf("from address [%s] online amount is null", tp.FromAddress)
	}
	balance, _ := decimal.NewFromString(onlineAmount)
	if balance.LessThanOrEqual(toAmount) {
		return nil, fmt.Errorf("online amount is less or equal than trans amount,OnlineAmount=[%s],TransAmount=[%s]", onlineAmount, tp.Amount)
	}
	txV2 := arTx.NewTransactionV2(lastTx, w.PubKeyModulus(), tp.Amount, tp.ToAddress, nil, price)
	tx, errSig := txV2.Sign(w)

	if errSig != nil {
		return nil, errSig
	}
	d, _ := json.Marshal(tx)
	log.Info(String(tp))
	log.Info("rawtx:" + string(d))
	//return "test",nil
	var resp string
	resp, err = ar.client.Commit(context.TODO(), d)
	if err != nil {
		return resp, err
	}
	log.Info("commit resp:" + resp)
	if resp == "OK" {
		return utils.EncodeToBase64(tx.ID()), nil
	}
	return resp, errors.New("broadcast transaction error,resp code is not OK")
}

/*
ref: github.com/Dec43/arweave-go/wallet/wallet.go  method； loadKey  page: 91
*/
func (ar *ArService) createAddressInfo() (util.AddrInfo, error) {
	var addrInfo util.AddrInfo
	priv, err := rsa.GenerateKey(rand.Reader, KeyLength)
	if err != nil {
		return addrInfo, err
	}
	key, err := gojwk.PrivateKey(priv)
	privBytes, err := gojwk.Marshal(key)
	if err != nil {
		return addrInfo, err
	}
	h := sha256.New()
	h.Write(priv.PublicKey.N.Bytes())
	address := utils.EncodeToBase64(h.Sum(nil))
	addrInfo.PrivKey = string(privBytes)
	addrInfo.Address = address
	return addrInfo, nil
}
func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
