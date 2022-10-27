package v1

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/Dipper-Labs/Dipper-Protocol/app/v0/bank"
	"github.com/Dipper-Labs/go-sdk/client/rpc"
	"github.com/Dipper-Labs/go-sdk/config"
	"github.com/Dipper-Labs/go-sdk/constants"
	"github.com/tendermint/tendermint/libs/bech32"
	"strconv"
	"strings"

	"fmt"
	"github.com/Dipper-Labs/Dipper-Protocol/app/v0/auth"

	sdk "github.com/Dipper-Labs/Dipper-Protocol/types"

	"github.com/Dipper-Labs/go-sdk/types"
	"github.com/btcsuite/btcd/btcec"
	"wallet-sign/conf"
	"wallet-sign/model"
	"wallet-sign/util"

	clitypes "github.com/Dipper-Labs/go-sdk/client/types"
	"github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"io/ioutil"
	"net/http"
)

/*
service模板
*/

/*
币种服务结构体
*/
type DipService struct {
	*BaseService
	chainId   string
	apiUrl    string
	rpcClient rpc.RpcClient
}

/*
初始化币种服务
	注意：
		方法接受者： BaseService
		方法命名： 币种大写 + Service
*/
func (bs *BaseService) DIPService() *DipService {
	tp := new(DipService)
	tp.BaseService = bs
	tp.apiUrl = conf.Config.DipCfg.ApiUrl
	var err error
	tp.chainId, err = initDipChainId()
	if err != nil {
		logrus.Errorf("init dip node info error: %v", err)
		//os.Exit(1)
	}
	tp.rpcClient = rpc.NewClient(conf.Config.DipCfg.NodeUrl)

	//初始化连接
	return tp
}

func initDipChainId() (string, error) {
	url := conf.Config.DipCfg.ApiUrl + "/node_info"
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("get dip node info error: %v", err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read resp dip node info error: %v", err)
	}
	var dni model.DipNodeInfo
	err = json.Unmarshal(data, &dni)
	if err != nil {
		return "", fmt.Errorf("json unmarshal dip node info error: %v", err)
	}
	return dni.NodeInfo.Network, nil
}

/*
接口创建地址服务
	无需改动
*/
func (ds *DipService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return ds.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, ds.createAddressInfo)
	}
	return ds.BaseService.createAddress(req, ds.createAddressInfo)
}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (ds *DipService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	_, err := ds.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, ds.createAddressInfo)
	return err
}

/*
签名服务
*/
func (ds *DipService) SignService(req *model.ReqSignParams) (interface{}, error) {
	return nil, nil
}

func (ds *DipService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	data, err := ds.httpGet(fmt.Sprintf("/auth/accounts/%s", req.Address))
	if err != nil {
		return nil, fmt.Errorf("获取from地址金额错误： %v", err)
	}
	var accountBody model.DipAccountBody
	err = json.Unmarshal(data, &accountBody)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal accounInfo error: %v", err)
	}
	currentDipAmount := ds.getCoin(accountBody.Result.Value.Coins, constants.TxDefaultDenom)
	return currentDipAmount.Amount.String(), nil
}
func (ds *DipService) ValidAddress(address string) error {
	if len(strings.TrimSpace(address)) == 0 {
		return fmt.Errorf("address is null :%s", address)
	}
	bz, err := ds.getFromBech32(address, "dip")
	if err != nil {
		return fmt.Errorf("dip address error: %v", err)
	}
	n := len(bz)

	if n == 10 || n == 20 {
		return nil
	}
	return fmt.Errorf("invalid address length %d", n)
}

// GetFromBech32 decodes a bytestring from a Bech32 encoded string.
func (ds *DipService) getFromBech32(bech32str, prefix string) ([]byte, error) {
	if len(bech32str) == 0 {
		return nil, errors.New("decoding Bech32 address failed: must provide an address")
	}

	hrp, bz, err := bech32.DecodeAndConvert(bech32str)
	if err != nil {
		return nil, err
	}

	if hrp != prefix {
		return nil, fmt.Errorf("invalid Bech32 prefix; expected %s, got %s", prefix, hrp)
	}

	return bz, nil
}

