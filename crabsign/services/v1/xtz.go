package v1

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/goat-systems/go-tezos/v4/forge"
	"github.com/goat-systems/go-tezos/v4/keys"
	"github.com/goat-systems/go-tezos/v4/rpc"
	"github.com/prometheus/common/log"
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
	"strings"
	"wallet-sign/conf"
	"wallet-sign/model"
	"wallet-sign/util"
)

/*
service模板
*/

/*
币种服务结构体
*/
type XtzService struct {
	*BaseService
	client *rpc.Client
}

/*
初始化币种服务

	注意：
		方法接受者： BaseService
		方法命名： 币种大写 + Service
*/
func (bs *BaseService) XTZService() *XtzService {
	tp := new(XtzService)
	tp.BaseService = bs
	tp.client, _ = rpc.New(conf.Config.XtzCfg.NodeUrl)
	//初始化连接
	return tp
}

/*
接口创建地址服务
	无需改动
*/
func (tp *XtzService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return tp.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, tp.createAddressInfo)
	}
	return tp.BaseService.createAddress(req, tp.createAddressInfo)
}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (tp *XtzService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	_, err := tp.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, tp.createAddressInfo)
	return err
}

/*
签名服务
*/
func (tp *XtzService) SignService(req *model.ReqSignParams) (interface{}, error) {
	return nil, nil
}

