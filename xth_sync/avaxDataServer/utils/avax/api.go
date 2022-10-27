package avax

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

//节点通信api
//func (rpc *RpcClient) GetUtxos(address ...string) (utxos []string, err error) {
//	err = rpc.CallWithAuth("/ext/bc/X", "avm.getUTXOs", rpc.Credentials, &utxos, address)
//	return
//}
//func (rpc *RpcClient) GetBlockCount() (height int64, err error) {
//	resp := struct {
//		Height string `json:"height"`
//	}{}
//	err = rpc.CallWithAuth("/ext/P", "platform.getHeight", rpc.Credentials, &resp)
//	height, _ = strconv.ParseInt(resp.Height, 10, 64)
//	return
//}
//
//func (rpc *RpcClient) GetBlockByHeight(h int64) (block Block, err error) {
//	return Block{}, errors.New("GetBlockByHeight not implement")
//}
//func (rpc *RpcClient) GetBlockByHeight2(h int64) (block BlockWithTx, err error) {
//	return BlockWithTx{}, errors.New("GetBlockByHeight2 not implement")
//}
var host = "http://avax.rylink.io:20490/X"

//var host = "https://explorerapi.avax.network/x"

func (rpc *RpcClient) GetBlockCount() (height int64, err error) {
	return rpc.GetCount()
}
func GetTx(host string, txid string) (string, error) {
	host += "/ext/bc/X"
	params := struct {
		Id      string `json:"id"`
		Jsonrpc string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  struct {
			TxID string `json:"txID"`
		}
	}{
		Id:      "test",
		Jsonrpc: "2.0",
		Method:  "avm.getTx",
		Params: struct {
			TxID string `json:"txID"`
		}{TxID: txid},
	}
	log.Println(host, String(params))
	body, _ := json.Marshal(params)
	req, err := http.NewRequest("POST", host, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("ReadAll err: %v", err)
	}
	ret := &struct {
		ID      string `json:"id"`
		JSONRPC string `json:"jsonrpc"`
		Result  struct {
			Tx string `json:"tx"`
		} `json:"result"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{}
	if err = json.Unmarshal(data, ret); err != nil {
		return "", err
	}
	if ret.Error.Code != 0 {
		return "", errors.New(ret.Error.Message)
	}
	return ret.Result.Tx, nil
}
func GetTxFromScan(txid string) ([]byte,error){
	client := http.Client{
		Timeout: 5* time.Second,
	}
	resp,err := client.Get("https://explorerapi.avax.network/v2/transactions/"+txid)
	if err != nil {
		return nil, err
	}
	body,err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body,nil

}
func (rpc *RpcClient) GetRawTransaction(h string) (tx Tx, err error) {
	rawtx, err := GetTx(rpc.Node, h)
	if err != nil {
		return tx, err
	}
	log.Println(h, rawtx)
	tmptx, err := ParseRawTransaction(rawtx)
	if err != nil {
		return tx, err
	}
	log.Println("GetRawTransaction", h, String(tmptx))
	return *tmptx, nil
	//resp := struct {
	//	Tx string `json:"tx"`
	//}{}
	//err = rpc.CallWithAuth("/ext/bc/X", "avm.getTx", rpc.Credentials, &resp)
	//if err != nil {
	//	return tx, err
	//}
	//rawtx,err :=ParseRawTransaction(resp.Tx)
	//if err != nil {
	//	return tx,err
	//}
	//
	//return *rawtx, errors.New("GetRawTransaction not implement")
}
func (rpc *RpcClient) GetRawTransactionFromScan(h string) (tx Transaction, err error) {
	rawtx, err := GetTxFromScan(h)
	if err != nil {
		return tx, err
	}
	//log.Println(h, rawtx)
	tx = Transaction{}
	err = json.Unmarshal(rawtx, &tx)
	for k, _ := range tx.Inputs {

		amount ,err := decimal.NewFromString(tx.Inputs[k].Output.Amount)
		if err != nil {
			tx.Inputs[k].Output.Amount = "0"
			continue
		}
		tx.Inputs[k].Output.Amount = amount.Shift(-9).String()
	}

	for k, _ := range tx.Outputs {
		amount ,err := decimal.NewFromString(tx.Outputs[k].Amount)
		if err != nil {
			tx.Outputs[k].Amount = "0"
			continue
		}
		tx.Outputs[k].Amount = amount.Shift(-9).String()
	}
	return tx, nil
	//resp := struct {
	//	Tx string `json:"tx"`
	//}{}
	//err = rpc.CallWithAuth("/ext/bc/X", "avm.getTx", rpc.Credentials, &resp)
	//if err != nil {
	//	return tx, err
	//}
	//rawtx,err :=ParseRawTransaction(resp.Tx)
	//if err != nil {
	//	return tx,err
	//}
	//
	//return *rawtx, errors.New("GetRawTransaction not implement")
}
func (rpc *RpcClient) GetBlockByHash(h string) (block Block, err error) {
	return Block{}, errors.New("GetBlockByHash not implement")
}

//获取交易列表
func (rpc *RpcClient) ListTransaction(offset, limit int64) (list ListTransaction, err error) {
	//timestamp-asc timestamp-desc
	url := fmt.Sprintf(rpc.url+"/X/transactions?sort=timestamp-asc&offset=%d&limit=%d", offset, limit)
	//fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(s, &list)

	return
}

//获取交易列表
func (rpc *RpcClient) TopTransaction(limit int64) (list ListTransaction, err error) {
	//timestamp-asc timestamp-desc
	url := fmt.Sprintf(rpc.url+"/X/transactions?sort=timestamp-desc&offset=0&limit=%d", limit)
	//fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(s, &list)

	return
}

//获取交易总数
func (rpc *RpcClient) GetCount() (num int64, err error) {
	list, err := rpc.TopTransaction(1)
	if err != nil {
		return 0, err
	}
	return list.Transactions[0].Timestamp.Unix(), err
}

//根据高度获取交易
func (rpc *RpcClient) GetTransactionByHeight(h int64) (*Transaction, error) {
	if h == 0 {
		return nil, nil
	}
	list, err := rpc.ListTransaction(h-1, 1)
	if err != nil {
		return nil, err
	}
	if len(list.Transactions) > 0 {
		return list.Transactions[0], nil
	}
	return nil, nil
}

func (rpc *RpcClient) GetTransactionByHash(txid string) (tx Transaction, err error) {
	rawTx, err := rpc.GetRawTransaction(txid)
	//log.Println("GetTransactionByHash", txid, String(rawTx))
	//panic("")
	if err != nil {
		return tx, err
	}
	for k, input := range rawTx.UnsignedTx.Inputs {
		tmTx, err := rpc.GetRawTransaction(input.TxID)
		if err != nil {
			return tx, err
		}
		rawTx.UnsignedTx.Inputs[k].Input.Addresses = tmTx.UnsignedTx.Outputs[rawTx.UnsignedTx.Inputs[k].OutputIndex].Output.Addresses
	}
	//log.Println("GetRawTransactionend", String(rawTx))
	return ToTransaction(&rawTx, txid), nil

}

func String(d interface{}) string {
	str, _ := json.Marshal(d)
	return string(str)
}
