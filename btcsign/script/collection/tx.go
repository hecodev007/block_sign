package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/group-coldwallet/btcsign/model/bo"
	"github.com/group-coldwallet/btcsign/util"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

//合并utxo
func main() {
	fromAddress, err := util.ReadCsv("/Users/hoo/workspace/go/gopath/src/github.com/group-coldwalle/btcsign/script/collection/addrs.csv", 0)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	fromAddress = util.StringArrayRemoveRepeatByMap(fromAddress)
	if len(fromAddress) == 0 {
		fmt.Errorf("error addr")
		return
	}
	fmt.Println(len(fromAddress))

	//fromAddress := []string{"36XnuCAGhEy4hoc7eovrSwyadErUHexi8M", "3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw", "38YZAMTu99dVf7jmm9jVAZ1Z5UyDTN5HVy", "36DiaRKnqiU3g8N6eNHCaLsgoR9zosNV5G", "36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76", "3JppbtYBmyoRB3WZ9Xutqwwqgy1pfx3P7s"}
	changeAddr := "3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw"
	toAddress := "3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw"
	fee, _ := decimal.NewFromString("0.007")

	tpl, err := createMakeTpl(fromAddress, toAddress, fee, changeAddr)
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(tpl)
	fmt.Println(string(data))
}

//合并utxo，未使用
func main1() {
	fromAddress, err := util.ReadCsv("/Users/hoo/workspace/go/gopath/src/github.com/group-coldwalle/btcsign/script/collection/addrs.csv", 0)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	fromAddress = util.StringArrayRemoveRepeatByMap(fromAddress)
	if len(fromAddress) == 0 {
		fmt.Errorf("error addr")
		return
	}
	fmt.Println(len(fromAddress))

	//fromAddress := []string{"36XnuCAGhEy4hoc7eovrSwyadErUHexi8M", "3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw", "38YZAMTu99dVf7jmm9jVAZ1Z5UyDTN5HVy", "36DiaRKnqiU3g8N6eNHCaLsgoR9zosNV5G", "36p1iTj5sBzAYpV25e1vLe9Y77gwLFGv76", "3JppbtYBmyoRB3WZ9Xutqwwqgy1pfx3P7s"}
	changeAddr := "3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw"
	toAddress := "3J5ZgXpkCffMoDi1snLMw9bY5GCUxyN8nw"
	fee, _ := decimal.NewFromString("0.007")

	tpl, err := createMakeTpl(fromAddress, toAddress, fee, changeAddr)
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(tpl)
	fmt.Println(string(data))
}

//参数：
//fromAddress 来源地址，
//toAddress 接收地址，
//toBtc 接收的btc
//fee 手续费
//changeAddr 找零BTC地址，
func createMakeTpl(fromAddress []string, toAddress string, fee decimal.Decimal, changeAddr string) (*bo.BtcTxTpl, error) {

	txIns := make([]bo.BtcTxInTpl, 0)
	txOuts := make([]bo.BtcTxOutTpl, 0)

	//查询from地址的utxo可用金额
	data, _ := json.Marshal(fromAddress)
	dataByte, err := getUtxo(data)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(dataByte))
	utxoResultVO := new(BtcListUnSpentResult)
	json.Unmarshal(dataByte, utxoResultVO)
	if len(utxoResultVO.Data) == 0 {
		return nil, errors.New(string(dataByte))
	}
	var sortBtcUnspent BtcUnspentSliceDesc
	sortBtcUnspent = append(sortBtcUnspent, utxoResultVO.Data...)
	//排序unspent，先进行降序，找出大额的数值
	sort.Sort(sortBtcUnspent)
	utxoResult := make([]BtcUnSpentVO, 0)
	for i, v := range sortBtcUnspent {
		if strings.HasPrefix(v.Address, "1") {
			continue
		}
		if v.Address == "34QTSjeJYmVKqseGzYL8Q5UGigVmq5DQjj" {
			continue
		}

		if v.Confirmations <= 0 {
			continue
		}
		log.Print(v.Amount)
		if v.Amount > 700000000 {
			continue
		}
		//if v.Amount > 30000000 {
		//	continue
		//}
		if i > 100 {
			//最多允许两个进来
			break
		}
		utxoResult = append(utxoResult, v)
	}
	//form地址总额
	fromTotalAmount := decimal.Zero
	for _, v := range utxoResult {
		fa := decimal.New(v.Amount, -8)
		fromTotalAmount = fromTotalAmount.Add(fa)
		txIns = append(txIns, bo.BtcTxInTpl{
			FromAddr:   v.Address,
			FromTxid:   v.Txid,
			FromAmount: v.Amount,
			FromIndex:  uint32(v.Vout),
		})
	}

	//预想的输出总额
	fmt.Println(fromTotalAmount.String())
	fmt.Println(fee.String())
	toTotal := fromTotalAmount.Sub(fee)
	fmt.Println(toTotal.String())

	txOuts = append(txOuts, bo.BtcTxOutTpl{
		ToAddr:   toAddress,
		ToAmount: toTotal.Shift(8).IntPart(),
	})

	txTpl := &bo.BtcTxTpl{
		TxIns:  txIns,
		TxOuts: txOuts,
	}
	return txTpl, nil
}

func getUtxo(data []byte) ([]byte, error) {
	//jsonStr :=[]byte(`{ "username": "auto", "password": "auto123123" }`)
	url := "http://47.244.140.180:9999/api/v1/btc/unspents"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: 120 * time.Second,
	}
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
type BtcListUnSpentResult struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    []BtcUnSpentVO `json:"data"`
}
type BtcUnSpentVO struct {
	Txid          string `json:"txid"`
	Vout          int64  `json:"vout"`
	Address       string `json:"address"`
	ScriptPubKey  string `json:"scriptPubKey"`
	Amount        int64  `json:"amount"`
	Confirmations int64  `json:"confirmations"`
	Spendable     bool   `json:"spendable"`
	Solvable      bool   `json:"solvable"`
}

//========================================================BTC========================================================
//BTC unspents切片排序
type BtcUnspentSliceDesc []BtcUnSpentVO

//实现排序三个接口
//为集合内元素的总数
func (s BtcUnspentSliceDesc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s BtcUnspentSliceDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s BtcUnspentSliceDesc) Less(i, j int) bool {
	return s[i].Amount > s[j].Amount
}

//========================================================BTC========================================================
