package v1

import "C"
import (
	"errors"
	"fmt"
	"github.com/JFJun/go-substrate-crypto/ss58"
	"github.com/coldwallet-group/bifrost-go/client"
	substrateConf "github.com/coldwallet-group/substrate-go/config"
	s25519 "github.com/coldwallet-group/substrate-go/sr25519"
	sutil "github.com/itering/subscan/util"
	"github.com/itering/subscan/util/base58"
	"github.com/itering/substrate-api-rpc/metadata"
	"github.com/itering/substrate-api-rpc/rpc"
	"github.com/itering/substrate-api-rpc/websocket"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/blake2b"
	"math/big"
	"wallet-sign/conf"
	"wallet-sign/model"
	"wallet-sign/sign/signature"
	"wallet-sign/sign/types"
	"wallet-sign/util"
)

/*
service模板
*/

/*
币种服务结构体
*/

type CRABService struct {
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
func (bs *BaseService) CRABService() *CRABService {
	ks := new(CRABService)
	ks.BaseService = bs
	ks.url = conf.Config.CrabCfg.NodeUrl
	ks.nonceCtl = make(map[string]int64)
	return ks
}

/*
接口创建地址服务
	无需改动
*/
func (ks *CRABService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return ks.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, ks.createAddressInfo)
	}
	return ks.BaseService.createAddress(req, ks.createAddressInfo)
}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (ks *CRABService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	fmt.Println("start create Cring address")
	_, err := ks.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, ks.createAddressInfo)
	return err
}

func (ks *CRABService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	aInfo, err := getAccountInfo(req.Address)
	if err != nil {
		return nil, fmt.Errorf("get cring address amount error,err=%v", err)
	}
	return aInfo.Data.Free.String(), nil
}

func (ks *CRABService) ValidAddress(address string) error {
	return ss58.VerityAddress(address, ss58.SubstratePrefix)
}

/*
签名服务
*/
func (ks *CRABService) SignService(req *model.ReqSignParams) (interface{}, error) {
	//var tp model.CringColdParams
	//if err := ks.BaseService.parseData(req, &tp); err != nil {
	//
	//	return nil, err
	//}
	//if &tp == nil {
	//	return nil, errors.New("transfer params is null")
	//}
	//if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
	//	return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	//}
	//if tp.GenesisHash == "" {
	//	return nil, fmt.Errorf("params is null,genesisHash=[%s]", tp.GenesisHash)
	//}
	//amount, err := decimal.NewFromString(tp.Amount)
	//if err != nil {
	//	return nil, err
	//}
	//transaction := tx.NewSubstrateTransaction(tp.FromAddress, tp.Nonce)
	////5. 初始化metadata的扩张结构
	//ed, err := expand.NewMetadataExpand(ks.client.Meta)
	//if err != nil {
	//	return nil, err
	//}
	////6. 初始化Balances.transfer的call方法
	//call, err := ed.BalanceTransferCall(tp.ToAddress, amount.BigInt())
	//if err != nil {
	//	return nil, err
	//}
	///*
	//	//Balances.transfer_keep_alive  call方法
	//	btkac,err:=ed.BalanceTransferKeepAliveCall(to,amount)
	//*/
	//
	///*
	//	toAmount:=make(map[string]uint64)
	//	toAmount[to] = amount
	//	//...
	//	//true: user Balances.transfer_keep_alive  false: Balances.transfer
	//	ubtc,err:=ed.UtilityBatchTxCall(toAmount,false)
	//*/
	//
	////7. 设置交易的必要参数
	//transaction.SetGenesisHashAndBlockHash(ks.client.GetGenesisHash(), ks.client.GetGenesisHash()).
	//	SetSpecAndTxVersion(uint32(ks.client.SpecVersion), uint32(ks.client.TransactionVersion)).
	//	SetCall(call) //设置call
	//privateKey, err := ks.addressOrPublicKeyToPrivate(tp.FromAddress)
	//if err != nil {
	//	return nil, err
	//}
	////8. 签名交易
	//sig, err := transaction.SignTransaction(privateKey, crypto.Sr25519Type)
	//if err != nil {
	//	return nil, err
	//}
	return "sig", nil
}

