package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/btcsign/model/bo"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
)

//3N2BqsCExr5KKiu2h9oWySMheTuKSLP85D
//3BikiVKR24jNrx6jkbN9cq3hJEALaZNeES
//32Yas4gZL9kettai9PiMg1HzRtpf44YmTH
//3Cyy6Tgd2v74LCpxnymogQHZfLsiZ7TMEP
//3CWT1fi4cGZjQeYUYkxfowr19m6g1KbjaL
//3DwbYbd3MCVFZQUVz9SmNmnYzfmJMukkn8
//3P57PRuZywYGoiH4SYXRHDLgCyfnQMaCE5
//3J2FrVNyxgY3WiyVohCuUVvQwm7KCdK23y
//37rKn7og8dCcLRasCsYFPFo6LtnfkgxurT
//38Uw23x5G5pgmQL9pJDUfacJrwzpXKqr7S
//3926pPAdi4FcaqpzyvLfLuoDqZyKKQzeMy
//3Aw82z6ngsEjdMjjEnqzMShkRWpwr2heZj
//3FYdH1FM7r7anXBr2vjNC6zxyDxkiawYq6

func main() {
	fromAddress := []string{"34QTSjeJYmVKqseGzYL8Q5UGigVmq5DQjj"}
	toAddress := "1PXC1aiZGXy4j81wt16g2rrUAsYVGUbuA7"
	changeAddr := "34QTSjeJYmVKqseGzYL8Q5UGigVmq5DQjj"
	tpl, err := createMakeTpl(fromAddress, toAddress, changeAddr, decimal.New(100000, -8), 50, decimal.New(500000, -8))
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(tpl)
	fmt.Println(string(data))
}

//参数：
//fromAddress 来源地址，
//toAddress 发送地址，
//changeAddr 找零地址，
//everyOutAmount 每个out输出金额
//maxOutNum 打散总量
//fee 手续费
func createMakeTpl(fromAddress []string, toAddress string, changeAddr string, everyOutAmount decimal.Decimal, maxOutNum int64, fee decimal.Decimal) (*bo.BtcTxTpl, error) {

	if len(fromAddress) == 0 || toAddress == "" || changeAddr == "" {
		return nil, errors.New("miss address")
	}
	if everyOutAmount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("everyOutAmount error")
	}
	if fee.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("fee error")
	}
	txIns := make([]bo.BtcTxInTpl, 0)
	txOuts := make([]bo.BtcTxOutTpl, 0)

	//查询from地址的utxo可用金额
	data, _ := json.Marshal(fromAddress)
	dataByte, err := getUtxo(data)
	if err != nil {
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
		am := decimal.New(v.Amount, -8)
		fromTotalAmount = fromTotalAmount.Add(am)
		txIns = append(txIns, bo.BtcTxInTpl{
			FromAddr:   v.Address,
			FromTxid:   v.Txid,
			FromAmount: v.Amount,
			FromIndex:  uint32(v.Vout),
		})
	}
	if fromTotalAmount.Equals(decimal.Zero) {
		return nil, fmt.Errorf("来源地址,金额不足打散：%s", fromTotalAmount.String())
	}
	//自动计算
	if maxOutNum == 0 {
		maxOutNum = fromTotalAmount.Sub(fee).Div(everyOutAmount).Floor().IntPart()
		if maxOutNum == 0 {
			return nil, fmt.Errorf("from地址金额：%s,不满足输出总额，输出个数：%d,每个金额:%s,手续费:%s",
				fromTotalAmount.String(), maxOutNum, everyOutAmount.String(), fee.String())
		}
	}

	//预想的输出总额
	toTotal := everyOutAmount.Mul(decimal.New(maxOutNum, 0)).Add(fee)
	if fromTotalAmount.LessThan(toTotal) {
		//不满足的情况下全出
		//maxOutNum = fromTotalAmount / everyOutAmount
		//重新计算输出总额
		//toTotal = (everyOutAmount * maxOutNum) + fee
		return nil, fmt.Errorf("from地址金额：%s,不满足输出总额，输出个数：%d,每个金额:%s,手续费:%s，总输出金额：%s",
			fromTotalAmount.String(), maxOutNum, everyOutAmount.String(), fee.String(), toTotal.String())
	}
	//找零
	changeAmount := fromTotalAmount.Sub(toTotal)
	if changeAmount.GreaterThan(decimal.Zero) {
		if changeAmount.GreaterThan(decimal.NewFromFloat(0.00000546)) {
			txOuts = append(txOuts, bo.BtcTxOutTpl{
				ToAddr:   changeAddr,
				ToAmount: changeAmount.Shift(8).IntPart(),
			})
		} else {
			fee = changeAmount.Add(fee)
		}

	}
	for i := int64(0); i < maxOutNum; i++ {
		txOuts = append(txOuts, bo.BtcTxOutTpl{
			ToAddr:   toAddress,
			ToAmount: everyOutAmount.Shift(8).IntPart(),
		})
	}
	txTpl := &bo.BtcTxTpl{
		TxIns:  txIns,
		TxOuts: txOuts,
	}

	//最后做一次检查，输出总金额对比是否相等
	outam := fee
	for _, v := range txOuts {
		am := decimal.New(v.ToAmount, -8)
		outam = outam.Add(am)
	}
	if !outam.Equals(fromTotalAmount) {
		return nil, fmt.Errorf("模板校验错误，来源总金额：%s,输出总金额：%s", fromTotalAmount.String(), outam.String())
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
