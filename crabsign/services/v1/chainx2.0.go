package v1

import (
	"encoding/hex"
	"errors"
	"fmt"

	//"github.com/ChainSafe/gossamer/lib/crypto/sr25519"
	"github.com/JFJun/go-substrate-crypto/crypto"
	"github.com/JFJun/go-substrate-crypto/ss58"
	c "github.com/coldwallet-group/stafi-substrate-go/client"
	"github.com/coldwallet-group/stafi-substrate-go/expand"
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
type PcxService struct {
	*BaseService
	client   *c.Client
	url      string
	nonceCtl map[string]int64
	ed       *expand.MetadataExpand
}

const AddressType = 44

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
	////websocket 做假心跳处理，防止websocket断开
	//if strings.HasPrefix(ks.url, "ws") {
	//	err := ks.keepHeart()
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	return ks
}

//func (ks *PcxService) keepHeart() error {
//	if ks.client == nil {
//		client, err := c.New(ks.url)
//		if err != nil {
//			return err
//		}
//		client.SetPrefix(ss58.ChainXPrefix)
//		ks.client = client
//	}
//	go func() {
//		for true {
//			log.Println("发送心跳消息")
//			_, err := ks.client.C.RPC.State.GetRuntimeVersionLatest()
//			if err != nil {
//				log.Errorf("发送心跳消息失败：%v", err)
//			}
//			time.Sleep(time.Second * 10)
//		}
//	}()
//	return nil
//}

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
	return "不支持签名", nil
	//var tp model.PcxTransferParams
	//if err := ks.BaseService.parseData(req, &tp); err != nil {
	//	return nil, err
	//}
	//if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
	//	return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	//}
	//api, err := NewAzeroSubstrateAPI(conf.Config.PcxCfg.NodeUrl)
	//if err != nil {
	//	log.Infof("NewAzeroSubstrateAPI order: %s, err: %s", tp.OrderId, err.Error())
	//}
	//meta, err := api.RPC.State.GetMetadataLatest()
	//if err != nil {
	//	log.Infof("GetMetadataLatest order: %s, err: %s", tp.OrderId, err.Error())
	//	return "", err
	//}
	//
	//recvPub, err := ss58.DecodeToPub(tp.ToAddress)
	//if err != nil {
	//	log.Infof("DecodeToPub order: %s, err: %s", tp.OrderId, err.Error())
	//	return "", err
	//}
	//pub, err := sr25519.NewPublicKey(recvPub)
	//if err != nil {
	//	log.Infof("NewPublicKey order: %s, err: %s", tp.OrderId, err.Error())
	//	return "", err
	//}
	//bob := types.NewMultiAddressFromAccountID(pub.Encode())
	//// 1 unit of transfer
	//bal, ok := new(big.Int).SetString(tp.Amount, 10)
	//if !ok {
	//	log.Infof("NewMultiAddressFromAccountID order: %s, err: %s", tp.OrderId, "amount to big.int error")
	//	return "", err
	//}
	//c, err := types.NewCall(meta, "Balances.transfer", bob, types.NewUCompact(bal))
	//if err != nil {
	//	log.Infof("Balances.transfer order: %s, err: %s", tp.OrderId, err.Error())
	//	return "", err
	//}
	//// Create the extrinsic
	//ext := types.NewExtrinsic(c)
	//
	//genesisHash, err := api.RPC.Chain.GetBlockHash(0)
	//if err != nil {
	//	log.Infof("GetBlockHash order: %s, err: %s", tp.OrderId, err.Error())
	//	return "", err
	//}
	//
	//rv, err := api.RPC.State.GetRuntimeVersionLatest()
	//if err != nil {
	//	log.Infof("GetRuntimeVersionLatest order: %s, err: %s", tp.OrderId, err.Error())
	//	return "", err
	//}
	//
	//privateKey, err := ks.BaseService.GetKeyByAddress(tp.FromAddress)
	//if err != nil {
	//	log.Infof("GetKeyByAddress order: %s, err: %s", tp.OrderId, err.Error())
	//	return "", err
	//}
	//mn := privateKey
	//p, err := signature.KeyringPairFromSecret(mn, 42)
	//if err != nil {
	//	log.Infof("KeyringPairFromSecret order: %s, err: %s", tp.OrderId, err.Error())
	//	return "", err
	//}
	//
	//accountInfo, err := pcxGetAccountInfo(tp.FromAddress)
	//if err != nil {
	//	log.Infof("pcxGetAccountInfo order: %s, err: %s", tp.OrderId, err.Error())
	//	return "", err
	//}
	//nonce := accountInfo.Nonce
	//balance := accountInfo.Data.Free
	//
	//amountD, err := decimal.NewFromString(tp.Amount)
	//if err != nil {
	//	log.Infof("amountDNewFromString order: %s, err: %s", tp.OrderId, err.Error())
	//	return "", err
	//}
	//
	//if amountD.Cmp(decimal.NewFromFloat(0.1)) < 0 {
	//	log.Infof("order: %s, err: 最小发送金额为0.1", tp.OrderId)
	//	return "", errors.New("最小发送金额为0.1")
	//}
	//
	//if balance.Cmp(amountD) < 1 {
	//	log.Infof("order: %s, err: 余额不足", tp.OrderId)
	//	return "", errors.New("余额不足")
	//}
	//
	//o := types.SignatureOptions{
	//	BlockHash:          genesisHash,
	//	Era:                types.ExtrinsicEra{IsMortalEra: false},
	//	GenesisHash:        genesisHash,
	//	Nonce:              types.NewUCompactFromUInt(uint64(nonce)),
	//	SpecVersion:        rv.SpecVersion,
	//	Tip:                types.NewUCompactFromUInt(100),
	//	TransactionVersion: rv.TransactionVersion,
	//}
	//
	//// Sign the transaction using Alice's default account
	//err = ext.Sign(p, o)
	//if err != nil {
	//	log.Infof("Sign order: %s, err: %s", tp.OrderId, "error or not ok")
	//	return "", err
	//}
	//log.Infof("SignatureOptions %+v", o)
	//log.Infof("sign %+v", ext)
	//enc, err := types.EncodeToHexString(ext)
	//if err != nil {
	//	return "", err
	//}
	//return enc, nil
}

