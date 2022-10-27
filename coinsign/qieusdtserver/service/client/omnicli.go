package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwalle/coinsign/qieusdtserver/models"
	"github.com/group-coldwalle/coinsign/qieusdtserver/util"
	log "github.com/sirupsen/logrus"
)

type OmniClient struct {
	ConnCfg *util.RpcConnConfig
}

//创建一个新实例
func NewOmniClient(connCfg *util.RpcConnConfig) *OmniClient {
	return &OmniClient{connCfg}
}

//获取用户余额
func (o *OmniClient) GetBalance(m models.BalanceInput) (*models.BalanceOutput, error) {
	var (
		out *models.BalanceOutput
	)

	c := util.NewRpcClient(o.ConnCfg)
	//保持命令参数顺序
	byteData, err := c.Call("omni_getbalance", m.Address, m.Propertyid)
	if err != nil {
		log.Errorf("Error omni_getbalance: %v", err)
		return nil, err
	}
	out, err = util.DecodeBalanceOutput(byteData)
	if err != nil {
		log.Errorf("Error decode omni_getbalance: %v", err)
		return nil, err
	}
	if out.Error != "" {
		log.Errorf("Error omni_getbalance result: %v", out)
		return nil, errors.New(out.Error)
	}

	return out, nil
}

//构建omni代币数据
//amount float64字符串
func (o *OmniClient) GetSimpleSend(propertyid int, amount string) (*models.OmniSimpleSendResult, error) {
	var (
		out *models.OmniSimpleSendResult
	)
	c := util.NewRpcClient(o.ConnCfg)
	byteData, err := c.Call("omni_createpayload_simplesend", propertyid, amount)
	if err != nil {
		log.Errorf("Error omni_createpayload_simplesend: %v", err)
		return nil, err
	}
	out, err = util.DecodeOmniSimpleSendResult(byteData)
	if err != nil {
		log.Errorf("Error decode omni_createpayload_simplesend: %v", err)
		return nil, err
	}
	if out.Error != nil {
		dd, _ := json.Marshal(out.Error)
		log.Errorf("Error omni_createpayload_simplesend result: %v", out)
		return nil, errors.New(string(dd))
	}
	return out, nil
}

//创建交易事务
func (o *OmniClient) GetCreateTx(in models.OmniCreateTx) (*models.OmniCreaterawtransactionResult, error) {
	var (
		out *models.OmniCreaterawtransactionResult
	)
	c := util.NewRpcClient(o.ConnCfg)
	byteData, err := c.Call("createrawtransaction", in.TxIns, in.Out)
	if err != nil {
		log.Errorf("Error createrawtransaction: %v", err)
		return nil, err
	}
	out, err = util.DecodeOmniCreaterawtransactionResult(byteData)
	if err != nil {
		log.Errorf("Error decode createrawtransaction: %v, rpcerror: %v", err, string(byteData))
		return nil, err
	}
	if out.Error != nil {
		dd, _ := json.Marshal(out.Error)
		log.Errorf("Error createrawtransaction result: %v", out)
		return nil, errors.New(string(dd))
	}
	return out, nil

}

//组合交易事务 CreateTxRaw+omni代币数据
func (o *OmniClient) GetOpreturnTx(simpleSendRaw, createTxRaw string) (*models.OmniOpreturnResult, error) {
	var (
		out *models.OmniOpreturnResult
	)
	c := util.NewRpcClient(o.ConnCfg)
	byteData, err := c.Call("omni_createrawtx_opreturn", createTxRaw, simpleSendRaw)
	if err != nil {
		log.Errorf("Error omni_createrawtx_opreturn: %v", err)
		return nil, err
	}
	out, err = util.DecodeOmniOpreturnResult(byteData)
	if err != nil {
		log.Errorf("Error decode omni_createrawtx_opreturn: %v", err)
		return nil, err
	}
	if out.Error != nil {
		dd, _ := json.Marshal(out.Error)
		log.Errorf("Error omni_createrawtx_opreturn result: %v", out)
		return nil, errors.New(string(dd))
	}
	return out, nil
}

