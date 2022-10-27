package rylink

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/zecserver/util"
	"github.com/sirupsen/logrus"
)

var ZecRpcClient *ZecClient

type ZecClient struct {
	ConnCfg *util.RpcConnConfig
}

//创建一个新实例
func NewZecClient(connCfg *util.RpcConnConfig) *ZecClient {
	return &ZecClient{connCfg}
}

//创建交易结构
func (o *ZecClient) Createtx(createtx []*RpcCreatetx, data map[string]float64) (*CreaterawtransactionResult, error) {
	var (
		out *CreaterawtransactionResult
	)
	c := util.NewRpcClient(o.ConnCfg)
	createStr, _ := json.Marshal(createtx)
	outStr, _ := json.Marshal(data)
	fmt.Println(string(createStr))
	fmt.Println(string(outStr))
	byteData, err := c.Call("createrawtransaction", createtx, data, 0, 0)
	if err != nil {
		logrus.Errorf("Error createrawtransaction: %v", err)
		return nil, err
	}
	out, err = DecodeCreaterawtransactionResult(byteData)
	if err != nil {
		logrus.Errorf("Error decode createrawtransaction: %v, rpcerror: %v", err, string(byteData))
		return nil, err
	}
	if out.Error != nil {
		logrus.Errorf("Error createrawtransaction result: %v", out)
		dd, _ := json.Marshal(out.Error)
		return nil, errors.New(string(dd))
	}
	return out, nil
}

////交易签名
func (o *ZecClient) SignTx(m Signtx) (*SignResult, error) {
	fmt.Println("执行签名")
	var (
		out *SignResult
	)
	c := util.NewRpcClient(o.ConnCfg)
	byteData, err := c.Call("signrawtransaction", m.Rawtx, m.Prevtxs, m.Privatekeys)
	if err != nil {
		logrus.Errorf("GetPushTx Error signrawtransaction: %v", err)
		return nil, err
	}
	out, err = DecodeSignResult(byteData)
	if err != nil {
		logrus.Errorf("GetPushTx Error decode signrawtransaction: %v,RpcResult:%v", err, string(byteData))
		return nil, err
	}
	if out.Error != nil {
		logrus.Errorf("GetPushTx Error signrawtransaction result: %v", out)
		dd, _ := json.Marshal(out.Error)
		return nil, errors.New(string(dd))
	}
	return out, nil
}

//广播交易
func (o *ZecClient) PushTransaction(hex string) (*SenndTxResult, error) {
	var (
		out *SenndTxResult
	)
	c := util.NewRpcClient(o.ConnCfg)
	byteData, err := c.Call("sendrawtransaction", hex)
	if err != nil {
		logrus.Errorf("Error signrawtransaction: %v", err)
		return nil, err
	}
	out, err = DecodeSenndTxResult(byteData)
	if err != nil {
		logrus.Errorf("Error decode sendrawtransaction: %v,RpcResult:%+v", err, string(byteData))
		return nil, err
	}
	if out.Error != "" {
		logrus.Errorf("Error sendrawtransaction result: %v", out)
		return nil, errors.New(out.Error)
	}
	return out, nil
}

//解码交易
func (o *ZecClient) Decode(hex string) ([]byte, error) {
	c := util.NewRpcClient(o.ConnCfg)
	byteData, err := c.Call("decoderawtransaction", hex)
	if err != nil {
		logrus.Errorf("Error decoderawtransaction: %v", err)
		return nil, err
	}
	return byteData, nil
}

//创建新地址
//返回地址,私钥 公钥，脚本公钥
//func (o *OmniClient) GetNewAddress(tagName string) (address, prvkey, pubkey, scriptPubKey string, err error) {
func (o *ZecClient) GetNewAddress() (*AddressOutPut, error) {
	var (
		addressResult *GetNewAddressResult
		prvkeyResult  *DumpprivkeyResult
	)
	c := util.NewRpcClient(o.ConnCfg)
	byteData, err := c.Call("getnewaddress", "")
	//byteData, err := c.Call("getnewaddress", tagName)
	if err != nil {
		logrus.Errorf("Error getnewaddress: %v", err)
		return nil, err
	}
	addressResult, err = DecodeGetNewAddressResult(byteData)
	if err != nil {
		logrus.Errorf("Error decode getnewaddress: %v", err)
		return nil, err
	}
	byteData, err = c.Call("dumpprivkey", addressResult.Result)
	if err != nil {
		logrus.Errorf("Error dumpprivkey: %v", err)
		return nil, err
	}
	prvkeyResult, err = DecodeDumpprivkeyResult(byteData)
	if err != nil {
		logrus.Errorf("Error decode dumpprivkey: %v", err)
		return nil, err
	}
	newAddress := &AddressOutPut{
		Address:    addressResult.Result,
		PrivateKey: prvkeyResult.Result,
	}
	return newAddress, nil
}

//导入私钥,去除从头扫描
func (o *ZecClient) RpcImportprivkey(prvkey string) (*ImportprivkeyResult, error) {
	var (
		prvkeyResult *ImportprivkeyResult
	)
	c := util.NewRpcClient(o.ConnCfg)
	byteData, err := c.Call("importprivkey", prvkey, "", false)
	if err != nil {
		logrus.Errorf("Error importprivkey: %v", err)
		return nil, err
	}
	prvkeyResult, err = DecodeImportprivkeyResult(byteData)
	if err != nil {
		logrus.Errorf("Error decode importprivkey: %v", err)
		return nil, err
	}
	return prvkeyResult, nil

}
