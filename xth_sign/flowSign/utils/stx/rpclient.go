package stx

import (
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"strings"
)

package stx

import (
"github.com/shopspring/decimal"
"github.com/tidwall/gjson"
"io/ioutil"
"net/http"
"strings"
"errors"
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
	resp,err :=http.Post(rpc.Url+"/v2/transactions","",strings.NewReader(rawTx))
	if err != nil {
		return  "",err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		return "", err
	}
	txid = gjson.Get(string(body),"txid").String()
	errMsg := gjson.Get(string(body),"error").String()
	reason := gjson.Get(string(body),"reason").String()
	if txid == ""{
		return "",errors.New("交易广播失败,"+errMsg+":"+reason)
	}
	return txid,nil
}
func (rpc *RpcClient)GetBalance(addr string) (decimal.Decimal,error)  {

	resp,err :=http.Get(rpc.Url+"extended/v1/address/"+addr+"/balances")
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
	value.Shift(-6)
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