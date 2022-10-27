package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/group-coldwallet/zecserver/model/bo"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
	"sort"
)

func main() {
	fromAddress := []string{"t1fDakeqAPtujmn4Q4tDzUz2a22dgE9X8Fj", "t1fcq9dBEk3MYtBse82j7qVNLBkwFxj6NEJ", "t1fj543Dv5hJFRzM9A5DL9jLizSyDeDc2E5", "t1fCqZWKBud6NgnpQDs8QMpJn2xEu2QD6zC", "t1fCmVwQnYnBLmfx23Z65MiEZQR9DhgUkzC", "t1eLkvRDTqiuwyBtQNRVDqdddR8176rkiKn", "t1fjEDZaCYA3dPgvrKf9tAZLUkNFHMZt51z", "t1fGMsNQX76RKWYywfMoNbigto81Xe8f3uX"}
	changeAddr := "t1fDakeqAPtujmn4Q4tDzUz2a22dgE9X8Fj"
	toAddress := "t1fDakeqAPtujmn4Q4tDzUz2a22dgE9X8Fj"
	fee, _ := decimal.NewFromString("0.0001")
	tpl, total, err := createMakeTpl(10, fromAddress, toAddress, fee, changeAddr)
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(tpl)
	fmt.Println(string(data))
	fmt.Println("total:", total)
}

//参数：
//fromAddress 来源地址，
//toAddress 接收地址，
//toZec 接收的btc
//fee 手续费
//changeAddr 找零ZEC地址，
func createMakeTpl(utxoNum int, fromAddress []string, toAddress string, fee decimal.Decimal, changeAddr string) (*bo.ZecTxTpl, decimal.Decimal, error) {

	txIns := make([]bo.ZecTxInTpl, 0)
	txOuts := make([]bo.ZecTxOutTpl, 0)

	//查询from地址的utxo可用金额
	data, _ := json.Marshal(fromAddress)
	dataByte, err := getUtxo(data)
	if err != nil {
		return nil, decimal.Zero, err
	}
	utxoResultVO := new(ZecListUnSpentResult)
	json.Unmarshal(dataByte, utxoResultVO)
	if len(utxoResultVO.Data) == 0 {
		return nil, decimal.Zero, errors.New(string(dataByte))
	}
	var sortZecUnspent ZecUnspentSliceDesc
	sortZecUnspent = append(sortZecUnspent, utxoResultVO.Data...)
	//排序unspent，先进行降序，找出大额的数值
	sort.Sort(sortZecUnspent)
	utxoResult := make([]ZecUnSpentVO, 0)
	num := utxoNum - 1
	for i, v := range sortZecUnspent {
		if i > num {
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
		txIns = append(txIns, bo.ZecTxInTpl{
			FromAddr:   v.Address,
			FromTxid:   v.Txid,
			FromAmount: v.Amount,
			FromIndex:  uint32(v.Vout),
		})
	}

	//预想的输出总额
	toTotal := fromTotalAmount.Sub(fee)

	txOuts = append(txOuts, bo.ZecTxOutTpl{
		ToAddr:   toAddress,
		ToAmount: toTotal.Shift(8).IntPart(),
	})

	txTpl := &bo.ZecTxTpl{
		TxIns:  txIns,
		TxOuts: txOuts,
	}
	return txTpl, toTotal, nil
}

func getUtxo(data []byte) ([]byte, error) {
	//jsonStr :=[]byte(`{ "username": "auto", "password": "auto123123" }`)
	url := "http://47.244.140.180:9999/api/v1/zec/unspents"
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
type ZecListUnSpentResult struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    []ZecUnSpentVO `json:"data"`
}
type ZecUnSpentVO struct {
	Txid          string `json:"txid"`
	Vout          int64  `json:"vout"`
	Address       string `json:"address"`
	ScriptPubKey  string `json:"scriptPubKey"`
	Amount        int64  `json:"amount"`
	Confirmations int64  `json:"confirmations"`
	Spendable     bool   `json:"spendable"`
	Solvable      bool   `json:"solvable"`
}

//========================================================ZEC========================================================
//ZEC unspents切片排序
type ZecUnspentSliceDesc []ZecUnSpentVO

//实现排序三个接口
//为集合内元素的总数
func (s ZecUnspentSliceDesc) Len() int {
	return len(s)
}

//Swap 交换索引为 i 和 j 的元素
func (s ZecUnspentSliceDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//从大到小，最大金额排序
//如果index为i的元素大于index为j的元素，则返回true，否则返回false
func (s ZecUnspentSliceDesc) Less(i, j int) bool {
	return s[i].Amount > s[j].Amount
}

//========================================================ZEC========================================================

//=====================================================listunspent======================================================
