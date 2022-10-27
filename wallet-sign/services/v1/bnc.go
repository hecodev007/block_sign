package v1

import (
	"encoding/hex"
	"errors"
	"fmt"
	c "github.com/JFJun/bifrost-go/client"
	"github.com/JFJun/bifrost-go/expand"
	"github.com/JFJun/bifrost-go/tx"
	"github.com/JFJun/go-substrate-crypto/crypto"
	"github.com/JFJun/go-substrate-crypto/ss58"
	"github.com/group-coldwallet/wallet-sign/conf"
	"github.com/group-coldwallet/wallet-sign/model"
	"github.com/group-coldwallet/wallet-sign/util"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/stafiprotocol/go-substrate-rpc-client/types"
	"strings"
	"time"
)

/*
service模板
*/

/*
币种服务结构体
*/
type BncService struct {
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
func (bs *BaseService) BNCService() *BncService {
	ks := new(BncService)
	ks.BaseService = bs
	ks.url = conf.Config.NodeUrl
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

func (ks *BncService) keepHeart() error {
	if ks.client == nil {
		client, err := c.New(ks.url)
		if err != nil {
			return err
		}
		client.SetPrefix(ss58.BifrostPrefix)
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
func (ks *BncService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return ks.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, ks.createAddressInfo)
	}
	return ks.BaseService.createAddress(req, ks.createAddressInfo)
}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (ks *BncService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	fmt.Println("start create Bnc address")
	_, err := ks.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, ks.createAddressInfo)
	return err
}

/*
签名服务
*/
func (ks *BncService) SignService(req *model.ReqSignParams) (interface{}, error) {
	return nil, errors.New("unsupport cold sign")
}

/*
热钱包出账服务
*/
func (ks *BncService) TransferService(req interface{}) (interface{}, error) {
	if ks.client == nil {
		//初始化客户端
		client, err := c.New(ks.url)
		if err != nil {
			return nil, err
		}
		client.SetPrefix(ss58.BifrostPrefix)
		ks.client = client

	}
	if ks.ed == nil {
		ks.ed, _ = expand.NewMetadataExpand(ks.client.Meta)
	}
	var tp model.BncTransferParams
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

	types.SetSerDeOptions(types.SerDeOptions{NoPalletIndices: true})
	originTx := tx.NewSubstrateTransaction(tp.FromAddress, uint64(uint32(ai.Nonce)))
	ed, err := expand.NewMetadataExpand(ks.client.Meta)
	if err != nil {
		return nil, fmt.Errorf("get metadata expand error: %v", err)
	}
	call, err := ed.BalanceTransferKeepAliveCall(tp.ToAddress, uint64(amount.IntPart()))
	if err != nil {
		return nil, fmt.Errorf("get Balances.transfer call error: %v", err)
	}

	originTx.SetGenesisHashAndBlockHash(ks.client.GetGenesisHash(), ks.client.GetGenesisHash()).
		SetSpecVersionAndCallId(uint32(ks.client.SpecVersion), uint32(ks.client.TransactionVersion)).
		SetCall(call)
	//获取私钥
	privateKey, err2 := ks.addressOrPublicKeyToPrivate(tp.FromAddress)
	if err2 != nil {
		return nil, err2
	}
	var (
		sig    string
		errSig error
	)
	sig, errSig = originTx.SignTransaction(privateKey, crypto.Sr25519Type)
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
func (ks *BncService) createAddressInfo() (util.AddrInfo, error) {
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
	address, err = crypto.CreateSubstrateAddress(pub, ss58.BifrostPrefix)
	if err != nil {
		return addrInfo, err
	}
	addrInfo.PrivKey = "0x" + wif
	addrInfo.Address = address
	return addrInfo, nil
}

func (ks *BncService) ValidAddress(address string) error {
	return ss58.VerityAddress(address, ss58.BifrostPrefix)
}

func (ks *BncService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	ai, err := ks.client.GetAccountInfo(req.Address)
	if err != nil {
		return nil, fmt.Errorf("get bnc amount error: %v", err)
	}
	return ai.Data.Free.String(), nil
}
