package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

type EthScan struct {
}
type EthAccountReponse struct {
	Status bool   `json:"status"`
	Code   int64  `json:"code"`
	Msg    string `json:"msg"`
	Data   struct {
		Info struct {
			Balance decimal.Decimal `json:"balance"`
		} `json:"info"`

		Items []struct {
			Timestamp int64 `json:"timestamp"`
		} `json:"item"`
	} `json:"data"`
}

func (sc *EthScan) BalanceOf(addr string, contract string) (amount decimal.Decimal, t int64, err error) {
	if contract != "" {
		return decimal.Decimal{}, 0, errors.New("不支持代币")
	}
	resp, err := sc.Get(fmt.Sprintf("https://api.yitaifang.com/index/accountInfo/?page=1&address=%v", addr))
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return amount, t, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	//log.Info(string(body))
	response := new(EthAccountReponse)
	if err = json.Unmarshal(body, response); err != nil {
		return
	}
	if response.Status != true || response.Code != 10000 {
		return decimal.Decimal{}, 0, errors.New(response.Msg)
	}
	//log.Info(xutils.String(response))
	if len(response.Data.Items) > 0 {
		t = response.Data.Items[0].Timestamp
	}
	//log.Info(xutils.String(response))
	amount = response.Data.Info.Balance.Shift(-18)
	return
}

func (sc *EthScan) Get(url string) (*http.Response, error) {
	//log.Info(url)
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.90 Safari/537.36")

	return client.Do(req)

}
