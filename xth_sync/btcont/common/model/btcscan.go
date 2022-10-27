package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/shopspring/decimal"
)

type Scan struct {
}
type ResponseTxs struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		Type          string          `json:"type"`
		TxCount       int             `json:"txCount"`
		Spend         decimal.Decimal `json:"spend"`
		Receive       decimal.Decimal `json:"receive"`
		NormalTxCount int             `json:"normalTxCount"`
		Txs           []*Tx           `json:"txs"`
	} `json:"data"`
}
type Tx struct {
	Time    int64  `json:"time"`
	Txid    string `json:"txid"`
	Height  int64  `json:"height"`
	Fee     decimal.Decimal
	Inputs  []*Input  `json:"inputs"`
	Outputs []*Output `json:"outputs"`
}

type Input struct {
	InputNo       int             `json:"input_no"`
	Address       string          `json:"address"`
	Value         decimal.Decimal `json:"value"`
	Received_from struct {
		InputNo int    `json:"input_no"`
		Txid    string `json:"txid"`
	}
}
type Output struct {
	OutputNo int             `json:"output_no"`
	Address  string          `json:"address"`
	Value    decimal.Decimal `json:"value"`
}

func (sc *Scan) ListByAddr(addr string, page int, num int) (decimal.Decimal, []*Tx, error) {
	resp, err := http.Get(fmt.Sprintf("https://btc.tokenview.com/api/address/btc/%v/%v/%v", addr, page, num))
	//fmt.Println(fmt.Sprintf("https://btc.tokenview.com/api/address/btc/%v/%v/%v",addr,page,num))
	if err != nil {
		return decimal.Decimal{}, nil, err
	}
	if resp.StatusCode != 200 {
		return decimal.Decimal{}, nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	//log.Info(string(body))
	response := new(ResponseTxs)
	if err = json.Unmarshal(body, response); err != nil {
		return decimal.Decimal{}, nil, err
	}
	if len(response.Data) != 1 {
		return decimal.Decimal{}, nil, errors.New(addr + " ListByAddr response error")
	}
	return response.Data[0].Receive.Add(response.Data[0].Spend), response.Data[0].Txs, nil
}
func (sc *Scan) AllTxsByAddr(addr string) (a decimal.Decimal, txs []*Tx, err error) {
	a, txs, err = sc.ListByAddr(addr, 1, 100000)
	if len(txs) == 100000 {
		return a, txs, errors.New(addr + " 交易数过多")
	}
	return
}

type BlockResponse struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Data []*Tx  `json:"data"`
}

func (sc *Scan) BlockByHeight(height int64) ([]*Tx, error) {
	resp, err := http.Get(fmt.Sprintf("https://btc.tokenview.com/api/tx/btc/%v/1/5000", height))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	//log.Info(string(body))
	response := new(BlockResponse)
	if err = json.Unmarshal(body, response); err != nil {
		return nil, err
	}
	if response.Code != 1 {
		return nil, errors.New(response.Msg)
	}
	return response.Data, nil
}
func (sc *Scan) ToMap(txs []*Tx) map[string]*Tx {
	ret := make(map[string]*Tx, 0)
	for k, v := range txs {
		ret[v.Txid] = txs[k]
	}
	return ret
}

func (sc *Scan) BalanceOf(addr string) (decimal.Decimal, int64, error) {
	amount, txs, err := sc.ListByAddr(addr, 1, 1)
	if len(txs) > 0 {
		return amount, txs[0].Time, err
	}

	return amount, 0, err
}

type TxResponse struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Data *Tx    `json:"data"`
}

func (sc *Scan) GetRanTransaction(txid string) (*Tx, error) {
	resp, err := http.Get(fmt.Sprintf("https://btc.tokenview.com/api/tx/btc/%v", txid))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	//log.Info(string(body))
	response := new(TxResponse)
	if err = json.Unmarshal(body, response); err != nil {
		return nil, err
	}
	if response.Code != 1 {
		return nil, errors.New(response.Msg)
	}
	return response.Data, nil
}
