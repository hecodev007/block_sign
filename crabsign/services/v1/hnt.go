package v1

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/JFJun/helium-go/http"
	"github.com/JFJun/helium-go/keypair"
	"github.com/JFJun/helium-go/transactions"
	"github.com/btcsuite/btcutil/base58"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
	"wallet-sign/conf"
	"wallet-sign/model"
	"wallet-sign/util"
)

type HntService struct {
	*BaseService
	rpc           *http.HeliumRpc
	kp            *keypair.Keypair
	transLimitMap map[string]*HntTransfer
	limitMap      sync.Map
}
type HntTransfer struct {
	Amount    string
	Timestamp int64
}

func (bs *BaseService) HNTService() *HntService {
	hs := new(HntService)
	hs.BaseService = bs
	hs.kp = keypair.New(keypair.Ed25519Version)

	hs.rpc = http.NewHeliumRpc(conf.Config.HntCfg.NodeUrl)
	hs.transLimitMap = make(map[string]*HntTransfer)
	hs.limitMap = sync.Map{}
	return hs
}

//todo use new method
func (hs *HntService) getLimitMapData2(address, amount string) (int, error) {
	//获取链上余额
	fromAccount, err := hs.rpc.GetAccountByAddress(address)
	if err != nil {
		return -1, err
	}
	onlineAmount := decimal.NewFromInt(int64(fromAccount.Balance))
	log.Infof("[%s] online amount=[%d]", address, onlineAmount.IntPart())
	balance, _ := decimal.NewFromString(amount)
	log.Infof("[%s] local amount =[%d]", address, balance.IntPart())
	calcAmount := func(onlineAmount, transAmount decimal.Decimal) (string, error) {
		if transAmount.GreaterThan(onlineAmount) {
			return "", fmt.Errorf("address=%s amount is not enough,transAmount=%s,onlineAmount=%s", address, transAmount.String(), onlineAmount.String())
		}
		return onlineAmount.Sub(transAmount).String(), nil
	}
	v, ok := hs.limitMap.Load(address)

	if ok {
		trans := v.(HntTransfer)
		nowAmount, _ := decimal.NewFromString(trans.Amount)
		if nowAmount.Equal(onlineAmount) { //如果限制表里面的金额与链上相同，表示链上的余额不完成
			subAmount, err := calcAmount(onlineAmount, balance)
			if err != nil {
				return -1, fmt.Errorf("链上余额同步完成，但是余额不满足出账条件：%v", err)
			}
			trans := HntTransfer{
				Amount:    subAmount,
				Timestamp: time.Now().Unix(),
			}
			hs.limitMap.Store(address, trans)
			return fromAccount.SpeculativeNonce, nil
		} else {
			frzenTime := hs.addTimeSecond(trans.Timestamp, conf.Config.HntCfg.LockTime)
			now := time.Now().Unix()
			if frzenTime > now {
				//error
				return -1, fmt.Errorf("地址[%s]冻结,不能转账,解冻时间为： %s", address, util.Timestamp(frzenTime))
			}
			//链上余额可能充值，大雨冻结余额
			subAmount, err := calcAmount(onlineAmount, balance)
			if err != nil {
				hs.limitMap.Delete(address)
				return -1, fmt.Errorf("已过解冻时间,链上余额仍未同步完成，但是本地保存余额不满足出账条件：%v", err)
			}
			//过了冻结时间，可以出账
			trans := HntTransfer{
				Amount:    subAmount,
				Timestamp: time.Now().Unix(),
			}
			hs.limitMap.Store(address, trans)
			return fromAccount.SpeculativeNonce, nil
		}
	}
	subAmount, err := calcAmount(onlineAmount, balance)
	if err != nil {
		return -1, err
	}
	trans := HntTransfer{
		Amount:    subAmount,
		Timestamp: time.Now().Unix(),
	}
	hs.limitMap.Store(address, trans)
	//清理所有的解冻交易
	hs.limitMap.Range(func(key, value interface{}) bool {
		trans := value.(HntTransfer)
		now := time.Now().Unix()
		freznTime := trans.Timestamp
		if now > freznTime {
			hs.limitMap.Delete(key)
		}
		return true
	})
	return fromAccount.SpeculativeNonce, nil
}

