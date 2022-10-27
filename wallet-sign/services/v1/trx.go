package v1

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/JFJun/trx-sign-go/genkeys"
	"github.com/JFJun/trx-sign-go/grpcs"
	"github.com/JFJun/trx-sign-go/sign"
	"github.com/btcsuite/btcutil/base58"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/group-coldwallet/wallet-sign/conf"
	"github.com/group-coldwallet/wallet-sign/model"
	"github.com/group-coldwallet/wallet-sign/redis"
	"github.com/group-coldwallet/wallet-sign/util"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

/*
service模板
*/

/*
币种服务结构体
*/
type TrxService struct {
	*BaseService
	backClients map[string]*grpcs.Client
	mainClient  *grpcs.Client
	mainUrl     string
}

/*
初始化币种服务
	注意：
		方法接受者： BaseService
		方法命名： 币种大写 + Service
*/
func (bs *BaseService) TRXService() *TrxService {
	var err error
	cs := new(TrxService)
	cs.BaseService = bs
	cs.mainUrl = strings.TrimPrefix(conf.Config.TrxCfg.NodeUrl, "https://")
	cs.mainUrl = strings.TrimPrefix(conf.Config.TrxCfg.NodeUrl, "http://")
	cs.mainClient, err = grpcs.NewClient(cs.mainUrl)

	if err != nil {
		panic(fmt.Errorf("init rpc client error: %v", err))
	}
	err = cs.mainClient.SetTimeout(time.Second * 30)
	if err != nil {
		panic(fmt.Errorf("set timeout  error: %v", err))
	}
	cs.backClients = make(map[string]*grpcs.Client)
	cs.backClients[cs.mainUrl] = cs.mainClient
	cs.initClients()
	log.Infof("配置back节点：%d,可用节点数(包含主节点)：%d", len(conf.Config.TrxCfg.BackUrls), len(cs.backClients))

	return cs
}

func (cs *TrxService) initClients() {
	if len(conf.Config.TrxCfg.BackUrls) == 0 {
		return
	}
	for _, url := range conf.Config.TrxCfg.BackUrls {
		url = strings.TrimPrefix(url, "https://")
		url = strings.TrimPrefix(url, "http://")
		c, err := grpcs.NewClient(url)
		if err != nil {
			log.Errorf("init back url %s client error: %v", url, err)
			continue
		}
		err = c.SetTimeout(time.Second * 30)
		if err != nil {
			log.Errorf("back url %s client set timeout error: %v", url, err)
			continue
		}
		cs.backClients[url] = c
	}
}

func (cs *TrxService) getClient() *grpcs.Client {
	if len(cs.backClients) == 1 {
		log.Infof("map中仅有主节点，使用主节点IP=[%s]", cs.mainUrl)
		return cs.mainClient
	}
	for url, client := range cs.backClients {
		if client == nil {
			newClient, err := grpcs.NewClient(url)
			if err != nil {
				log.Errorf("[%s]备用节点不可用,将从map中移除该即节点", url)
				delete(cs.backClients, url)
				log.Infof("最新map节点个数为：%d", len(cs.backClients))
			}
			cs.backClients[url] = newClient
			continue
		}
		// 判断节点是否可用
		_, err := client.GRPC.GetNodeInfo()
		if err != nil {
			log.Errorf("节点[%s]不可用，将使用主节点并移除该节点 ，error: %v", url, err)
			delete(cs.backClients, url)
			log.Infof("最新map节点个数为：%d", len(cs.backClients))
			return cs.mainClient
		}
		log.Infof("使用节点IP=[%s]", url)
		return client
	}
	log.Infof("使用主节点IP=[%s]", cs.mainUrl)
	return cs.mainClient
}

/*
接口创建地址服务
	无需改动
*/
func (cs *TrxService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return cs.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, cs.createAddressInfo)
	}
	return cs.BaseService.createAddress(req, cs.createAddressInfo)
}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (cs *TrxService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	fmt.Println("start create cph address")
	_, err := cs.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, cs.createAddressInfo)
	return err
}

