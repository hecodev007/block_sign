package v1

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/group-coldwallet/wallet-sign/conf"
	"github.com/group-coldwallet/wallet-sign/model"
	"github.com/group-coldwallet/wallet-sign/util"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/sha3"
	"log"
	"math/big"
	"strings"
)

/*
service模板
*/

/*
币种服务结构体
*/
type CdsService struct {
	*BaseService
	client   *util.RpcClient
	nonceCtl map[string]int64
}

/*
初始化币种服务
	注意：
		方法接受者： BaseService
		方法命名： 币种大写 + Service
*/
func (bs *BaseService) CDSService() *CdsService {
	cs := new(CdsService)
	cs.BaseService = bs
	//初始化连接
	client := util.New(conf.Config.CdsCfg.NodeUrl, conf.Config.CdsCfg.User, conf.Config.CdsCfg.Password)
	cs.client = client
	cs.nonceCtl = make(map[string]int64)
	return cs
}

/*
接口创建地址服务
	无需改动
*/
func (cs *CdsService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return cs.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, cs.createAddressInfo)
	}
	return cs.BaseService.createAddress(req, cs.createAddressInfo)
}

func (cs *CdsService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	return nil, errors.New("unsupport it")
}
func (cs *CdsService) ValidAddress(address string) error {
	if !common.IsHexAddress(address) {
		return errors.New("valid cds address error")
	}
	return nil
}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (cs *CdsService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	fmt.Println("start create cds address")
	_, err := cs.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, cs.createAddressInfo)
	return err
}

/*
签名服务
*/
func (cs *CdsService) SignService(req *model.ReqSignParams) (interface{}, error) {
	return nil, nil
}

/*
热钱包出账服务
*/
func (cs *CdsService) TransferService(req interface{}) (interface{}, error) {
	var tp model.CdsTransferParams
	if err := cs.BaseService.parseData(req, &tp); err != nil {
		return nil, err
	}
	if &tp == nil {
		return nil, errors.New("transfer params is null")
	}
	if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	}
	if cs.client == nil {
		client := util.New(conf.Config.CdsCfg.NodeUrl, conf.Config.CdsCfg.User, conf.Config.CdsCfg.Password)
		cs.client = client
	}
	nonce := cs.getNonce("eth_getTransactionCount", tp.FromAddress)
	fmt.Println("nonce: ", nonce)
	if nonce < 0 {
		return nil, errors.New("get nonce error")
	}
	gasPrice, _ := cs.getBuildTxParams("eth_gasPrice", []interface{}{})
	if gasPrice < 0 {
		gasPrice = conf.Config.CdsCfg.GasPrice
	}
	var amount *big.Int
	toAmount, err := decimal.NewFromString(tp.Amount)
	if err != nil {
		return nil, fmt.Errorf("parse amount error,err=%v", err)
	}
	if tp.IsCollect > 0 {
		fee := new(big.Int)
		fee = big.NewInt(gasPrice * 60000)
		amount = new(big.Int).Sub(toAmount.BigInt(), fee)
	} else {
		amount = toAmount.BigInt()
	}

	blockCount, err1 := cs.getBuildTxParams("eth_blockNumber", []interface{}{})
	if err1 != nil {
		return nil, fmt.Errorf("get block number error,err=%v", err)
	}
	if blockCount < 0 {
		return nil, fmt.Errorf("get block number error,number=%d", blockCount)
	}
	var tx *types.Transaction
	// gas_limit 限制死为 60000
	toAddress := common.HexToAddress(tp.ToAddress)

	if strings.Compare(del0xToLower(toAddress.String()), del0xToLower(tp.ToAddress[:])) != 0 {
		return nil, fmt.Errorf("to address is not equal,address1=[%s],address2=[%s]", del0xToLower(toAddress.String()),
			del0xToLower(tp.ToAddress[:]))
	}
	tx = types.NewTransaction(uint64(nonce), toAddress, amount, uint64(60000), big.NewInt(gasPrice), nil)
	if tx == nil {
		return nil, errors.New("build tx error")
	}

	from := tp.FromAddress
	hexPrivateKey, err3 := cs.BaseService.addressOrPublicKeyToPrivate(from)
	if err3 != nil {
		return nil, fmt.Errorf("get private key error,Err=%v", err3)
	}
	privKey, err2 := crypto.HexToECDSA(hexPrivateKey)
	if privKey == nil || err2 != nil {
		return nil, fmt.Errorf("private key is null,err=%v", err)
	}

	//types.MakeSigner()

	// 签名
	var chainID *big.Int
	networkid := conf.Config.CdsCfg.NetWorkId
	if networkid == 1 {
		if config := params.MainnetChainConfig; config.IsEIP155(big.NewInt(int64(blockCount))) {
			chainID = config.ChainID
		}
	} else if networkid == 3 {
		if config := params.TestChainConfig; config.IsEIP155(big.NewInt(int64(blockCount))) {
			chainID = config.ChainID
		}
	} else if networkid == 4 {
		if config := params.RinkebyChainConfig; config.IsEIP155(big.NewInt(int64(blockCount))) {
			chainID = config.ChainID
		}
	}

	var signtx *types.Transaction
	var signerr error
	//chainID = nil
	if chainID != nil {
		signtx, signerr = types.SignTx(tx, types.NewEIP155Signer(chainID), privKey)
	} else {
		signtx, signerr = types.SignTx(tx, types.HomesteadSigner{}, privKey)
	}
	if signerr != nil {
		log.Println("could not sign transaction:", signerr)
		return nil, errors.New("could not sign transaction")
	}

	data, _ := rlp.EncodeToBytes(signtx)
	hextx := common.Bytes2Hex(data)
	if !strings.HasPrefix(hextx, "0x") {
		hextx = "0x" + hextx
	}

	res, err4 := cs.client.SendRequest("eth_sendRawTransaction", []interface{}{hextx})
	if err4 != nil {
		return nil, fmt.Errorf("send transaction error,Err=[%v]", err4)
	}
	if res == nil {
		return nil, errors.New("send transaction error,response null")
	}
	//维护nonce
	cs.nonceCtl[from] = nonce + 1
	return string(res), nil
}