func (hs *HntService) ValidAddress(address string) error {
	data := base58.Decode(address)
	if data == nil || len(data) == 0 {
		return errors.New("base58 decode address error")
	}

	if data[0] != byte(0) {
		return errors.New("invalid version")
	}
	if data[1] > byte(1) {
		return fmt.Errorf("invalid curve version,[0 is NIST p256,1 is ed25519],curve=%d", data[1])
	}
	checksum1 := data[len(data)-4:]
	ck := util.DoubleSha256(data[:len(data)-4])
	checksum2 := ck[:4]
	if !bytes.Equal(checksum1, checksum2) {
		return errors.New("invalid checksum")
	}
	return nil
}
func (hs *HntService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	fromAccount, err := hs.rpc.GetAccountByAddress(req.Address)
	if err != nil {
		return nil, fmt.Errorf("get hnt balance error: %v", err)
	}
	onlineAmount := decimal.NewFromInt(int64(fromAccount.Balance))
	return onlineAmount.String(), nil
}
func (hs *HntService) getLimitMapData(address, amount string) (int, error) {
	//获取链上余额
	fromAccount, err := hs.rpc.GetAccountByAddress(address)
	if err != nil {
		return -1, err
	}
	onlineAmount := decimal.NewFromInt(int64(fromAccount.Balance))
	log.Printf("[%s] online amount=[%d]", address, onlineAmount.IntPart())
	balance, _ := decimal.NewFromString(amount)
	log.Printf("[%s] send amount =[%d]", address, balance.IntPart())
	calcAmount := func(onlineAmount, transAmount decimal.Decimal) (string, error) {
		if transAmount.GreaterThan(onlineAmount) {
			return "", fmt.Errorf("address=%s amount is not enough,transAmount=%s,onlineAmount=%s", address, transAmount.String(), onlineAmount.String())
		}
		return onlineAmount.Sub(transAmount).String(), nil
	}

	if len(hs.transLimitMap) >= 0 {
		for k, v := range hs.transLimitMap {
			if k == address { //说明这个地址的交易存在限制表中
				//查看限制表的金额
				nowAmount, _ := decimal.NewFromString(v.Amount)
				log.Printf("[%s] 冻结池中的金额 [%s]", k, nowAmount.String())
				if nowAmount.Equal(onlineAmount) { //如果限制表里面的金额与链上相同，表示链上的余额同步完成
					subAmount, err := calcAmount(onlineAmount, balance)
					if err != nil {
						return -1, fmt.Errorf("链上余额同步完成，但是余额不满足出账条件：%v", err)
					}
					trans := HntTransfer{
						Amount:    subAmount,
						Timestamp: time.Now().Unix(),
					}
					hs.transLimitMap[address] = &trans
					return fromAccount.SpeculativeNonce, nil
				} else {
					//链上余额还没有返回
					// 1. 判断冷冻时间是否大于三分钟
					//saveAmount, _ := decimal.NewFromString(v.Amount)

					frzenTime := hs.addTimeSecond(v.Timestamp, conf.Config.HntCfg.LockTime)
					now := time.Now().Unix()
					if frzenTime > now {
						return -1, fmt.Errorf("地址[%s]冻结,不能转账,冻结时间为： %d分钟，解冻时间为： %s", address, conf.Config.HntCfg.LockTime, util.Timestamp(frzenTime))
					}
					subAmount, err := calcAmount(onlineAmount, balance)
					if err != nil {
						delete(hs.transLimitMap, k)
						return -1, fmt.Errorf("已过解冻时间,链上余额仍未同步完成，但是本地保存余额不满足出账条件：%v", err)
					}
					//过了冻结时间，可以出账
					trans := HntTransfer{
						Amount:    subAmount,
						Timestamp: time.Now().Unix(),
					}
					hs.transLimitMap[address] = &trans
					return fromAccount.SpeculativeNonce, nil
				}
			}
		}
	}
	log.Printf("第一次进入冻结池： %s", address)
	subAmount, err := calcAmount(onlineAmount, balance)
	if err != nil {
		return -1, err
	}
	trans := HntTransfer{
		Amount:    subAmount,
		Timestamp: time.Now().Unix(),
	}
	hs.transLimitMap[address] = &trans

	//清理map内存,过了解冻时间的地址可以从内存中移除掉
	for k, v := range hs.transLimitMap {
		now := time.Now().Unix()
		freznTime := v.Timestamp
		if now > freznTime {
			log.Printf("解除地址冻结： %s", address)
			delete(hs.transLimitMap, k)
		}
	}

	return fromAccount.SpeculativeNonce, nil
}

func (hs *HntService) addTimeSecond(now, second int64) int64 {
	return time.Unix(now, 0).Add(time.Minute * time.Duration(second)).Unix()
}

func (hs *HntService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return hs.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, hs.createAddressInfo)
	}
	return hs.BaseService.createAddress(req, hs.createAddressInfo)
}

func (hs *HntService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	_, err := hs.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, hs.createAddressInfo)
	return err
}
func (hs *HntService) SignService(req *model.ReqSignParams) (interface{}, error) {

	return nil, nil
}

