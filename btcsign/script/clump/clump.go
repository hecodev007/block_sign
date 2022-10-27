package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/btcsign/model/bo"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"sort"
)

//应对usdt零散的utxo 归集为一笔utxo (比如usdt会产生很多546的utxo无法使用，目的就是把这些归集为1笔)

func main() {
	//fromAddress, err := util.ReadCsv("/Users/zwj/gopath/src/github.com/group-coldwallet/btcsign/script/collection/addrs.csv", 0)
	//if err != nil {
	//	fmt.Println(err)
	//	panic(err)
	//}
	//
	//fromAddress = util.StringArrayRemoveRepeatByMap(fromAddress)
	//if len(fromAddress) == 0 {
	//	fmt.Errorf("error addr")
	//	return
	//}

	fromAddress := []string{"1PXC1aiZGXy4j81wt16g2rrUAsYVGUbuA7"}
	changeAddr := "1PXC1aiZGXy4j81wt16g2rrUAsYVGUbuA7"
	toAddress := "1PXC1aiZGXy4j81wt16g2rrUAsYVGUbuA7"
	fee, _ := decimal.NewFromString("0.005")

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
	count := 0

	for _, v := range sortBtcUnspent {
		if v.Amount >= 10000 {
			continue
		}
		//if v.Amount < 10000 {
		//	continue
		//}
		if v.Confirmations > 0 {
			continue
		}
		if count > 80 {
			//最多允许两个进来
			break
		}
		count++
		fmt.Println("添加：", v.Amount)
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

	fmt.Println("fromTotalAmount:", fromTotalAmount.String())
	//预想的输出总额
	toTotal := fromTotalAmount.Sub(fee)

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
	log.Infof("返回内容：%s", string(body))
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
