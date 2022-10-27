package v1

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	acc "github.com/JFJun/near-go/account"
	"github.com/JFJun/near-go/rpc"
	"github.com/JFJun/near-go/serialize"
	"github.com/JFJun/near-go/transaction"
	"github.com/shopspring/decimal"
	"regexp"
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
type NearService struct {
	*BaseService
	client *rpc.Client
	url    string
	//nonceCtl map[string]int64
}

/*
初始化币种服务
	注意：
		方法接受者： BaseService
		方法命名： 币种大写 + Service
*/
func (bs *BaseService) NEARService() *NearService {
	ks := new(NearService)
	ks.BaseService = bs
	ks.url = conf.Config.NearCfg.NodeUrl
	//ks.nonceCtl = make(map[string]int64)
	return ks
}

/*
接口创建地址服务
	无需改动
*/
func (ks *NearService) CreateAddressService(req *model.ReqCreateAddressParams) (*model.RespCreateAddressParams, error) {
	if conf.Config.IsStartThread {
		return ks.BaseService.multiThreadCreateAddress(req.Num, req.CoinName, req.MchId, req.OrderId, ks.createAddressInfo)
	}
	return ks.BaseService.createAddress(req, ks.createAddressInfo)
}

/*
离线创建地址服务，通过多线程创建
	无需改动
*/
func (ks *NearService) MultiThreadCreateAddrService(nums int, coinName, mchId, orderId string) error {
	fmt.Println("start create Near address")
	_, err := ks.BaseService.multiThreadCreateAddress(nums, coinName, mchId, orderId, ks.createAddressInfo)
	return err
}

/*
签名服务
*/
func (ks *NearService) SignService(req *model.ReqSignParams) (interface{}, error) {
	return nil, errors.New("unsopport")
}

func (ks *NearService) GetBalance(req *model.ReqGetBalanceParams) (interface{}, error) {
	amount, locked, err := ks.client.GetAccountBalance(req.Address)
	if err != nil {
		return nil, fmt.Errorf("get account balance error,Err=%v", err)
	}
	resp := make(map[string]string)
	resp["amount"] = amount
	resp["locked"] = locked
	return resp, nil
}

func (ks *NearService) ValidAddress(address string) error {
	if len(address) < 2 || len(address) > 64 {
		return fmt.Errorf("地址长度不正确，大于64或者小于2，length=%d", len(address))
	}
	neg := `^(([a-z\d]+[\-_])*[a-z\d]+\.)*([a-z\d]+[\-_])*[a-z\d]+$`
	reg, err := regexp.Compile(neg)
	if err != nil {
		return fmt.Errorf("正则验证地址错误，Err=%v", err)
	}
	if !reg.MatchString(address) {
		return errors.New("正则验证地址错误，地址不满足正则条件")
	}
	return nil
}

/*
热钱包出账服务
*/
func (ks *NearService) TransferService(req interface{}) (interface{}, error) {
	if ks.client == nil {
		//初始化客户端
		c, err := rpc.NewRpcClient(ks.url, "", "")
		if err != nil {
			return nil, err
		}
		ks.client = c
	}
	var tp model.NearTransferParams
	if err := ks.BaseService.parseData(req, &tp); err != nil {

		return nil, err
	}
	if &tp == nil {
		return nil, errors.New("transfer params is null")
	}
	if tp.FromAddress == "" || tp.ToAddress == "" || tp.Amount == "" {
		return nil, fmt.Errorf("params is null,from=[%s],to=[%s],amount=[%s]", tp.FromAddress, tp.ToAddress, tp.Amount)
	}

	// 1. 判断金额是否足够出账
	amount, locked, err := ks.client.GetAccountBalance(tp.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("get account balance error,Err=%v", err)
	}
	ad, _ := decimal.NewFromString(amount)
	ld, _ := decimal.NewFromString(locked)
	transferAmount, _ := decimal.NewFromString(tp.Amount)
	chainAmount := ad.Sub(ld)
	if transferAmount.GreaterThanOrEqual(chainAmount) {
		return nil, fmt.Errorf("amount is not engouth,chain amount=%s,transfer amount=%s",
			chainAmount.Shift(-24).String(), transferAmount.Shift(-24).String())
	}
	// 2. 获取最新区块hash
	blockHash, err := ks.client.GetLatestBlockHash()
	if err != nil {
		return nil, fmt.Errorf("get latest block hash error,err=%v", err)
	}
	// 3. 获取公钥 （因为我们的地址都是public key的16进制，所以我就直接从地址转公钥了，如果是其他的地址，
	//需要使用rpc接口去获取公钥）
	pub, err := hex.DecodeString(tp.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("hex decode from error,From=%s,Err=%v", tp.FromAddress, err)
	}
	pubKey := acc.PublicKeyToString(pub)
	// 4. 获取nonce
	nonce, err := ks.client.GetNonce(tp.FromAddress, pubKey, "")
	if err != nil {
		return nil, fmt.Errorf("get nonce error,Err=%v", err)
	}
	//5. 创建交易
	tx, err := transaction.CreateTransaction(
		tp.FromAddress,
		tp.ToAddress,
		pubKey,
		blockHash,
		nonce,
	)
	if err != nil {
		return nil, fmt.Errorf("create tx error,Err=%v", err)
	}
	// 6. 创建transfer action
	transferAction, err := serialize.CreateTransfer(tp.Amount)
	if err != nil {
		return nil, fmt.Errorf("create transfer action error,Err=%v", err)
	}
	//7. 设置action
	tx.SetAction(transferAction)
	//8. 序列化交易
	txData, err := tx.Serialize()
	if err != nil {
		return nil, fmt.Errorf("tx serialize error,Err=%v", err)
	}
	txHex := hex.EncodeToString(txData)
	//9. 根据dizhi获取私钥
	privateKey, err := ks.addressOrPublicKeyToPrivate(tp.FromAddress)
	if err != nil {
		return nil, fmt.Errorf("获取死要失败：Address=%s,Err=%v", tp.FromAddress, err)
	}
	//10. 签名交易
	sig, err := transaction.SignTransaction(txHex, privateKey)
	if err != nil {
		return nil, fmt.Errorf("签名失败： Err=%v", err)
	}
	//11. 创建签名后的交易
	stx, err := transaction.CreateSignatureTransaction(tx, sig)
	if err != nil {
		return nil, fmt.Errorf("create sign tx error,Err=%v", err)
	}
	//12. 序列化签名后的交易
	stxData, err := stx.Serialize()
	if err != nil {
		return nil, fmt.Errorf("stx serialize error,Err=%v", err)
	}
	//13. 转换为base64
	b64Data := base64.StdEncoding.EncodeToString(stxData)
	//14. 广播交易
	txid, err := ks.client.BroadcastTransaction(b64Data)
	if err != nil {
		return nil, fmt.Errorf("广播交易失败：，Err=%v", err)
	}
	return txid, nil
}

/*
创建地址实体方法
*/
func (ks *NearService) createAddressInfo() (util.AddrInfo, error) {
	var addrInfo util.AddrInfo
	priv, pub, err := acc.GenerateKey()
	if err != nil {
		return addrInfo, err
	}
	if len(priv) != 32 || len(pub) != 32 {
		return addrInfo, fmt.Errorf("priv or pub length is not equal 32,Priv=%d,Pub=%d", len(priv), len(pub))
	}
	secret := hex.EncodeToString(priv)
	address := acc.PublicKeyToAddress(pub)
	addrInfo.PrivKey = secret
	addrInfo.Address = address
	return addrInfo, nil
}