func (ks *PcxService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	aInfo, err := pcxGetAccountInfo(req.Address)
	if err != nil {
		return nil, fmt.Errorf("get cring address amount error,err=%v", err)
	}
	return aInfo.Data.Free.String(), nil
}
func (ks *PcxService) ValidAddress(address string) error {
	return ss58.VerityAddress(address, ss58.ChainXPrefix)
}

// TransferService  热钱包出账服务
func (ks *PcxService) TransferService(req interface{}) (interface{}, error) {
	var tp model.PcxTransferParams
	if err := ks.BaseService.parseData(req, &tp); err != nil {
		return nil, err
	}
	if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	}
	api, err := NewAzeroSubstrateAPI(conf.Config.PcxCfg.NodeUrl)
	if err != nil {
		log.Infof("NewAzeroSubstrateAPI order: %s, err: %s", tp.OrderId, err.Error())
	}
	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		log.Infof("GetMetadataLatest order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}

	recvPub, err := ss58.DecodeToPub(tp.ToAddress)
	if err != nil {
		log.Infof("DecodeToPub order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}
	//pub, err := sr25519.NewPublicKey(recvPub)
	//if err != nil {
	//	log.Infof("NewPublicKey order: %s, err: %s", tp.OrderId, err.Error())
	//	return "", err
	//}
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
		log.Infof("GetBlockHash order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}

	rv, err := api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		log.Infof("GetRuntimeVersionLatest order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}

	privateKey, err := ks.BaseService.GetKeyByAddress(tp.FromAddress)
	if err != nil {
		log.Infof("GetKeyByAddress order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}
	mn := privateKey
	var p signature.KeyringPair
	p, err = signature.KeyringPairFromSecretEd25519(mn, 44) //这个密钥怎么会是44的
	if err != nil {
		log.Infof("KeyringPairFromSecret order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}

	scanUrl := "https://chainx.api.subscan.io"
	scanKey := "494f2c39fa73f17cc38104f7e1cd4841"
	apis := util.NewScanApi(scanUrl, scanKey)
	balance, nonce, err := apis.AccountInfo(tp.FromAddress)
	if err != nil {
		log.Infof("order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}
	//这个不知道行不行,密钥方式不对了,不是默认的那种了
	//accountInfo, err := pcxGetAccountInfo(tp.FromAddress)
	//if err != nil {
	//	log.Infof("pcxGetAccountInfo order: %s, err: %s", tp.OrderId, err.Error())
	//	return "", err
	//}
	//nonce := accountInfo.Nonce
	//balance := accountInfo.Data.Free

	amountD, err := decimal.NewFromString(tp.Amount)
	if err != nil {
		log.Infof("amountDNewFromString order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}

	//todo 最小发送金额判断不对, 需要加一下精度
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

	err = ext.SignEd25519(p, o)
	if err != nil {
		log.Infof("Sign order: %s, err: %s", tp.OrderId, "error or not ok")
		return "", err
	}
	log.Infof("SignatureOptions %+v", o)
	log.Infof("sign %+v", ext)
	enc := ""
	//Send the extrinsic
	hash, err := api.RPC.Author.SubmitExtrinsic(ext)
	if err != nil {
		enc, _ = types.EncodeToHexString(ext)
		log.Infof("SubmitExtrinsic order: %s ,%s, err: %s,16进制的那串交易原文: %s", tp.OrderId, "SubmitExtrinsic error or not ok", err.Error(), enc)
		return "", err
	}
	enc, _ = types.EncodeToHexString(ext)
	log.Infof("order: %s, hash: %s,16进制的那串交易原文: %s", tp.OrderId, hash.Hex(), enc)
	return hash.Hex(), nil
}

/*
创建地址实体方法
*/
func (ks *PcxService) createAddressInfo() (util.AddrInfo, error) {
	var addrInfo util.AddrInfo

	//priv, pub, err := crypto.GenerateSubstrateKey(crypto.Sr25519Type)
	priv, pub, err := crypto.GenerateSubstrateKey(crypto.Ed25519Type) //这个是线上的密钥格式
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
func pcxGetAccountInfo(fromAddress string) (AccountInfo, error) {
	accountPubKey := fmt.Sprintf("0x%s", pcxDecode(fromAddress, AddressType)) //获取公钥
	websocket.SetEndpoint(conf.Config.PcxCfg.WsUrl)
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

func pcxDecode(address string, addressType int) string {
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
