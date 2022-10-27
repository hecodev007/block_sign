package v1

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/JFJun/go-substrate-crypto/ss58"
	"github.com/group-coldwallet/wallet-sign/client/ksm"
	"github.com/group-coldwallet/wallet-sign/conf"
	"github.com/group-coldwallet/wallet-sign/model"
	"github.com/group-coldwallet/wallet-sign/util"
	"github.com/prometheus/common/log"
	"github.com/shopspring/decimal"
)

/*
service模板
*/

/*
币种服务结构体
*/
type KsmService struct {
	*BaseService
	//client   *util.RpcClient
	url      string
	nonceCtl map[string]int64
}

/*
初始化币种服务
	注意：
		方法接受者： BaseService
		方法命名： 币种大写 + Service
*/
func (bs *BaseService) KSMService() *KsmService {
	ks := new(KsmService)
	ks.BaseService = bs
	ks.url = conf.Config.KsmCfg.NodeUrl
	ks.nonceCtl = make(map[string]int64)
	return ks
}

/*
接口创建地址服务
	无需改动
*/
func (ks *KsmService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	return nil, errors.New("do not support this coin create address service")
}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (ks *KsmService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	fmt.Println("start create Ksm address")
	_, err := ks.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, ks.createAddressInfo)
	return err
}
func (ks *KsmService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	reqBalanceParams := map[string]string{
		"addr": req.Address,
	}

	respBData, err := util.PostJson(ks.url+"/balance", reqBalanceParams)
	if err != nil {
		return nil, fmt.Errorf("get address %s balance error,err=%v", req.Address, err)
	}
	var r model.KsmRespNodeParams
	if err := json.Unmarshal(respBData, &r); err != nil {
		return nil, err
	}
	if r.Code != 200 || r.Message != "ok" {
		return nil, fmt.Errorf("get address %s balance error,Message=[%s]", req.Address, r.Message)
	}
	balance, err := decimal.NewFromString(r.Data)
	if err != nil {
		return nil, fmt.Errorf("ksm balance change to decimal error: %v", err)
	}
	return balance.String(), nil
}

func (ks *KsmService) ValidAddress(address string) error {
	return ss58.VerityAddress(address, ss58.KsmPrefix)
}

/*
签名服务
*/
func (ks *KsmService) SignService(req *model.ReqSignParams) (interface{}, error) {
	return nil, nil
}

/*
热钱包出账服务
*/
func (ks *KsmService) TransferService(req interface{}) (interface{}, error) {
	var tp model.KsmTransferParams
	if err := ks.BaseService.parseData(req, &tp); err != nil {
		return nil, err
	}

	if &tp == nil {
		return nil, errors.New("transfer params is null")
	}
	if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	}

	cli, err := ksm.NewClient(ks.url)
	if err != nil {
		log.Info("err:", err)
		return nil, err
	}

	meta, err := cli.Api.RPC.State.GetMetadataLatest()
	if err != nil {
		return nil, err
	}

	params := ksm.Txparam{
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

	rawTx, err := ksm.SignTx(&params, privateKey)
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
func (ks *KsmService) createAddressInfo() (util.AddrInfo, error) {
	return util.AddrInfo{}, errors.New("do not support this method")
}