func (cs *CdsService) getBuildTxParams(method string, params []interface{}) (int64, error) {
	res, err := cs.client.SendRequest(method, params)
	if err != nil {
		return -1, err
	}
	if res == nil {
		return -1, nil
	}
	ns := string(res)
	var nonceStr string
	if strings.HasPrefix(ns, "0x") {
		nonceStr = ns[2:]
	} else {
		nonceStr = ns
	}
	nonce := util.HexToDec(nonceStr)
	return nonce, nil
}

func (cs *CdsService) getNonce(method, address string) int64 {
	if cs.nonceCtl == nil {
		cs.nonceCtl = make(map[string]int64)
	}
	if cs.nonceCtl[address] > 0 {
		return cs.nonceCtl[address]
	}
	params := []interface{}{address, "latest"}
	nonce, _ := cs.getBuildTxParams(method, params)
	cs.nonceCtl[address] = nonce
	return nonce
}

/*
创建地址实体方法
*/
func (cs *CdsService) createAddressInfo() (util.AddrInfo, error) {
	privkey, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	var (
		addrInfo util.AddrInfo
		address  string
	)
	priv := privkey.D.Bytes()
	//避免priv的len不是32
	if len(priv) != 32 {
		for true {
			newPrivKey, err := crypto.GenerateKey()
			if err != nil {
				//if have some error ,cut this exe
				return addrInfo, err
			}
			priv = newPrivKey.D.Bytes()
			if len(priv) == 32 {
				address = strings.ToLower(crypto.PubkeyToAddress(privkey.PublicKey).Hex())
				break
			}
		}
	} else {
		address = strings.ToLower(crypto.PubkeyToAddress(privkey.PublicKey).Hex())
	}
	wif := hex.EncodeToString(priv)
	addrInfo.PrivKey = wif
	addrInfo.Address = address
	return addrInfo, nil
}

// Keccak256 calculates and returns the Keccak256 hash of the input data.
func keccak256(data ...[]byte) []byte {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

//创建地址

func del0xToLower(address string) string {
	if strings.HasPrefix(address, "0x") {
		return strings.ToLower(strings.TrimPrefix(address, "0x"))
	}
	return strings.ToLower(address)
}
