package v1

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/JFJun/go-substrate-crypto/crypto"
	"github.com/JFJun/go-substrate-crypto/ss58"
	"github.com/prometheus/common/log"

	"github.com/shopspring/decimal"
	"wallet-sign/conf"
	"wallet-sign/model"
	"wallet-sign/util"

	types2 "github.com/yanyushr/go-substrate-rpc-client/v3/types"

	"github.com/yanyushr/go-substrate-rpc-client/v3/client"
	"github.com/yanyushr/go-substrate-rpc-client/v3/rpc"
	"wallet-sign/client/dot"
)

type SubstrateAPI struct {
	RPC    *rpc.RPC
	Client client.Client
}

func NewSubstrateAPI(url string) (*SubstrateAPI, error) {
	cl, err := client.Connect(url)
	if err != nil {
		return nil, err
	}

	newRPC, err := rpc.NewRPC(cl)
	if err != nil {
		return nil, err
	}

	return &SubstrateAPI{
		RPC:    newRPC,
		Client: cl,
	}, nil
}

/*
service模板
*/

/*
币种服务结构体
*/
type DotService struct {
	*BaseService
	client *dot.Client
	url    string
	meta   *types2.MetadataV14
}

/*
初始化币种服务
	注意：
		方法接受者： BaseService
		方法命名： 币种大写 + Service
*/
func (bs *BaseService) DOTService() *DotService {
	ks := new(DotService)
	ks.BaseService = bs
	ks.url = conf.Config.DotCfg.NodeUrl
	return ks
}

/*
接口创建地址服务
	无需改动
*/
func (ks *DotService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return ks.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, ks.createAddressInfo)
	}
	return ks.BaseService.createAddress(req, ks.createAddressInfo)
}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (ks *DotService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	fmt.Println("start create Dot address")
	_, err := ks.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, ks.createAddressInfo)
	return err
}

/*
签名服务
*/
func (ks *DotService) SignService(req *model.ReqSignParams) (interface{}, error) {
	var tp model.DotColdParams
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

func (ks *DotService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {

	cli, err := dot.NewClient(ks.url)
	if err != nil {
		return nil, err
	}

	meta, err := cli.Api.RPC.State.GetMetadataLatest()
	if err != nil {
		return nil, err
	}

	ai, err := cli.GetAccountInfo(req.Address, meta)
	if err != nil {
		return nil, fmt.Errorf("get dot amount error: %v", err)
	}
	return ai.Data.Free.String(), nil
}

func (ks *DotService) ValidAddress(address string) error {
	return ss58.VerityAddress(address, ss58.PolkadotPrefix)
}

/*
热钱包出账服务
*/
func (ks *DotService) TransferService(req interface{}) (interface{}, error) {
	if ks.client == nil {
		//初始化客户端
		client, err := dot.NewClient(ks.url)
		if err != nil {
			log.Info("err:", err)
			return nil, err
		}
		ks.client = client
	}

	var tp model.DotTransferParams
	if err := ks.BaseService.parseData(req, &tp); err != nil {
		return nil, err
	}

	if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	}

	cli, err := dot.NewClient(ks.url)
	if err != nil {
		log.Info("err:", err)
		return nil, err
	}

	meta, err := cli.Api.RPC.State.GetMetadataLatest()
	if err != nil {
		log.Info("err:", err)
		return nil, err
	}

	params := dot.Txparam{
		MchName:     tp.MchId,
		FromAddress: tp.FromAddress,
		ToAddress:   tp.ToAddress,
		Amount:      decimal.RequireFromString(tp.Amount),
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

	accounrInfo, err := cli.GetAccountInfo(params.FromAddress, meta)
	if err != nil {
		return nil, err
	}

	// if accounrInfo.Data.Free.Int == nil || !(accounrInfo.Data.Free.Int.Cmp(params.Amount.BigInt()) > 0) {
	// 	return nil, errors.New("资金不足")
	// }

	params.Nonce = uint64(accounrInfo.Nonce)
	if params.Nonce == 0 {
		params.Nonce = uint64(accounrInfo.Nonce)
	}

	privateKey, err := ks.BaseService.GetKeyByAddress(params.FromAddress)
	if err != nil {
		return nil, err
	}

	rawTx, err := dot.SignTx(&params, privateKey)
	if err != nil {
		return nil, err
	}

	var result interface{}

	fmt.Println("**********rawTx:", rawTx)

	err = cli.Api.Client.Call(&result, "author_submitExtrinsic", rawTx)
	if err != nil {
		return nil, err
	}

	txid := result.(string)
	return txid, nil
}

/*
创建地址实体方法
*/
func (ks *DotService) createAddressInfo() (util.AddrInfo, error) {
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
	address, err = crypto.CreateSubstrateAddress(pub, ss58.PolkadotPrefix)
	if err != nil {
		return addrInfo, err
	}
	addrInfo.PrivKey = "0x" + wif
	addrInfo.Address = address
	return addrInfo, nil
}
