package v1

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ChainSafe/gossamer/lib/crypto/sr25519"
	"github.com/JFJun/go-substrate-crypto/crypto"
	"github.com/JFJun/go-substrate-crypto/ss58"
	types2 "github.com/yanyushr/go-substrate-rpc-client/v3/types"
	"math/big"
	"wallet-sign/sign/signature"

	"wallet-sign/sign/client"
	"wallet-sign/sign/rpc"

	//"github.com/JFJun/go-substrate-crypto/crypto"
	//"github.com/JFJun/go-substrate-crypto/ss58"
	"github.com/prometheus/common/log"

	"github.com/shopspring/decimal"
	"wallet-sign/conf"
	"wallet-sign/model"
	"wallet-sign/util"

	//types2 "github.com/yanyushr/go-substrate-rpc-client/v3/types"

	"wallet-sign/client/dot"
	"wallet-sign/sign/types"
)

/*
service模板
*/

/*
币种服务结构体
*/
type AzeroService struct {
	*BaseService
	client   *dot.Client
	url      string
	meta     *types2.MetadataV14
	AzeroApi *AzeroSubstrateAPI
}

/*
初始化币种服务
	注意：
		方法接受者： BaseService
		方法命名： 币种大写 + Service
*/
func (bs *BaseService) AZEROService() *AzeroService {
	ks := new(AzeroService)
	ks.BaseService = bs
	ks.url = conf.Config.AzeroCfg.WsUrl
	azeroApi, err := NewAzeroSubstrateAPI(conf.Config.AzeroCfg.WsUrl)
	if err != nil {
		panic(err)
	}
	ks.AzeroApi = azeroApi
	return ks
}

type AzeroSubstrateAPI struct {
	RPC     *rpc.RPC
	Client  client.Client
	ScanApi *util.ScanApi
}

func NewAzeroSubstrateAPI(url string) (*AzeroSubstrateAPI, error) {
	cl, err := client.Connect(url)
	if err != nil {
		return nil, err
	}

	newRPC, err := rpc.NewRPC(cl)
	if err != nil {
		return nil, err
	}

	ScanApi := util.NewScanApi(conf.Config.AzeroCfg.ScanUrl, conf.Config.AzeroCfg.ScanKey)

	return &AzeroSubstrateAPI{
		RPC:     newRPC,
		Client:  cl,
		ScanApi: ScanApi,
	}, nil
}

func (ks *AzeroService) transfer(tp *model.AzeroTransferParams) (string, error) {
	//api := ks.AzeroApi
	api, err := NewAzeroSubstrateAPI(conf.Config.AzeroCfg.WsUrl)
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
		log.Infof("order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}
	pub, err := sr25519.NewPublicKey(recvPub)
	if err != nil {
		log.Infof("order: %s, err: %s", tp.OrderId, err.Error())
		return "", err
	}
	bob := types.NewMultiAddressFromAccountID(pub.Encode())
	// 1 unit of transfer
	bal, ok := new(big.Int).SetString(tp.Amount, 10)
	if !ok {
		log.Infof("order: %s, err: %s", tp.OrderId, "amount to big.int error")
		return "", err
	}
	c, err := types.NewCall(meta, "Balances.transfer", bob, types.NewUCompact(bal))
	if err != nil {
		log.Infof("order: %s, err: %s", tp.OrderId, err.Error())
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

	balance, nonce, err := api.ScanApi.AccountInfo(tp.FromAddress)
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
接口创建地址服务
	无需改动
*/
func (ks *AzeroService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return ks.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, ks.createAddressInfo)
	}
	return ks.BaseService.createAddress(req, ks.createAddressInfo)
}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (ks *AzeroService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	fmt.Println("start create Azero address")
	_, err := ks.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, ks.createAddressInfo)
	return err
}

/*
签名服务
*/
func (ks *AzeroService) SignService(req *model.ReqSignParams) (interface{}, error) {
	var tp model.AzeroColdParams
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

	cli, err := dot.NewClient(ks.url)
	if err != nil {
		return nil, err
	}

	meta, err := cli.Api.RPC.State.GetMetadataLatest()
	if err != nil {
		return nil, err
	}

	params := dot.Txparam{
		FromAddress: tp.FromAddress,
		ToAddress:   tp.ToAddress,
		Amount:      amount,
	}

	params.Meta = meta

	params.GenesisHash = cli.GetGenesisHash().Hex()
	params.BlockHash = params.GenesisHash

	runVer, err := cli.Api.RPC.State.GetRuntimeVersionLatest()
	if err != nil {
		return nil, err
	}
	params.TransactionVersion = uint32(runVer.TransactionVersion)
	params.SpecVersion = uint32(runVer.SpecVersion)

	lastBlockHash, err := cli.Api.RPC.Chain.GetBlockHashLatest()
	if err != nil {
		return nil, err
	}

	lastBlock, err := cli.Api.RPC.Chain.GetBlock(lastBlockHash)
	if err != nil {
		return nil, err
	}
	params.BlockNumber = uint64(lastBlock.Block.Header.Number)

	privateKey, err := ks.BaseService.GetKeyByAddress(params.FromAddress)
	if err != nil {
		return nil, err
	}

	sigRaw, err := dot.SignTx(&params, privateKey)
	if err != nil {
		return nil, err
	}

	return sigRaw, nil
}

func (ks *AzeroService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	balance, _, err := ks.AzeroApi.ScanApi.AccountInfo(req.Address)
	if err != nil {
		return nil, err
	}
	return balance.String(), nil
	//cli, err := dot.NewClient(conf.Config.AzeroCfg.RpcUrl)
	//if err != nil {
	//	return nil, err
	//}
	//meta, err := cli.Api.RPC.State.GetMetadataLatest()
	//if err != nil {
	//	return nil, err
	//}
	//ai, err := cli.GetAccountInfo(req.Address, meta)
	//if err != nil {
	//	return nil, fmt.Errorf("get dot amount error: %v", err)
	//}
	//return ai.Data.Free.String(), nil
}

func (ks *AzeroService) ValidAddress(address string) error {
	return ss58.VerityAddress(address, ss58.SubstratePrefix)
}

/*
热钱包出账服务 1.1.1.1
*/
func (ks *AzeroService) TransferService(req interface{}) (interface{}, error) {
	var tp model.AzeroTransferParams
	if err := ks.BaseService.parseData(req, &tp); err != nil {
		return nil, err
	}
	if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	}

	hash, err := ks.transfer(&tp)
	return hash, err
}

/*
创建地址实体方法
*/
func (ks *AzeroService) createAddressInfo() (util.AddrInfo, error) {
	var addrInfo util.AddrInfo

	priv, _, err := crypto.GenerateSubstrateKey(crypto.Sr25519Type)
	if err != nil {
		return addrInfo, err
	}
	if len(priv) != 32 {
		return addrInfo, errors.New("private key length is not equal 32")
	}
	wif := "0x" + hex.EncodeToString(priv)

	address, err := getAddr(wif)
	if err != nil {
		return addrInfo, err
	}
	addrInfo.PrivKey = wif
	addrInfo.Address = address
	return addrInfo, nil
}

func getAddr(word string) (string, error) {
	p, err := signature.KeyringPairFromSecret(word, 42)
	if err != nil {
		return "", err
	}
	return p.Address, nil
}
