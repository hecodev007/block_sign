package avax

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"errors"
	"testing"


)
func Test_txfee(t *testing.T){
	fee,err :=GetTxFee("https://api.avax.network")
	if err != nil {
		panic(err)
	}
	fmt.Println("fee:",fee)
	utxos,err :=GetUtxos("https://api.avax.network","X-avax17q7l3eantj70yq7fy4adp66gkyy5fnnd7celpr")
	if err != nil {
		panic(err)
	}
	fmt.Println("utxos:",utxos)
}

func GetTxFee(host string) (int64,error){
	host += "/ext/info"
	params:= struct {
		Id      string        `json:"id"`
		Jsonrpc string        `json:"jsonrpc"`
		Method  string        `json:"method"`
	}{
		Id:"test",
		Jsonrpc: "2.0",
		Method: "info.getTxFee",
	}
	body, _ := json.Marshal(params)
	req, err := http.NewRequest("POST", host, bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return 0, err
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, fmt.Errorf("ReadAll err: %v", err)
	}
	ret := &struct {
		ID      string          `json:"id"`
		JSONRPC string          `json:"jsonrpc"`
		Result  struct{	 
			TxFee string `json:"txFee"`
		} `json:"result"`
		Error   struct{
			Code int `json:"code"`
			Message string `json:"message"`
		}          `json:"error"`
	}{}
	if err = json.Unmarshal(data,ret);err != nil{
		return 0,err
	}
	if ret.Error.Code!=0{
		return 0,errors.New(ret.Error.Message)
	}
	return strconv.ParseInt(ret.Result.TxFee,10,64)
}
func GetUtxos(host string,address ...string )([]string,error){
	host += "/ext/bc/X"
	params:= struct {
		Id      string        `json:"id"`
		Jsonrpc string        `json:"jsonrpc"`
		Method  string        `json:"method"`
		Params struct{
			Addresses []string `json:"addresses"`
			Limit int `json:"params"`
		} `json:"params"`
	}{
		Id:"test",
		Jsonrpc: "2.0",
		Method: "avm.getUTXOs",
		Params:struct{
			Addresses []string `json:"addresses"`
			Limit int `json:"params"`
		}{Addresses: address,Limit:100},
	}
	body, _ := json.Marshal(params)
	req, err := http.NewRequest("POST", host, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("ReadAll err: %v", err)
	}
	ret := &struct {
		ID      string          `json:"id"`
		JSONRPC string          `json:"jsonrpc"`
		Result  struct{
			Utxos []string `json:"utxos"`
		} `json:"result"`
		Error   struct{
			Code int `json:"code"`
			Message string `json:"message"`
		}          `json:"error"`
	}{}
	if err = json.Unmarshal(data,ret);err != nil{
		return nil,err
	}
	if ret.Error.Code!=0{
		return nil,errors.New(ret.Error.Message)
	}
	return ret.Result.Utxos,nil
}