/*
热钱包出账服务
*/
func (tp *XtzService) TransferService(req interface{}) (interface{}, error) {
	if tp.client == nil {
		client, err := rpc.New(conf.Config.XtzCfg.NodeUrl)
		if err != nil || client == nil {
			return nil, fmt.Errorf("init xtz rpc client error: %v", err)
		}
		tp.client = client
	}
	var parmas model.XtzTransferParams
	if err := tp.BaseService.parseData(req, &parmas); err != nil {
		return nil, err
	}
	if &parmas == nil {
		return nil, errors.New("transfer params is null")
	}
	if parmas.FromAddress == "" || parmas.ToAddress == "" || parmas.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", parmas.FromAddress, parmas.ToAddress, parmas.Amount)
	}
	if parmas.FromAddress == parmas.ToAddress {
		return nil, fmt.Errorf("from地址不能和to地址相同；from=[%s],to=[%s]", parmas.FromAddress, parmas.ToAddress)
	}
	// 判断from地址是否有足够的钱去出账

	balance, err := tp.GetBalance(&model.ReqGetBalanceParams{Address: parmas.FromAddress})
	if err != nil || balance == nil {
		return nil, fmt.Errorf("get from amount error: %v", err)
	}
	ba, _ := decimal.NewFromString(balance.(string))
	amt, _ := decimal.NewFromString(parmas.Amount)
	if ba.LessThan(amt) {
		return nil, fmt.Errorf("from地址[%s]没有足够的钱去出账，from chain amount=[%s],transfer amount=[%s]", ba.String(), amt.String())
	}
	//构建交易
	// 1。 先根据地址找到对应的私钥
	privateKey, err := tp.BaseService.addressOrPublicKeyToPrivate(parmas.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("get private key error,Err=%v", err)
	}
	privateKey = strings.TrimPrefix(privateKey, "0x")

	//2. 根据私钥创建key
	key, err := keys.FromHex(privateKey, keys.Secp256k1) //我们生成地址的时候都用Secp256k1曲线
	if err != nil {
		return nil, fmt.Errorf("create xtz keys error: %v", err)
	}
	//3. 判断key的地址是否和from地址相等
	if key.PubKey.GetAddress() != parmas.FromAddress {
		return nil, fmt.Errorf("keys address is not equal from address,please check"+
			" this address:keysAddress=[%s],fromAddress=[%s]", key.PubKey.GetAddress(), parmas.FromAddress)
	}
	//4. 获取from地址的nonce
	resp, counter, err := tp.client.ContractCounter(rpc.ContractCounterInput{
		BlockID:    &rpc.BlockIDHead{},
		ContractID: key.PubKey.GetAddress(),
	})
	if err != nil {
		return nil, fmt.Errorf("get %s nonce error: %v", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("get %s nonce error: %v,status=[%s]",
			parmas.FromAddress, err, resp.Status())
	}

	counter++
	//todo 5. 判断counter是否已经被使用过
	log.Infof("当前counter： %d", counter)
	// 6. 获取最新区块的hash
	resp, head, err := tp.client.Block(&rpc.BlockIDHead{})
	if err != nil {
		return nil, fmt.Errorf("get latest block hash error: %v", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("get latest block hash error: %v,status=[%s]",
			err, resp.Status())
	}

	if parmas.Fee == "" || parmas.Fee == "0" {
		parmas.Fee = "2941"

	}

	if parmas.GasLimit == "" || parmas.GasLimit == "0" {
		parmas.GasLimit = "30000"
	}
	// 设置最大的手续费不能超过多少
	fd, _ := decimal.NewFromString(parmas.Fee)
	if fd.GreaterThan(decimal.NewFromInt(100000)) {
		parmas.Fee = "100000"
	}
	log.Infof("Fee=[%s],GasLimit=[%s]", parmas.Fee, parmas.GasLimit)
	//7. 创建交易
	big.NewInt(0).SetString("10000000000000000000000000000", 10)
	transaction := rpc.Transaction{
		Kind:         rpc.TRANSACTION,
		Source:       parmas.FromAddress,
		Fee:          parmas.Fee,
		GasLimit:     parmas.GasLimit,
		StorageLimit: "300",
		Counter:      strconv.Itoa(counter),
		Amount:       parmas.Amount,
		Destination:  parmas.ToAddress,
	}
	//fmt.Println(head.Hash)
	fmt.Println(head.ChainID)
	op, err := forge.Encode(head.ChainID, transaction.ToContent())
	if err != nil {
		return nil, fmt.Errorf("encode transaction error: %v", err)
	}
	// 8. 签名交易
	signature, err := key.SignHex(op)
	if err != nil {
		return nil, fmt.Errorf("sign error: %v", err)
	}
	// 9. 广播交易
	resp, ophash, err := tp.client.InjectionOperation(rpc.InjectionOperationInput{
		Operation: signature.AppendToHex(op),
		ChainID:   head.ChainID,
	})
	if err != nil {
		return nil, fmt.Errorf("broadcast transaction error: %v", err)
	}
	// 避免广播虽然返回错误了，但是返回了txid，而这笔却又出去了
	if resp.IsError() && ophash == "" {
		return nil, fmt.Errorf("broadcast transaction error: %v,status=[%s]",
			err, resp.Status())
	}

	return ophash, nil
}

func (tp *XtzService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	if tp.client == nil {
		client, err := rpc.New(conf.Config.XtzCfg.NodeUrl)
		if err != nil || client == nil {
			return nil, fmt.Errorf("init xtz rpc client error: %v", err)
		}
		tp.client = client
	}
	err := tp.ValidAddress(req.Address)
	if err != nil {
		return nil, err
	}
	resp, balance, err := tp.client.ContractBalance(rpc.ContractBalanceInput{
		ContractID: req.Address,
		BlockID:    &rpc.BlockIDHead{},
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.IsError() {
		return nil, fmt.Errorf("get balance error: %v,status=[%s]", resp.Error(), resp.Status())
	}
	return balance, nil
}
func (tp *XtzService) ValidAddress(address string) error {
	data, err := util.XTZDecode(address)
	if err != nil {
		return fmt.Errorf("base58 decode [%s] error: %v", address, err)
	}
	if len(data) != 23 { //prefix:3 + pubhash:20
		return fmt.Errorf("base58 decode address len is not equal 23,len=%d", len(data))
	}
	preifx := data[:3]
	err = tp.validAddressPrefix(preifx)
	if err != nil {
		return err
	}
	return nil
}

/*
创建地址实体方法
*/
func (tp *XtzService) createAddressInfo() (util.AddrInfo, error) {
	var (
		key *keys.Key
		err error
	)
	for true {
		key, err = keys.Generate(keys.Secp256k1)
		if err != nil {
			continue
		}
		if len(key.GetBytes()) != 32 {
			continue
		}
		break
	}
	if key == nil {
		return util.AddrInfo{}, errors.New("key is nil")
	}
	secretKey := hex.EncodeToString(key.GetBytes())
	address := key.PubKey.GetAddress()
	var addrInfo util.AddrInfo
	addrInfo.PrivKey = secretKey
	addrInfo.Address = address
	return addrInfo, nil
}

func (tp *XtzService) validAddressPrefix(prefix []byte) error {
	if len(prefix) != 3 {
		return fmt.Errorf("prefix length is not equal 3,len=%d", len(prefix))
	}
	if bytes.Equal(prefix, []byte{6, 161, 159}) || //	ed25519
		bytes.Equal(prefix, []byte{6, 161, 164}) || //nistP256
		bytes.Equal(prefix, []byte{6, 161, 161}) { //secp256k1
		return nil
	}
	return fmt.Errorf("unknown curve prefix: %v", prefix)
}
