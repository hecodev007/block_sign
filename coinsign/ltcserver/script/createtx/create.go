package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/ltcserver/model/bo"
	"github.com/group-coldwallet/ltcserver/util"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"log"
	"net/http"
)

type utxoresult struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    []*utxo `json:"data"`
}

type utxo struct {
	Txid         string          `json:"txid"`
	Vout         uint32          `json:"vout"`
	Address      string          `json:"address"`
	ScriptPubKey string          `json:"scriptPubKey"`
	Amount       decimal.Decimal `json:"amount"`
}

func main() {
	addrs, err := util.ReadCsv("/Users/zwj/gopath/src/github.com/group-coldwallet/ltcserver/script/createtx/ltc.csv", 0)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	addrs = removeRepByLoop(addrs)
	if len(addrs) == 0 {
		panic(fmt.Sprintf("len(addrs):%d", len(addrs)))
	}

	txIns := make([]bo.LtcTxInTpl, 0)
	txOuts := make([]bo.LtcTxOutTpl, 0)

	dd, _ := json.Marshal(addrs)
	datas, err := doBytesPost("http://47.244.140.180:9999/api/v1/ltc/unspents", dd)
	if err != nil {
		panic(err)
	}
	utxos := new(utxoresult)
	err = json.Unmarshal(datas, utxos)
	if err != nil {
		panic(err)
	}
	if len(utxos.Data) == 0 {
		panic(err)
	}

	totalAmount := decimal.Zero
	for _, v := range utxos.Data {
		totalAmount = totalAmount.Add(v.Amount)
		txIns = append(txIns, bo.LtcTxInTpl{
			FromAddr:   v.Address,
			FromTxid:   v.Txid,
			FromIndex:  v.Vout,
			FromAmount: v.Amount.IntPart(),
		})
	}
	fmt.Printf("totalAmount :%s", totalAmount.Shift(-8).String())

	fee, _ := decimal.NewFromString("0.0005")
	toAmount := totalAmount.Sub(fee.Shift(8))
	txOuts = append(txOuts, bo.LtcTxOutTpl{
		//ToAddr: "MMytC9qE9S57evaRjPSmNCjdkUvZT9rFfP",
		ToAddr:   "LRSgpmsHjCfLpsfxtPfUM1oxj5q7TXw6wn",
		ToAmount: toAmount.IntPart(),
	})
	tpl := &bo.LtcTxTpl{
		TxIns:  txIns,
		TxOuts: txOuts,
	}

	d, _ := json.Marshal(tpl)
	fmt.Printf("tpl:%s", string(d))

}

//body提交二进制数据
func doBytesPost(url string, data []byte) ([]byte, error) {
	body := bytes.NewReader(data)
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Printf("http.NewRequest,[err=%s][url=%s] \n", err.Error(), url)
		return []byte(""), err
	}
	request.Header.Set("Content-Type", "application/json")
	var resp *http.Response
	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("http.Do failed,[err=%s][url=%s] \n", err.Error(), url)
		return []byte(""), err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("http.Do failed,[err=%s][url=%s] \n", err.Error(), url)
	}
	return b, err
}

// 通过两重循环过滤重复元素
func removeRepByLoop(slc []string) []string {
	result := []string{} // 存放结果

	for i, v := range slc {
		if v == "" {
			continue
		}
		flag := true
		for j, _ := range result {
			if slc[i] == result[j] {
				flag = false // 存在重复元素，标识为false
				break
			}
		}
		if flag { // 标识为false，不添加进结果
			result = append(result, slc[i])
		}
	}
	return result
}
