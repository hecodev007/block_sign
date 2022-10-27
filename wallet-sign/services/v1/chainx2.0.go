package v1

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/JFJun/go-substrate-crypto/crypto"
	"github.com/JFJun/go-substrate-crypto/ss58"
	c "github.com/coldwallet-group/stafi-substrate-go/client"
	"github.com/coldwallet-group/stafi-substrate-go/expand"
	"github.com/coldwallet-group/stafi-substrate-go/tx"
	"github.com/group-coldwallet/wallet-sign/conf"
	"github.com/group-coldwallet/wallet-sign/model"
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
type PcxService struct {
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
func (bs *BaseService) PCXService() *PcxService {
	ks := new(PcxService)
	ks.BaseService = bs
	ks.url = conf.Config.PcxCfg.NodeUrl
	ks.nonceCtl = make(map[string]int64)
	//websocket 做假心跳处理，防止websocket断开
	if strings.HasPrefix(ks.url, "ws") {
		err := ks.keepHeart()
		if err != nil {
			panic(err)
		}
	}
	return ks
}

func (ks *PcxService) keepHeart() error {
	if ks.client == nil {
		client, err := c.New(ks.url)
		if err != nil {
			return err
		}
		client.SetPrefix(ss58.ChainXPrefix)
		ks.client = client
	}
	go func() {
		for true {
			log.Println("发送心跳消息")
			_, err := ks.client.C.RPC.State.GetRuntimeVersionLatest()
			if err != nil {
				log.Errorf("发送心跳消息失败：%v", err)
			}
			time.Sleep(time.Second * 10)
		}
	}()
	return nil
}

/*
接口创建地址服务
	无需改动
*/
func (ks *PcxService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return ks.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, ks.createAddressInfo)
	}
	return ks.BaseService.createAddress(req, ks.createAddressInfo)
}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (ks *PcxService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	fmt.Println("start create Pcx address")
	_, err := ks.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, ks.createAddressInfo)
	return err
}

/*
签名服务
*/
func (ks *PcxService) SignService(req *model.ReqSignParams) (interface{}, error) {
	return nil, errors.New("unsupport cold sign")
}

func (ks *PcxService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	ai, err := ks.client.GetAccountInfo(req.Address)
	if err != nil {
		return nil, fmt.Errorf("get pcx address amount error,err=%v", err)
	}

	return ai.Data.Free.String(), nil
}
func (ks *PcxService) ValidAddress(address string) error {
	return ss58.VerityAddress(address, ss58.ChainXPrefix)
}

/*
热钱包出账服务
*/
func (ks *PcxService) TransferService(req interface{}) (interface{}, error) {
	if ks.client == nil {
		//初始化客户端
		client, err := c.New(ks.url)
		if err != nil {
			return nil, err
		}
		client.SetPrefix(ss58.ChainXPrefix)
		ks.client = client

	}
	if ks.ed == nil {
		ks.ed, _ = expand.NewMetadataExpand(ks.client.Meta)
	}
	var tp model.PcxTransferParams
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
	var (
		sig    string
		errSig error
	)
	if in_array(tp.FromAddress, []string{"5RCQeCPaCxumnkKTw4Yry4tU3nQDt6UERk4Arp4QA12xFLho", "5TBDRF5Qw6GXS1t5Gn9nbgdFykrvPS4gaN9K9eZE841cpiMg",
		"5TvYpyq8wkE4amDwSvKYzyyb2NogRcJuiTC3x6xuUPsm4KVw", "5QiXSvfjh8EGpiVY34EMXfd1NaKFsPBYAeFfL48af96CD6gj"}) {
		sig, errSig = originTx.SignTransaction(privateKey, crypto.Ed25519Type)
	} else {
		sig, errSig = originTx.SignTransaction(privateKey, crypto.Sr25519Type)
	}

	if errSig != nil {
		return nil, fmt.Errorf("sign error,Err=[%v]", errSig)
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
func (ks *PcxService) createAddressInfo() (util.AddrInfo, error) {
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
	address, err = crypto.CreateSubstrateAddress(pub, ss58.ChainXPrefix)
	if err != nil {
		return addrInfo, err
	}
	addrInfo.PrivKey = "0x" + wif
	addrInfo.Address = address
	return addrInfo, nil
}
func in_array(need string, needArr []string) bool {
	for _, v := range needArr {
		if need == v {
			return true
		}
	}
	return false
}