/*
热钱包出账服务
*/
func (ds *DipService) TransferService(req interface{}) (interface{}, error) {

	var tp model.DipTransferParams
	if err := ds.BaseService.parseData(req, &tp); err != nil {

		return nil, err
	}
	if &tp == nil {
		return nil, errors.New("transfer params is null")
	}
	if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	}
	if tp.Denom == "" {
		return nil, fmt.Errorf("denom is null,无法区分是主链还是代币转账")
	}
	fromAddr, err := sdk.AccAddressFromBech32(tp.FromAddress)
	if err != nil {
		return "", fmt.Errorf("from address [%s] is error: %v", tp.FromAddress, err)
	}
	toAddr, err := sdk.AccAddressFromBech32(tp.ToAddress)
	if err != nil {
		return "", fmt.Errorf("to address [%s] is error: %v", tp.ToAddress, err)
	}
	coins, err := ds.buildCoins(tp.Denom, tp.Amount)
	if err != nil {
		return nil, fmt.Errorf("build coins error: %v", err)
	}
	if tp.Fee == 0 {
		tp.Fee = int64(1200000000000)
	}
	fee := sdk.Coins{
		{
			Denom:  constants.TxDefaultDenom,
			Amount: sdk.NewInt(tp.Fee),
		},
	}
	var (
		data []byte
	)
	if ds.chainId == "" {
		ds.chainId, err = initDipChainId()
	}
	data, err = ds.httpGet(fmt.Sprintf("/auth/accounts/%s", tp.FromAddress))
	if err != nil {
		return nil, fmt.Errorf("获取from地址金额错误： %v", err)
	}
	var accountBody model.DipAccountBody
	err = json.Unmarshal(data, &accountBody)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal accounInfo error: %v", err)
	}
	currentDipAmount := ds.getCoin(accountBody.Result.Value.Coins, constants.TxDefaultDenom)
	if tp.Denom == constants.TxDefaultDenom {
		toSpend := coins[0].Amount.Add(fee.AmountOf(constants.TxDefaultDenom))
		if currentDipAmount.Amount.LT(toSpend) {
			return nil, fmt.Errorf("from address amount is not enough: chainAmount:%s,toAmount: %s,Fee=%s",
				currentDipAmount.Amount.String(), coins[0].Amount.String(), fee.String())
		}
	} else {
		if currentDipAmount.Amount.LT(fee.AmountOf(constants.TxDefaultDenom)) {
			return nil, fmt.Errorf("transfer token:from address [dip] amount is not enough,chainAmount:%s,Fee=%s ",
				currentDipAmount.Amount.String(), fee.String())
		}
		ca := ds.getCoin(accountBody.Result.Value.Coins, tp.Denom)
		if ca.Amount.LT(coins[0].Amount) {
			return nil, fmt.Errorf("denom=[%s] amount is not enough,chainAmount: %s,toAmount:%s",
				tp.Denom, ca.Amount.String(), coins[0].Amount.String())
		}
	}

	msg := bank.MsgSend{
		FromAddress: fromAddr,
		ToAddress:   toAddr,
		Amount:      coins,
	}

	accountNumber, err := strconv.Atoi(accountBody.Result.Value.AccountNumber)
	if err != nil {
		return nil, fmt.Errorf("get account Number error: %v", err)
	}

	sequence, err := strconv.Atoi(accountBody.Result.Value.Sequence)
	if err != nil {
		return nil, fmt.Errorf("get sequence error: %v", err)
	}
	if tp.Gas == 0 {
		tp.Gas = config.TxDefaultGas
	}
	stdSignMsg := types.StdSignMsg{
		ChainID:       ds.chainId,
		AccountNumber: uint64(accountNumber),
		Sequence:      uint64(sequence),
		Fee:           auth.NewStdFee(tp.Gas, fee),
		Msgs:          []sdk.Msg{msg},
		Memo:          tp.Memo,
	}
	for _, msg := range stdSignMsg.Msgs {
		if err := msg.ValidateBasic(); err != nil {
			return nil, fmt.Errorf("valid base sign msg error: %v", err)
		}
	}
	//获取私钥
	//获取私钥
	privateKey, err := ds.addressOrPublicKeyToPrivate(tp.FromAddress)
	if err != nil {
		return nil, err
	}
	dkm, err := NewDipKeyManager(privateKey)
	if err != nil {
		return nil, fmt.Errorf("init key manager error: %v", err)
	}
	if dkm.GetAddr().String() != tp.FromAddress {
		return nil, fmt.Errorf("完蛋了，生成的私钥和地址不匹配：dkmAddress=[%s],fromAddress=[%s]",
			dkm.GetAddr().String(), tp.FromAddress)
	}
	//签名
	txBytes, err := dkm.Sign(stdSignMsg)
	if err != nil {
		return nil, fmt.Errorf("sign error: %v", err)
	}
	//广播交易
	txBroadcastType := constants.TxBroadcastTypeCommit //只管提交，不管是否入链
	result, err := ds.rpcClient.BroadcastTx(txBroadcastType, txBytes)
	if err != nil {
		return nil, fmt.Errorf("broad tx error: %v", err)
	}

	return result.CommitResult.Hash.String(), nil
}