//添加接收人地址
func (o *OmniClient) GetReferenceTx(rawtx, toAddress string, btcamount string) (*models.OmniReferenceResult, error) {
	var (
		out      *models.OmniReferenceResult
		byteData []byte
		err      error
	)
	c := util.NewRpcClient(o.ConnCfg)
	if btcamount == "" {
		byteData, err = c.Call("omni_createrawtx_reference", rawtx, toAddress)

	} else {
		byteData, err = c.Call("omni_createrawtx_reference", rawtx, toAddress, btcamount)
	}
	log.Infof("omni_createrawtx_reference 返回内容：：%s", string(byteData))
	if err != nil {
		log.Errorf("Error omni_createrawtx_reference: %v", err)
		return nil, err
	}
	out, err = util.DecodeOmniReferenceResult(byteData)
	if err != nil {
		log.Errorf("Error decode omni_createrawtx_reference: %v", err)
		return nil, err
	}
	if out.Error != nil {
		dd, _ := json.Marshal(out.Error)
		log.Errorf("Error omni_createrawtx_reference result: %v", out)
		return nil, errors.New(string(dd))
	}
	return out, nil
}

//设置找零地址和手续费
func (o *OmniClient) GetChangeTx(m models.OmniChangeTx) (*models.OmniChangeResult, error) {
	var (
		out *models.OmniChangeResult
	)
	c := util.NewRpcClient(o.ConnCfg)
	byteData, err := c.Call("omni_createrawtx_change", m.Rawtx, m.Prevtxs, m.Destination, m.Fee)
	if err != nil {
		log.Errorf("Error omni_createrawtx_change: %v", err)
		return nil, err
	}
	out, err = util.DecodeOmniChangeResult(byteData)
	if err != nil {
		log.Errorf("Error decode omni_createrawtx_change: %v", err)
		return nil, err
	}
	if out.Error != nil {
		dd, _ := json.Marshal(out.Error)
		log.Errorf("Error omni_createrawtx_change result: %v", out)
		return nil, errors.New(string(dd))
	}
	return out, nil
}

//模板签名
func (o *OmniClient) GetSignTx(m models.OmniSigntx) (*models.OmniSignResult, error) {
	var (
		out *models.OmniSignResult
	)
	c := util.NewRpcClient(o.ConnCfg)
	byteData, err := c.Call("signrawtransaction", m.Rawtx, m.Prevtxs)
	if err != nil {
		log.Errorf("GetSignTx Error signrawtransaction: %v", err)
		return nil, err
	}
	out, err = util.DecodeOmniSignResult(byteData)
	if err != nil {
		log.Errorf("GetSignTx Error decode signrawtransaction: %v,RpcResult:%v", err, string(byteData))
		return nil, err
	}
	if out.Error != nil {
		dd, _ := json.Marshal(out.Error)
		log.Errorf("GetSignTx Error signrawtransaction result: %v", out)
		return nil, errors.New(string(dd))
	}
	return out, nil
}

//模板签名
func (o *OmniClient) GetSignTx2(m models.OmniSigntx) (*models.OmniSignResult, error) {
	var (
		out *models.OmniSignResult
	)
	c := util.NewRpcClient(o.ConnCfg)
	byteData, err := c.Call("signrawtransaction", m.Rawtx, m.Prevtxs, m.Prvkey)
	if err != nil {
		log.Errorf("GetSignTx Error signrawtransaction: %v", err)
		return nil, err
	}
	out, err = util.DecodeOmniSignResult(byteData)
	if err != nil {
		log.Errorf("GetSignTx Error decode signrawtransaction: %v,RpcResult:%v", err, string(byteData))
		return nil, err
	}
	if out.Error != nil {
		dd, _ := json.Marshal(out.Error)
		log.Errorf("GetSignTx Error signrawtransaction result: %v", out)
		return nil, errors.New(string(dd))
	}
	return out, nil
}

//广播交易
func (o *OmniClient) PushTransaction(hex string) (*models.OmniSenndTxResult, error) {
	var (
		out *models.OmniSenndTxResult
	)
	c := util.NewRpcClient(o.ConnCfg)
	fmt.Println(hex)
	fmt.Println(hex)
	fmt.Println(hex)
	fmt.Println(hex)
	byteData, err := c.Call("sendrawtransaction", hex)
	if err != nil {
		log.Errorf("Error signrawtransaction: %v", err)
		return nil, err
	}
	out, err = util.DecodeOmniSenndTxResult(byteData)
	if err != nil {
		log.Errorf("Error decode signrawtransaction: %v,RpcResult:%+v", err, string(byteData))
		return nil, err
	}
	if out.Error != nil {
		dd, _ := json.Marshal(out.Error)
		log.Errorf("Error signrawtransaction result: %v", out)
		return nil, errors.New(string(dd))
	}
	return out, nil
}

