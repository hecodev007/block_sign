package stx

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type RpcClient struct {
	client *http.Client
	Url    string
}
func NewRpcClient(url, user, pwd string) *RpcClient {
	client := &RpcClient{
		client: http.DefaultClient,
		Url:    url,
	}
	return client
}
func (rpc *RpcClient)GetBlockCount()(int64,error){
	resp,err :=http.Get(rpc.Url+"/v2/info")
	if err != nil {
		return 0,err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0,err
	}
	h :=gjson.Get(string(bytes),"stacks_tip_height").Int()
	return h,nil
}
func (rpc *RpcClient)SendRawTransaction(rawTx string)(txid string,err error){
	rawTxBytes ,_ := hex.DecodeString(rawTx)
	resp,err :=http.Post(rpc.Url+"/v2/transactions","application/octet-stream",bytes.NewReader(rawTxBytes))
	if err != nil {
		return  "",err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		return "", err
	}
	errMsg := gjson.Get(string(body),"error").String()
	reason := gjson.Get(string(body),"reason").String()
	if errMsg != "" || reason != "" {
		return "",errors.New("交易广播失败,"+errMsg+":"+reason)
	}
	txid = strings.Replace(string(body),"\"","",-1)

	return "0x"+strings.TrimPrefix(txid,"0x"),nil
}
func (rpc *RpcClient)GetBalance(addr string) (decimal.Decimal,error)  {

	resp,err :=http.Get(rpc.Url+"/extended/v1/address/"+addr+"/balances")
	if err != nil {
		return decimal.Zero,err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return decimal.Zero,err
	}
	balanceStr :=gjson.Get(string(bytes),"stx.balance").String()
	value ,err :=decimal.NewFromString(balanceStr)
	if err != nil {
		return decimal.Zero,err
	}
	value = value.Shift(-6)
	return value,nil
}

func (rpc *RpcClient) GetNonce(addr string) (uint64,error)  {
	resp,err :=http.Get(rpc.Url+"/v2/accounts/"+addr)
	if err != nil {
		return 0,err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0,err
	}
	nonce :=gjson.Get(string(bytes),"nonce").Uint()
	return nonce,nil
}

func (rpc *RpcClient) GetNonceAndBalance(addr string)(n uint64,b uint64,err error){
	resp,err :=http.Get(rpc.Url+"/v2/accounts/"+addr)
	if err != nil {
		return 0,0,err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0,0,err
	}
	nonce :=gjson.Get(string(bytes),"nonce").Uint()
	balanceStr := gjson.Get(string(bytes),"balance").String()
	balance ,err :=strconv.ParseUint(strings.TrimPrefix(balanceStr,"0x"),16,64)
	if err != nil {
		return 0,0,err
	}

	return nonce,balance,nil
}