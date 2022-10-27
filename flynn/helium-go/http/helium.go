package http

import (
	"encoding/json"
	"errors"
	"fmt"
)

type HeliumRpc struct {
	url string
}

func NewHeliumRpc(nodelUrl string) *HeliumRpc {
	hr := new(HeliumRpc)
	hr.url = nodelUrl
	return hr
}

/*
获取当前区块最新高度
*/
func (hr *HeliumRpc) GetLatestBlockHeight() (int64, error) {
	url := hr.url + "/v1/blocks/height"
	var data map[string]map[string]int64
	err := HttpGet(url, &data)
	if err != nil {
		return -1, err
	}
	if data == nil || data["data"] == nil {
		return -1, errors.New("get latest block height error,response data is null")
	}
	d := data["data"]
	return d["height"], nil
}

/*
获取区块所在高度
*/
func (hr *HeliumRpc) GetBlock(height int64) (*RespBlock, error) {
	url := fmt.Sprintf("%s/v1/blocks/%d", hr.url, height)
	var data map[string]RespBlock
	err := HttpGet(url, &data)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("get block by height error,response data is null,height=%d", height)
	}
	resp := data["data"]
	if &resp == nil {
		return nil, fmt.Errorf("get  block by height error,response data.data is null,height=%d", height)
	}
	return &resp, nil
}

/*
根据高度获取指定区块中的交易数据
*/
func (hr *HeliumRpc) GetBlockTransactionByHeight(height int64, cursor string) (*RespTransaction, error) {
	var url string
	if cursor == "" {
		url = fmt.Sprintf("%s/v1/blocks/%d/transactions", hr.url, height)
	} else {
		url = fmt.Sprintf("%s/v1/blocks/%d/transactions?cursor=%s", hr.url, height, cursor)
	}

	var data RespTransaction

	err := HttpGet(url, &data)
	if err != nil {
		return nil, err
	}
	if data.Cursor != "" {
		data2, err := hr.GetBlockTransactionByHeight(height, data.Cursor)
		if err != nil {
			return nil, err
		}
		data.Data = append(data.Data, data2.Data...)
	}

	if &data == nil {
		return nil, fmt.Errorf("get transaction by height error,response data is null,height=%d", height)
	}

	//return data["data"],nil
	return &data, nil
}

func (hr *HeliumRpc) GetBlockByHash(hash string) (*RespBlock, error) {
	url := fmt.Sprintf("%s/v1/blocks/%s", hr.url, hash)
	fmt.Println(url)
	var data map[string]RespBlock
	err := HttpGet(url, &data)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("get block by hash error,response data is null,hash=%s", hash)
	}
	resp := data["data"]
	if &resp == nil {
		return nil, fmt.Errorf("get  block by hash error,response data.data is null,hash=%s", hash)
	}
	return &resp, nil
}

/*
根据txid获取交易信息
*/
func (hr *HeliumRpc) GetTransactionByTxid(hash string) (*RespPaymentTransaction, error) {
	url := fmt.Sprintf("%s/v1/transactions/%s", hr.url, hash)
	var data map[string]RespPaymentTransaction
	err := HttpGet(url, &data)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("get transaction by txid error,response data is null,Txid=%s", hash)
	}
	resp := data["data"]
	if &resp == nil {
		return nil, fmt.Errorf("get transaction by txid error,response data.data is null,Txid=%s", hash)
	}
	return &resp, nil
}

/*
根据txid获取交易的状态
*/
func (hr *HeliumRpc) GetPendingTransactionByTxid(txid string) (*RespPendingStatus, error) {
	url := fmt.Sprintf("%s/v1/pending_transactions/%s", hr.url, txid)
	var data map[string]RespPendingStatus
	err := HttpGet(url, &data)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("get pending transaction by txid error,response data is null,Txid=%s", txid)
	}
	resp := data["data"]
	if &resp == nil {
		return nil, fmt.Errorf("get pending transaction by txid error,response data.data is null,Txid=%s", txid)
	}
	return &resp, nil
}

/*
发送交易
*/
func (hr *HeliumRpc) BroadcastTransaction(txn string) (string, error) {
	url := fmt.Sprintf("%s/v1/pending_transactions", hr.url)
	params := map[string]string{
		"txn": txn,
	}
	reqData, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	postData := string(reqData)
	var data map[string]RespBroadcastTx
	err = HttpPost(url, postData, &data)
	if err != nil {
		return "", err
	}
	if data == nil {
		return "", fmt.Errorf("broadcast transaction error,data is null,txn=[%s]", txn)
	}
	resp := data["data"]
	if &resp == nil {
		return "", fmt.Errorf("broadcast transaction error,data.data is null,txn=[%s]", txn)
	}
	return resp.Hash, nil
}

/*
获取所有热点
*/
func (hr *HeliumRpc) GetHotspots() {
	url := fmt.Sprintf("%s/v1/hotspots", hr.url)
	var data map[string]interface{}
	err := HttpGet(url, &data)
	if err != nil {
		panic(err)
	}
}

/*
根据地址获取热点
*/
func (hr *HeliumRpc) GetHotspotsByAddress(address string) {
	url := fmt.Sprintf("%s/v1/hotspots/%s", hr.url, address)
	var data map[string]interface{}
	err := HttpGet(url, &data)
	if err != nil {
		panic(err)
	}
}

/*
获取所有账户信息
*/
func (hr *HeliumRpc) GetAccounts() {
	url := fmt.Sprintf("%s/v1/accounts", hr.url)
	var data map[string]interface{}
	err := HttpGet(url, &data)
	if err != nil {
		panic(err)
	}
}

/*
根据地址获取账户信息
*/
func (hr *HeliumRpc) GetAccountByAddress(address string) (*RespAccount, error) {
	url := fmt.Sprintf("%s/v1/accounts/%s", hr.url, address)
	var data map[string]RespAccount
	err := HttpGet(url, &data)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, fmt.Errorf("get account by address errpr,data is null,address=%s", address)
	}
	resp := data["data"]
	if &resp == nil {
		return nil, fmt.Errorf("get account by address errpr,data.data is null,address=%s", address)
	}
	return &resp, nil
}

/*
根据账户获取热点
*/
func (hr *HeliumRpc) GetHotpotsByAccount(address string) {
	url := fmt.Sprintf("%s/v1/accounts/%s/hotspots", hr.url, address)
	var data map[string]interface{}
	err := HttpGet(url, &data)
	if err != nil {
		panic(err)
	}
}

func (hr *HeliumRpc) GetVars() (*RespVars, error) {
	url := fmt.Sprintf("%s/v1/vars", hr.url)
	var data map[string]RespVars
	err := HttpGet(url, &data)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, errors.New("get vars error,resp data is null")
	}
	resp := data["data"]
	if &resp == nil {
		return nil, errors.New("parse resp data error, data is null")
	}
	return &resp, nil
}
func (hr *HeliumRpc) GetCurrentPrices() (*RespCurrentPrices, error) {
	url := fmt.Sprintf("%s/v1/oracle/prices/current", hr.url)
	var data map[string]RespCurrentPrices
	err := HttpGet(url, &data)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, errors.New("get current prices error,resp data is null")
	}
	resp := data["data"]
	if &resp == nil {
		return nil, errors.New("parse resp current price error, data is null")
	}
	return &resp, nil
}
