package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	gofio "github.com/fioprotocol/fio-go"
	"github.com/fioprotocol/fio-go/eos/ecc"
	"github.com/group-coldwallet/wallet-sign/conf"
	"github.com/group-coldwallet/wallet-sign/model"
	"github.com/group-coldwallet/wallet-sign/util"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ripemd160"
	"strings"
	"time"
)

/*
service模板
*/

/*
币种服务结构体
*/
type FioService struct {
	*BaseService
}

/*
初始化币种服务
	注意：
		方法接受者： BaseService
		方法命名： 币种大写 + Service
*/
func (bs *BaseService) FIOService() *FioService {
	tp := new(FioService)
	tp.BaseService = bs
	//tp.cleint,_,_ =gofio.NewConnection(nil,conf.Config.FioCfg.NodeUrl)
	return tp
}

/*
接口创建地址服务
	无需改动
*/
func (fs *FioService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return fs.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, fs.createAddressInfo)
	}
	return fs.BaseService.createAddress(req, fs.createAddressInfo)
}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (fs *FioService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	_, err := fs.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, fs.createAddressInfo)
	return err
}

/*
签名服务
*/
func (fs *FioService) SignService(req *model.ReqSignParams) (interface{}, error) {
	var tp model.FioSignParams
	if err := fs.BaseService.parseData(req, &tp); err != nil {
		return nil, err
	}
	if &tp == nil {
		return nil, errors.New("transfer params is null")
	}
	if tp.FromAddress == "" || tp.Amount == "" || tp.ToAddress == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]",
			tp.FromAddress, tp.ToAddress, tp.Amount)
	}
	if tp.HeadBlockId == "" || tp.ChainId == "" {
		return nil, fmt.Errorf("head block id or chain id is null,hId=[%s],cId=[%s]",
			tp.HeadBlockId, tp.ChainId)
	}
	wif, err := fs.GetKeyByAddress(tp.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("get [%s] private key error,err=[%v]", tp.FromAddress, err)
	}
	kb, err := gofio.NewAccountFromWif(wif)
	if err != nil {
		return nil, fmt.Errorf("create account error,err=[%v]", err)
	}
	amount, _ := decimal.NewFromString(tp.Amount)
	//转换amount的为float64
	transAmount, _ := amount.Shift(-9).Float64()
	action := gofio.NewTransferTokensPubKey(kb.Actor, tp.ToAddress, gofio.Tokens(transAmount))
	if action == nil {
		return nil, fmt.Errorf("build action error")
	}
	//fs.signTx(action)
	return nil, nil
}

/*
热钱包出账服务
*/
func (fs *FioService) TransferService(req interface{}) (interface{}, error) {
	var tp model.FioTransferParams
	if err := fs.BaseService.parseData(req, &tp); err != nil {
		return nil, err
	}
	if &tp == nil {
		return nil, errors.New("transfer params is null")
	}
	if tp.FromAddress == "" || tp.Amount == "" || tp.ToAddress == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]",
			tp.FromAddress, tp.ToAddress, tp.Amount)
	}
	wif, err := fs.GetKeyByAddress(tp.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("get [%s] private key error,err=[%v]", tp.FromAddress, err)
	}
	kb, err := gofio.NewAccountFromWif(wif)
	if err != nil {
		return nil, fmt.Errorf("create account error,err=[%v]", err)
	}
	c, _, err := gofio.NewConnection(kb.KeyBag, conf.Config.FioCfg.NodeUrl)
	if err != nil {
		return nil, fmt.Errorf("connect node error,err=[%v]", err)
	}
	// 判断from金额是否足够
	fromBalance, err := fs.getFioBalance(tp.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("get from address [%s] balance error,err=[%v]", tp.FromAddress, err)
	}
	fb := decimal.NewFromInt(fromBalance)
	logrus.Printf("发送地址金额：%s \n", fb.String())
	//减掉手续费
	fee := decimal.NewFromInt(2).Shift(9)
	logrus.Printf("手续费金额：%s \n", fee.String())
	fb = fb.Sub(fee)
	logrus.Printf("实际需要使用金额：%s \n", fb.String())
	amount, _ := decimal.NewFromString(tp.Amount)
	logrus.Printf("订单需要使用金额：%s \n", amount.String())
	if fb.LessThan(amount) {
		return nil, fmt.Errorf("[%s] amount is not enough,online=[%v],transfer=[%v],fee=[2]", tp.FromAddress, fb.Shift(-9).Add(decimal.NewFromInt(2)), amount.Shift(-9))
	}
	//转换amount的为float64
	transAmount, _ := amount.Shift(-9).Float64()

	action := gofio.NewTransferTokensPubKey(kb.Actor, tp.ToAddress, gofio.Tokens(transAmount))
	if action == nil {
		return nil, fmt.Errorf("build action error")
	}
	out, err := c.SignPushActions(action)
	if err != nil {
		return nil, fmt.Errorf("sign and send transaction error,err=[%v]", err)
	}
	return out.TransactionID, nil
}

