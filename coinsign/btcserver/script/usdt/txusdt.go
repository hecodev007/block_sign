package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/btcserver/model/bo"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

//应对打错usdt资产的脚本
func main() {
	fromAddress := "34QTSjeJYmVKqseGzYL8Q5UGigVmq5DQjj"
	changeAddr := "34QTSjeJYmVKqseGzYL8Q5UGigVmq5DQjj"
	feeAddr := "34QTSjeJYmVKqseGzYL8Q5UGigVmq5DQjj"
	toAddress := "18R5EfZVSVCeq3FgqyLRbpTrqwhqdwSNCz"
	toUsdt, _ := decimal.NewFromString("519.00")
	toBtc, _ := decimal.NewFromString("0.00000546")
	//fee, _ := decimal.NewFromString("0.0001")
	fee, _ := decimal.NewFromString("0.0008")
	tpl, err := createMakeTpl(fromAddress, toAddress, feeAddr, toUsdt, toBtc, fee, changeAddr)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	data, _ := json.Marshal(tpl)
	fmt.Println(string(data))
}

//参数：
//fromAddress 来源地址，
//toAddress 接收地址，
//toBtc 接收的btc
//toUsdt 接收的usdt
//fee 手续费
//changeAddr 找零BTC地址，
func createMakeTpl(fromAddress, toAddress, feeAddr string, toUsdt, toBtc, fee decimal.Decimal, changeAddr string) (*bo.BtcTxTpl, error) {

	txIns := make([]bo.BtcTxInTpl, 0)
	txOuts := make([]bo.BtcTxOutTpl, 0)

	//查询from地址的utxo可用金额
	data := []byte{}
	if fromAddress == feeAddr {
		data, _ = json.Marshal([]string{fromAddress})
	} else {
		data, _ = json.Marshal([]string{fromAddress, feeAddr})
	}

	dataByte, err := getUtxo(data)
	if err != nil {
		log.Println("err err")
		return nil, err
	}
	utxoResult := new(BtcListUnSpentResult)
	json.Unmarshal(dataByte, utxoResult)
	if len(utxoResult.Data) == 0 {
		return nil, errors.New(string(dataByte))
	}

	//form地址总额
	fromTotalAmount := decimal.Zero
	for _, v := range utxoResult.Data {
		fa := decimal.New(v.Amount, -8)
		fromTotalAmount = fromTotalAmount.Add(fa)
		txIns = append(txIns, bo.BtcTxInTpl{
			FromAddr:   v.Address,
			FromTxid:   v.Txid,
			FromAmount: v.Amount,
			FromIndex:  uint32(v.Vout),
		})
	}

	//usdt附加在第一个from
	txIns[0].UsdtAmount = toUsdt.Shift(8).IntPart()

	//预想的输出总额
	toTotal := toBtc.Add(fee)
	if fromTotalAmount.LessThan(toTotal) {
		return nil, fmt.Errorf("from btc:%s,to :%s,fee:%s ", fromTotalAmount.String(), toBtc.String(), fee.String())
	}

	txOuts = append(txOuts, bo.BtcTxOutTpl{
		ToAddr:       toAddress,
		ToAmount:     toBtc.Shift(8).IntPart(),
		ToUsdtAmount: toUsdt.Shift(8).IntPart(),
	})

	//找零
	changeAmount := fromTotalAmount.Sub(toTotal)
	if !changeAmount.IsZero() {
		txOuts = append(txOuts, bo.BtcTxOutTpl{
			ToAddr:   changeAddr,
			ToAmount: changeAmount.Shift(8).IntPart(),
		})
	}

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

//=====================================================listunspent======================================================
