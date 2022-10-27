package stx

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"strings"
	"fmt"
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

func (rpc *RpcClient) GetBlockByHeight(best, h int64)(*Block,error){
	url := fmt.Sprintf("%v%v%v",rpc.Url,"/extended/v1/block?limit=2&offset=",best-h)
	resp,err := http.Get(url)
	if err != nil {
		return nil,err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	total := gjson.Get(string(bytes),"total").Int()
	if total < best{
		return nil,errors.New("超过最大高度")
	}
	if total > best+1{
		return rpc.GetBlockByHeight(total,h)
	}
	result := new(BlocksResult)
	if err =json.Unmarshal(bytes, result);err != nil{
		return nil,err
	}
	return result.Results[total-best],nil
}

func (rpc *RpcClient)Getransaction(txid string)(*Transaction,error){
	resp,err :=http.Get(rpc.Url+"/extended/v1/tx/"+txid)
	if err != nil {
		return nil,err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil,err
	}
	tx := new(Transaction)
	if err = json.Unmarshal(bytes,tx);err != nil{
		return nil,err
	}
	if tx.Error != ""{
		return nil,errors.New(tx.Error)
	}
	return tx,nil
}