/*
创建地址实体方法
*/
func (fs *FioService) createAddressInfo() (util.AddrInfo, error) {
	acc, err := gofio.NewRandomAccount()
	if err != nil {
		return util.AddrInfo{}, err
	}
	wif := acc.KeyBag.Keys[0].String()
	address := acc.PubKey
	info := util.AddrInfo{
		"",
		address,
		wif,
	}
	return info, nil
}
func (fs *FioService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	balance, err := fs.getFioBalance(req.Address)
	if err != nil {
		return nil, fmt.Errorf("get %s amount error: %v", req.Address, err)
	}
	return decimal.NewFromInt(balance).String(), nil
}
func (fs *FioService) getFioBalance(publicKey string) (int64, error) {

	url := conf.Config.FioCfg.NodeUrl + "/v1/chain/get_fio_balance"
	params := map[string]string{
		"fio_public_key": publicKey,
	}
	data, err := util.PostJson(url, params)
	if err != nil {
		return 0, err
	}
	var resp map[string]interface{}
	err1 := json.Unmarshal(data, &resp)
	if err1 != nil {
		return 0, err1
	}
	return int64(resp["balance"].(float64)), nil
}
func (fs *FioService) ValidAddress(address string) error {
	return fs.validAddress(address)
}
func (fs *FioService) addTimeSecond(now, second int64) int64 {
	return time.Unix(now, 0).Add(time.Minute * time.Duration(second)).Unix()
}

//func (fs *FioService)signTx(headBlockId,chainId string ,action ...*gofio.Action)(string,error){
//	b := make([]*eos.Action, len(action))
//	for i, act := range action {
//		b[i] = act.ToEos()
//	}
//	opts:=&eos.TxOptions{}
//	hbId,_:=hex.DecodeString(headBlockId)
//	cId,_:=hex.DecodeString(chainId)
//	opts.HeadBlockID = hbId
//	opts.ChainID = cId
//	tx:=eos.NewTransaction(b,opts)
//	stx:=eos.NewSignedTransaction(tx)
//
//}

func (fs *FioService) validAddress(address string) error {
	if !strings.HasPrefix(address, "FIO") {
		return errors.New("dont have 'FIO' prefix")
	}
	data := base58.Decode(address[3:])
	if len(data) != 37 {
		return errors.New("bas58 decode data length is not equal 37")
	}
	checksum1 := data[33:]
	ck2 := fs.ripemd160checksum(data[:33], ecc.CurveK1)
	checksum2 := ck2[:4]
	if bytes.Compare(checksum1, checksum2) != 0 {
		return errors.New("checksum is not equal")
	}
	return nil
}

func (fs *FioService) ripemd160checksum(in []byte, curve ecc.CurveID) []byte {
	h := ripemd160.New()
	_, _ = h.Write(in) // this implementation has no error path

	if curve != ecc.CurveK1 {
		_, _ = h.Write([]byte(curve.String()))
	}

	sum := h.Sum(nil)
	return sum[:4]
}
