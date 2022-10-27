package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/bchserver/model/bo"
	"github.com/group-coldwallet/bchserver/util"
	"io/ioutil"
	"net/http"
	"sort"
)

func main() {
	//fromAddrs := "34QTSjeJYmVKqseGzYL8Q5UGigVmq5DQjj"
	fromAddrs, err := util.ReadCsv("/Users/zwj/gopath/src/github.com/group-coldwallet/bchserver/script/transf/bch.csv", 0)
	if err != nil {
		panic(err)
	}

	if len(fromAddrs) == 0 {
		panic("empty fromAddrs")
	}
	toAddress := "qp9ppr7tuvmxsy4cm72t5tup6g5tx8zs0qrrynuzdf"  //出账地址
	changeAddr := "qp9ppr7tuvmxsy4cm72t5tup6g5tx8zs0qrrynuzdf" //找零地址
	tpl, err := createMakeTpl(fromAddrs, toAddress, changeAddr, 96111171925, 10000)
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(tpl)
	fmt.Println(string(data))
}

//只选择满足出账的utxo
//参数：
//fromAddress 来源地址，
//toAddress 发送地址，
//changeAddr 找零地址，
//everyOutAmount 每个out输出金额
//maxOutNum 打散总量
//fee 手续费
func createMakeTpl(fromAddrs []string, toAddress string, changeAddr string, toAmount, fee int64) (*bo.BchTxTpl, error) {
	txIns := make([]bo.BchTxInTpl, 0)
	txOuts := make([]bo.BchTxOutTpl, 0)

	//查询from地址的utxo可用金额
	data, _ := json.Marshal(fromAddrs)
	dataByte, err := getUtxo(data)
	if err != nil {
		return nil, err
	}
	utxoResult := new(BchListUnSpentResult)
	json.Unmarshal(dataByte, utxoResult)
	if len(utxoResult.Data) == 0 {
		return nil, errors.New(string(dataByte))
	}

	//排序
	var sortBtcUnspent BchUnspentSliceDesc
	sortBtcUnspent = append(sortBtcUnspent, utxoResult.Data...)
	sort.Sort(sortBtcUnspent)

	//form地址总额
	fromTotalAmount := int64(0)
	for _, v := range sortBtcUnspent {

		fromTotalAmount = fromTotalAmount + v.Amount
		txIns = append(txIns, bo.BchTxInTpl{
			FromAddr:   v.Address,
			FromTxid:   v.Txid,
			FromAmount: v.Amount,
			FromIndex:  uint32(v.Vout),
		})

		if fromTotalAmount >= (toAmount + fee) {
			fmt.Println("满足")
			break
		}
	}

	fmt.Println(fmt.Sprintf("fromTotalAmount：%d", fromTotalAmount))

	if fromTotalAmount < toAmount {
		return nil, fmt.Errorf("fromTotalAmount :%d,toAmount:%d", fromTotalAmount, toAmount)
	}

	//out输出
	txOuts = append(txOuts, bo.BchTxOutTpl{
		ToAddr:   toAddress,
		ToAmount: toAmount,
	})
	//找零
	changeAmount := fromTotalAmount - toAmount - fee
	if changeAmount > 546 {
		txOuts = append(txOuts, bo.BchTxOutTpl{
			ToAddr:   changeAddr,
			ToAmount: changeAmount,
		})
	}
	txTpl := &bo.BchTxTpl{
		TxIns:  txIns,
		TxOuts: txOuts,
	}
	return txTpl, nil
}

func getUtxo(data []byte) ([]byte, error) {
	//jsonStr :=[]byte(`{ "username": "auto", "password": "auto123123" }`)
	url := "http://47.244.140.180:9999/api/v1/bch/unspents"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// handle error
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}
	return body, nil

}

//=====================================================listunspent======================================================
type BchListUnSpentResult struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    []BchUnSpentVO `json:"data"`
}
type BchUnSpentVO struct {
	Txid          string `json:"txid"`
	Vout          int64  `json:"vout"`
	Address       string `json:"address"`
	ScriptPubKey  string `json:"scriptPubKey"`
	Amount        int64  `json:"amount"`
	Confirmations int64  `json:"confirmations"`
	Spendable     bool   `json:"spendable"`
	Solvable      bool   `json:"solvable"`
}

//=====================================================listunspent======================================================

//Bch unspents切片排序
type BchUnspentSliceDesc []BchUnSpentVO

//实现排序三个接口
//为集合内元素的总数
func (s BchUnspentSliceDesc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s BchUnspentSliceDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s BchUnspentSliceDesc) Less(i, j int) bool {
	return s[i].Amount > s[j].Amount
}