func (hs *HntService) TransferService(req interface{}) (interface{}, error) {
	var tp model.HntTransferParams
	if err := hs.BaseService.parseData(req, &tp); err != nil {

		return nil, err
	}
	if &tp == nil {
		return nil, errors.New("transfer params is null")
	}
	if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	}

	speculativeNonce, err := hs.getLimitMapData(tp.FromAddress, tp.Amount)
	if err != nil {
		return nil, err
	}
	nonce := speculativeNonce + 1
	log.Printf("%s nonce is %d", tp.FromAddress, nonce)
	vars, err := hs.rpc.GetVars()
	if err != nil || vars == nil {
		log.Errorf("get vars error,err=%v", err)
		return nil, fmt.Errorf("get vars error,err=%v", err)
	}
	var dc_payload_size int64
	if vars.DcPayloadSize <= 0 {
		dc_payload_size = int64(transactions.DC_Payload_Size)
	} else {
		dc_payload_size = vars.DcPayloadSize
	}
	//	构建交易
	from := keypair.NewAddressable(tp.FromAddress)

	amountDec, _ := decimal.NewFromString(tp.Amount)
	amount := amountDec.IntPart()
	tmpSig := make([]byte, 64)
	to := keypair.NewAddressable(tp.ToAddress)
	v1 := transactions.NewPaymentV1Tx(from, to, uint64(amount), 0, uint64(nonce), tmpSig)
	payload, err := v1.Serialize()
	if err != nil {
		log.Errorf("serialize v1 error,err=%v", err)
		//return nil, fmt.Errorf()
	}
	////write by jun --->add tx fee    2020/07/01
	//fee,errFee:=hs.calcFee(tp.FeeInt)
	//if errFee != nil {
	//	return nil,fmt.Errorf("calc fee error,err=%v",errFee)
	//}
	//onlineFee  := uint64(0)
	//if vars !=nil{
	//	onlineFee = transactions.CalculateFee(int64(len(payload)), dc_payload_size, vars.TxnFeeMultiplier)
	//}
	fee := transactions.CalculateFee(int64(len(payload)), dc_payload_size, vars.TxnFeeMultiplier)
	// 提高手续费
	//fee  = fee +10000
	v1.SetFee(fee)
	v1Tx, err1 := v1.BuildTransaction(true)
	if err1 != nil {
		return nil, fmt.Errorf("build v1 tx error,err=%v", err1)
	}
	//获取私钥
	wif, err2 := hs.BaseService.addressOrPublicKeyToPrivate(tp.FromAddress)
	if err2 != nil {
		return nil, fmt.Errorf("get private key error,address=%s,err=%v", tp.FromAddress, err2)
	}
	kp := keypair.NewKeypairFromWIF(1, wif)

	//签名交易
	sig, err3 := kp.Sign(v1Tx)
	if err3 != nil {
		return nil, fmt.Errorf("sign v1 transaction error,err=%v", err3)
	}
	v1.SetSignature(sig)

	ser, err4 := v1.Serialize()
	if err4 != nil {
		return nil, fmt.Errorf("serialize v1 transaction error,err=%v", err4)
	}
	txn := base64.StdEncoding.EncodeToString(ser)
	//获取当前的price
	cp, errCp := hs.rpc.GetCurrentPrices()
	if errCp != nil || cp == nil {
		return nil, fmt.Errorf("get current price error,err=[%v]", errCp)
	}
	feeDec := decimal.NewFromInt(int64(fee)).Shift(-5)

	price := decimal.NewFromInt(cp.Price).Shift(-8)

	hntFee := feeDec.DivRound(price, int32(8)).String()
	log.Printf("%s use fee is %s", tp.FromAddress, hntFee)
	//提交交易
	txid, err5 := hs.rpc.BroadcastTransaction(txn[:])
	if err5 != nil {
		return nil, fmt.Errorf("broadcast v1 transaction error,err=%v", err5)
	}
	log.Infof("success send a tx,txid=[%s],fee=[%s]", txid, hntFee)
	resp := map[string]interface{}{
		"txid":      txid,
		"fee_float": hntFee,
	}
	return resp, nil
}
func (hs *HntService) createAddressInfo() (util.AddrInfo, error) {
	var addrInfo util.AddrInfo
	wif, address := hs.kp.GenerateWifAndAddress()
	if wif == "" || address == "" {
		return addrInfo, fmt.Errorf("wif or address is null,wif=[%s],address=[%s]", wif, address)
	}
	addrInfo.PrivKey = wif
	addrInfo.Address = address
	return addrInfo, nil
}
func (hs *HntService) calcFee(feeStr string) (uint64, error) {
	feeDec, err := decimal.NewFromString(feeStr)
	if err != nil {
		return 0, err
	}

	cp, err := hs.rpc.GetCurrentPrices()
	if err != nil {
		return 0, fmt.Errorf("get current price error,Err=[%v]", err)
	}
	price := decimal.NewFromInt(cp.Price)
	fee := feeDec.Mul(price).Shift(-3).IntPart()
	return uint64(fee), nil
}