/*
热钱包出账服务
*/
func (ks *CRABService) TransferService(req interface{}) (interface{}, error) {
	var tp model.CringTransferParams
	if err := ks.BaseService.parseData(req, &tp); err != nil {
		return nil, err
	}
	if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	}
	api, err := NewAzeroSubstrateAPI(conf.Config.CrabCfg.NodeUrl)
	if err != nil {
		log.Infof("order: %s, err: %s", tp.OrderId, err.Error())
	}
	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		log.Infof("order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}

	recvPub, err := ss58.DecodeToPub(tp.ToAddress)
	if err != nil {
		log.Infof("DecodeToPub order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}
	bob := types.NewMultiAddressFromAccountID(recvPub)
	// 1 unit of transfer
	bal, ok := new(big.Int).SetString(tp.Amount, 10)
	if !ok {
		log.Infof("NewMultiAddressFromAccountID order: %s, err: %s", tp.OrderId, "amount to big.int error")
		return "", err
	}
	c, err := types.NewCall(meta, "Balances.transfer", bob, types.NewUCompact(bal))
	if err != nil {
		log.Infof("Balances.transfer order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}
	// Create the extrinsic
	ext := types.NewExtrinsic(c)

	genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	if err != nil {
		log.Infof("order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		log.Infof("order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}

	privateKey, err := ks.BaseService.GetKeyByAddress(tp.FromAddress)
	if err != nil {
		log.Infof("order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}
	mn := privateKey
	p, err := signature.KeyringPairFromSecret(mn, 42)
	if err != nil {
		log.Infof("order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}

	accountInfo, err := getAccountInfo(tp.FromAddress)
	if err != nil {
		log.Infof("order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}
	nonce := accountInfo.Nonce
	balance := accountInfo.Data.Free
	if err != nil {
		log.Infof("order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}

	amountD, err := decimal.NewFromString(tp.Amount)
	if err != nil {
		log.Infof("order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}

	if amountD.Cmp(decimal.NewFromFloat(0.1)) < 0 {
		log.Infof("order: %s, err: 最小发送金额为0.1", tp.OrderId)
		return "", errors.New("最小发送金额为0.1")
	}

	if balance.Cmp(amountD) < 1 {
		log.Infof("order: %s, err: 余额不足", tp.OrderId)
		return "", errors.New("余额不足")
	}

	o := types.SignatureOptions{
		BlockHash:          genesisHash,
		Era:                types.ExtrinsicEra{IsMortalEra: false},
		GenesisHash:        genesisHash,
		Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
		SpecVersion:        rv.SpecVersion,
		Tip:                types.NewUCompactFromUInt(100),
		TransactionVersion: rv.TransactionVersion,
	}

	// Sign the transaction using Alice's default account
	err = ext.Sign(p, o)
	if err != nil {
		log.Infof("order: %s, err: %s", tp.OrderId, "GetStorageLatest error or not ok")
		return "", err
	}

	//Send the extrinsic
	hash, err := api.RPC.Author.SubmitExtrinsic(ext)
	if err != nil {
		log.Infof("order: %s, err: %s", tp.OrderId, "GetStorageLatest error or not ok")
		return "", err
	}
	log.Infof("order: %s, hash: %s", tp.OrderId, hash.Hex())
	return hash.Hex(), nil
}

/*
创建地址实体方法
*/
func (ks *CRABService) createAddressInfo() (util.AddrInfo, error) {
	var addrInfo util.AddrInfo
	priv, pub, err := s25519.GenerateKey()
	if err != nil {
		return addrInfo, err
	}
	secret, err1 := s25519.PrivateKeyToHex(priv)
	if err1 != nil {
		return addrInfo, err1
	}
	address, err2 := s25519.CreateAddress(pub, substrateConf.SubstratePrefix)
	if err2 != nil {
		return addrInfo, err2
	}
	addrInfo.PrivKey = secret
	addrInfo.Address = address
	return addrInfo, nil
}

type AccountInfo struct {
	Nonce int `json:"nonce"`
	Data  struct {
		Free decimal.Decimal `json:"free"`
	} `json:"data"`
}

func getAccountInfo(fromAddress string) (AccountInfo, error) {
	accountPubKey := fmt.Sprintf("0x%s", Decode(fromAddress, 42)) //获取公钥
	websocket.SetEndpoint(conf.Config.CrabCfg.WsUrl)
	rawData := "" // rpc state_getMetadata
	coded, err := rpc.GetMetadataByHash(nil)
	if err == nil {
		rawData = coded
	} else {
		return AccountInfo{}, err
	}
	metadata.Latest(&metadata.RuntimeRaw{Spec: 0, Raw: rawData})
	accountDataRaw, err := rpc.ReadStorage(nil, "system", "account", "", accountPubKey)
	if err != nil {
		return AccountInfo{}, err
	}
	var data AccountInfo
	accountDataRaw.ToAny(&data)
	return data, nil
}
func Decode(address string, addressType int) string {
	checksumPrefix := []byte("SS58PRE")
	ss58Format := base58.Decode(address)
	if len(ss58Format) == 0 || ss58Format[0] != byte(addressType) {
		return ""
	}
	var checksumLength int
	if sutil.IntInSlice(len(ss58Format), []int{3, 4, 6, 10}) {
		checksumLength = 1
	} else if sutil.IntInSlice(len(ss58Format), []int{5, 7, 11, 35}) {
		checksumLength = 2
	} else if sutil.IntInSlice(len(ss58Format), []int{8, 12}) {
		checksumLength = 3
	} else if sutil.IntInSlice(len(ss58Format), []int{9, 13}) {
		checksumLength = 4
	} else if sutil.IntInSlice(len(ss58Format), []int{14}) {
		checksumLength = 5
	} else if sutil.IntInSlice(len(ss58Format), []int{15}) {
		checksumLength = 6
	} else if sutil.IntInSlice(len(ss58Format), []int{16}) {
		checksumLength = 7
	} else if sutil.IntInSlice(len(ss58Format), []int{17}) {
		checksumLength = 8
	} else {
		return ""
	}
	bss := ss58Format[0 : len(ss58Format)-checksumLength]
	checksum, _ := blake2b.New(64, []byte{})
	w := append(checksumPrefix[:], bss[:]...)
	_, err := checksum.Write(w)
	if err != nil {
		return ""
	}

	h := checksum.Sum(nil)
	if sutil.BytesToHex(h[0:checksumLength]) != sutil.BytesToHex(ss58Format[len(ss58Format)-checksumLength:]) {
		return ""
	}
	return sutil.BytesToHex(ss58Format[1 : len(ss58Format)-checksumLength])
}