/*
签名服务
*/
func (cs *TrxService) SignService(req *model.ReqSignParams) (interface{}, error) {
	return nil, nil
}
func (cs *TrxService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	var resp = make(map[string]string)
	if req.ContractAddress != "" {
		resp["coin"] = req.Token
		if cs.isTrc10(req.ContractAddress) {
			balance, err := cs.getClient().GetTrc10Balance(req.Address, req.ContractAddress)
			if err != nil {
				return nil, fmt.Errorf("get trc10 balance error: %v", err)
			}
			resp["amount"] = decimal.NewFromInt(balance).String()
		} else {
			balance, err := cs.getClient().GetTrc20Balance(req.Address, req.ContractAddress)
			if err != nil {
				return nil, fmt.Errorf("get trc20 balance error: %v", err)
			}
			a, _ := decimal.NewFromString(balance.String())
			resp["amount"] = a.String()
		}
	} else {
		//主链金额
		acc, err := cs.getClient().GetTrxBalance(req.Address)
		if err != nil {
			return nil, fmt.Errorf("get trx balance error: %v", err)
		}
		resp["coin"] = req.CoinName
		resp["amount"] = decimal.NewFromInt(acc.Balance).String()
	}
	return resp, nil
}
func (cs *TrxService) ValidAddress(address string) error {
	if !strings.HasPrefix(address, "T") {
		return fmt.Errorf("address is not has prefix 'T' :%s ", address)
	}
	decodeCheck := base58.Decode(address)
	if len(decodeCheck) == 0 {
		return fmt.Errorf("b58 decode %s error", address)
	}

	if len(decodeCheck) < 4 {
		return fmt.Errorf("b58 data length is less 4 : %s ", address)
	}

	decodeData := decodeCheck[:len(decodeCheck)-4]

	h256h0 := sha256.New()
	h256h0.Write(decodeData)
	h0 := h256h0.Sum(nil)

	h256h1 := sha256.New()
	h256h1.Write(h0)
	h1 := h256h1.Sum(nil)

	if h1[0] == decodeCheck[len(decodeData)] &&
		h1[1] == decodeCheck[len(decodeData)+1] &&
		h1[2] == decodeCheck[len(decodeData)+2] &&
		h1[3] == decodeCheck[len(decodeData)+3] {
		return nil
	}
	return fmt.Errorf("b58 check sum error: %s", address)
}
func (cs *TrxService) isTrc10(contractAddress string) bool {
	for _, a := range contractAddress {
		if a > 57 || a < 48 {
			return false
		}
	}
	return true
}