//创建新地址
//返回地址,私钥 公钥，脚本公钥
//func (o *OmniClient) GetNewAddress(tagName string) (address, prvkey, pubkey, scriptPubKey string, err error) {
func (o *OmniClient) GetNewAddress() (*models.AddressOutPut, error) {
	var (
		addressResult *models.OmniGetNewAddressResult
		prvkeyResult  *models.OmniDumpprivkeyResult
		result        *models.OmniValidateaddressResult
	)

	c := util.NewRpcClient(o.ConnCfg)
	byteData, err := c.Call("getnewaddress", "")
	//byteData, err := c.Call("getnewaddress", tagName)
	if err != nil {
		log.Errorf("Error getnewaddress: %v", err)
		return nil, err
	}
	addressResult, err = util.DecodeOmniGetNewAddressResult(byteData)
	if err != nil {
		log.Errorf("Error decode getnewaddress: %v", err)
		return nil, err
	}
	byteData, err = c.Call("dumpprivkey", addressResult.Result)
	if err != nil {
		log.Errorf("Error dumpprivkey: %v", err)
		return nil, err
	}
	prvkeyResult, err = util.DecodeOmniDumpprivkeyResult(byteData)
	if err != nil {
		log.Errorf("Error decode dumpprivkey: %v", err)
		return nil, err
	}
	byteData, err = c.Call("validateaddress", addressResult.Result)
	result, err = util.DecodeOmniValidateaddressResult(byteData)
	if err != nil {
		log.Errorf("Error decode validateaddress: %v", err)
		return nil, err
	}
	newAddress := &models.AddressOutPut{
		//TageName:   tagName,
		Address:    addressResult.Result,
		PrivateKey: prvkeyResult.Result,
		PublicKey:  result.Result.Pubkey,
		ScriptKey:  result.Result.ScriptPubKey,
	}
	return newAddress, nil
}

//导出私钥
func (o *OmniClient) Dumpprivkey(addr string) (*models.OmniDumpprivkeyResult, error) {
	var (
		prvkeyResult *models.OmniDumpprivkeyResult
	)
	c := util.NewRpcClient(o.ConnCfg)
	byteData, err := c.Call("dumpprivkey", addr)
	if err != nil {
		log.Errorf("Error dumpprivkey: %v", err)
		return nil, err
	}
	prvkeyResult, err = util.DecodeOmniDumpprivkeyResult(byteData)
	if err != nil {
		log.Errorf("Error decode dumpprivkey: %v", err)
		return nil, err
	}
	return prvkeyResult, nil
}

//导入私钥,去除从头扫描
func (o *OmniClient) RpcImportprivkey(prvkey string) (*models.OmniImportprivkeyResult, error) {
	var (
		prvkeyResult *models.OmniImportprivkeyResult
	)
	c := util.NewRpcClient(o.ConnCfg)
	byteData, err := c.Call("importprivkey", prvkey, "", false)
	if err != nil {
		log.Errorf("Error importprivkey: %v", err)
		return nil, err
	}
	prvkeyResult, err = util.DecodeOmniImportprivkeyResult(byteData)
	if err != nil {
		log.Errorf("Error decode importprivkey: %v", err)
		return nil, err
	}
	return prvkeyResult, nil

}

//导入私钥,从头扫描,是否从头扫描
func (o *OmniClient) RpcImportAddrs(addr, label string, rescan bool) (*models.ImportaddrResult, error) {
	var (
		result *models.ImportaddrResult
	)
	c := util.NewRpcClient(o.ConnCfg)
	byteData, err := c.Call("importaddress", addr, label, rescan)
	if err != nil {
		log.Errorf("Error importaddress: %v", err)
		return nil, err
	}
	fmt.Println(string(byteData))
	result, err = util.DecodeImportaddrResult(byteData)
	if err != nil {
		log.Errorf("Error decode importaddress: %v", err)
		return nil, err
	}
	return result, nil
}
