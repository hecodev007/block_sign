package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/shopspring/decimal"
)

type TrxScan struct {
}
type AccountReponse struct {
	Latest_opration_time int64             `json:"latest_opration_time"`
	WithPriceTokens      []*withPriceToken `json:"withPriceTokens"`
}

type withPriceToken struct {
	Amount       decimal.Decimal `json:"amount"`
	Balance      decimal.Decimal `json:"balance"`
	TokenDecimal int32           `json:"tokenDecimal"`
	TokenId      string          `json:"tokenId"`
}

func (sc *TrxScan) BalanceOf(addr string, contract string) (amount decimal.Decimal, t int64, err error) {
	if contract == "" {
		contract = "_"
	}
	resp, err := http.Get(fmt.Sprintf("https://apilist.tronscan.org/api/accountv2?address=%v", addr))
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		return amount, t, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	//log.Info(string(body))
	response := new(AccountReponse)
	if err = json.Unmarshal(body, response); err != nil {
		return
	}
	t = response.Latest_opration_time / 1000
	//log.Info(xutils.String(response))
	for _, v := range response.WithPriceTokens {
		if v.TokenId == contract {
			amount = v.Balance.Shift(0 - v.TokenDecimal)
			break
		}
	}
	return
}