/*
热钱包出账服务
*/
func (cs *TrxService) TransferService(req interface{}) (interface{}, error) {
	var tp model.TrxTransferParams
	if err := cs.BaseService.parseData(req, &tp); err != nil {
		return nil, err
	}
	if &tp == nil {
		return nil, errors.New("transfer params is null")
	}
	log.Infof("待签名订单入参 %v", tp)
	if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	}
	var err error
	client := cs.getClient()

	var amount decimal.Decimal
	amount, err = decimal.NewFromString(tp.Amount)
	if err != nil {
		return nil, fmt.Errorf("parse decimal amount error: %v", err)
	}
	var aTx *api.TransactionExtention
	// trc20合约转账
	if tp.ContractAddress != "" && tp.AssetId == "" {
		//验证地址余额，看是否足够转账
		chainAmount, err := client.GetTrc20Balance(tp.FromAddress, tp.ContractAddress)
		if err != nil {
			return nil, fmt.Errorf("get trc20 %s chain balance error: %v", tp.ContractAddress, err)
		}
		caDec, _ := decimal.NewFromString(chainAmount.String())
		if amount.GreaterThan(caDec) {
			return nil, fmt.Errorf("[%s] amount is not engouth,contract_address=[%s],"+
				"transAmount=[%s],chainAmount=[%s]",
				tp.FromAddress,
				tp.ContractAddress,
				amount.String(),
				caDec.String())
		}
		//判断一下手续费够不够
		fee, err := client.GetTrxBalance(tp.FromAddress)
		if err != nil {
			return nil, fmt.Errorf("get from %s chain fee balance error: %v", tp.FromAddress, err)
		}
		feeAmount := decimal.NewFromInt(fee.GetBalance())
		feeMin := decimal.NewFromFloat(1.5).Shift(6)
		if feeAmount.LessThan(feeMin) {
			return nil, fmt.Errorf("from=[%s] fee[%s] is less than 1.5 trx", tp.FromAddress, feeAmount.Shift(-6).String())
		}
		aTx, err = client.TransferTrc20(tp.FromAddress, tp.ToAddress, tp.ContractAddress, amount.BigInt(), tp.FeeLimit)
		if err != nil {
			return nil, fmt.Errorf("create trc20 tx error: %v,contract_address: %s", err, tp.ContractAddress)
		}
	} else if tp.AssetId != "" && tp.ContractAddress == "" {
		//trc10转账
		chainAmount, err := client.GetTrc10Balance(tp.FromAddress, tp.AssetId)
		if err != nil {
			return nil, fmt.Errorf("get trc10 %s chain balance error: %v", tp.AssetId, err)
		}
		caDec := decimal.NewFromInt(chainAmount)
		if amount.GreaterThan(caDec) {
			return nil, fmt.Errorf("[%s] amount is not engouth,asset_id=[%s],"+
				"transAmount=[%d],chainAmount=[%d]",
				tp.FromAddress,
				tp.AssetId,
				amount.IntPart(),
				caDec.IntPart())
		}
		aTx, err = client.TransferTrc10(tp.FromAddress, tp.ToAddress, tp.AssetId, amount.IntPart())
		if err != nil {
			return nil, fmt.Errorf("crete trc10 tx error: %v,asset_id: %s", err, tp.AssetId)
		}
	} else if tp.ContractAddress == "" && tp.AssetId == "" {
		//判断地址余额是否足够
		acc, err := client.GetTrxBalance(tp.FromAddress)
		if err != nil {
			return nil, fmt.Errorf("get trx amount error: %v", err)
		}
		chainAmount := acc.GetBalance()
		cd := decimal.NewFromInt(chainAmount)
		if amount.GreaterThanOrEqual(cd) {
			return nil, fmt.Errorf("[%s] amount is not enougth,transAmount=[%s],chainAnount=[%s]",
				tp.FromAddress,
				amount.String(),
				cd.String())
		}
		//trx 转账
		aTx, err = client.Transfer(tp.FromAddress, tp.ToAddress, amount.IntPart())
		if err != nil {
			return nil, fmt.Errorf("ctrate trx tx error: %v", err)
		}
	} else {
		return nil, errors.New("unknown transfer")
	}
	// 签名交易
	var hexPrivateKey string
	hexPrivateKey, err = cs.BaseService.addressOrPublicKeyToPrivate(tp.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("get private key error,Err=%v", err)
	}
	var tx *core.Transaction
	tx, err = sign.SignTransaction(aTx.Transaction, hexPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("sign transaction error: %v", err)
	}
	////广播交易
	err = client.BroadcastTransaction(tx)
	if tp.OuterOrderNo != "" {
		log.Infof("OuterOrderNo 不为空 = %s", tp.OuterOrderNo)
		cache, err := redis.Client.Get(redis.GetBroadcastOuterOrderNoKey(tp.OuterOrderNo))
		if err != nil {
			log.Infof("从redis获取广播订单KEY失败:%v", err)
		} else {
			if cache != "" {
				return nil, fmt.Errorf("订单: %s 已被广播，再次广播会造成重复出账", tp.OuterOrderNo)
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("broadcast tx error: %v", err)
	}
	txid := common.BytesToHexString(aTx.GetTxid())
	_ = hexPrivateKey
	if strings.HasPrefix(txid, "0x") {
		txid = strings.TrimPrefix(txid, "0x")
	}
	if tp.OuterOrderNo != "" {
		redis.Client.Set(redis.GetBroadcastOuterOrderNoKey(tp.OuterOrderNo), tp.OuterOrderNo, time.Hour*24)
	}

	log.Infof("send txid is: %s", txid)
	return txid, nil
}

/*
创建地址实体方法
*/
func (cs *TrxService) createAddressInfo() (util.AddrInfo, error) {
	wif, address := genkeys.GenerateKey()
	if wif == "" || address == "" {
		return util.AddrInfo{}, errors.New("wif or address is null")
	}
	return util.AddrInfo{
		PrivKey: wif,
		Address: address,
	}, nil
}