/*
创建地址实体方法
*/
func (ds *DipService) createAddressInfo() (util.AddrInfo, error) {
	priv := secp256k1.GenPrivKey()
	p := [32]byte(priv)
	wif := hex.EncodeToString(p[:])
	addr := sdk.AccAddress(priv.PubKey().Address())
	address := addr.String()
	var addrInfo util.AddrInfo
	addrInfo.PrivKey = wif
	addrInfo.Address = address
	return addrInfo, nil
}

type DipKeyManager struct {
	privKey crypto.PrivKey
	addr    sdk.AccAddress
}

func NewDipKeyManager(privateKey string) (*DipKeyManager, error) {
	privateKey = strings.TrimPrefix(privateKey, "0x")
	privBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, fmt.Errorf("decode private hex error: %v", err)
	}
	var p [32]byte
	copy(p[:], privBytes[:])
	priv := secp256k1.PrivKeySecp256k1(p)
	dkm := new(DipKeyManager)
	dkm.privKey = priv
	dkm.addr = sdk.AccAddress(priv.PubKey().Address())

	return dkm, nil
}
func (d *DipKeyManager) Sign(msg types.StdSignMsg) ([]byte, error) {
	sig, err := d.makeSignature(msg)
	if err != nil {
		return nil, err
	}

	newTx := auth.NewStdTx(msg.Msgs, msg.Fee, []auth.StdSignature{sig}, msg.Memo)
	bz, err := types.Cdc.MarshalBinaryLengthPrefixed(newTx)
	if err != nil {
		return nil, err
	}

	return bz, nil
}

func (d *DipKeyManager) SignBytes(msg []byte) ([]byte, error) {
	return d.privKey.Sign(msg)
}

func (d *DipKeyManager) GetPrivKey() crypto.PrivKey {
	return d.privKey
}

func (d *DipKeyManager) GetAddr() sdk.AccAddress {
	return d.addr
}

func (d *DipKeyManager) GetUCPubKey() (UCPubKey []byte, err error) {
	pubkey, err := btcec.ParsePubKey(d.GetPrivKey().PubKey().Bytes()[5:], btcec.S256())
	if err != nil {
		return nil, err
	}

	return pubkey.SerializeUncompressed(), nil
}

func (d *DipKeyManager) makeSignature(msg types.StdSignMsg) (sig auth.StdSignature, err error) {
	sigBytes, err := d.privKey.Sign(msg.Bytes())
	if err != nil {
		return
	}
	return auth.StdSignature{
		PubKey:    d.privKey.PubKey(),
		Signature: sigBytes,
	}, nil
}

func (ds *DipService) buildCoins(denom, amount string) (sdk.Coins, error) {
	var coin clitypes.Coin
	var inputCoins []clitypes.Coin
	coin.Denom = denom
	coin.Amount = amount
	inputCoins = append(inputCoins, coin)

	var coins []sdk.Coin

	if len(inputCoins) == 0 {
		return coins, nil
	}

	for _, coin := range inputCoins {
		if amount, ok := sdk.NewIntFromString(coin.Amount); ok {
			coins = append(coins, sdk.Coin{
				Denom:  coin.Denom,
				Amount: amount,
			})
		} else {
			return coins, fmt.Errorf("can't parse str to Int, coin is %+v", inputCoins)
		}
	}

	return coins, nil
}

func (ds *DipService) httpGet(path string) ([]byte, error) {
	url := ds.apiUrl + path
	req := util.HttpGet(url)
	return req.Bytes()
}

func (ds *DipService) getCoin(inputCoins []clitypes.Coin, denom string) sdk.Coin {
	for _, coin := range inputCoins {
		if coin.Denom == denom {
			if amount, ok := sdk.NewIntFromString(coin.Amount); ok {
				return sdk.Coin{
					Denom:  coin.Denom,
					Amount: amount,
				}
			}

			break
		}
	}

	return sdk.Coin{}
}
