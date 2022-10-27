package models

import "github.com/shopspring/decimal"

type TxInput struct {
	Txins         []Txins         `json:"txins"`
	ChangeAddress string          `json:"changeAddress"` //找零地址
	ToAmount      decimal.Decimal `json:"toAmount"`      //发送USDT金额
	ToBtc         decimal.Decimal `json:"toBtc"`         //发送BTC金额，默认0.00000546
	ToAddress     string          `json:"toAddress"`     //发送地址
	Fee           decimal.Decimal `json:"fee"`           //矿工手续费0.00000001
	MchInfo
}

//未花费的余额
type Txins struct {
	Txid         string          `json:"txid"`                   //交易ID
	Vout         int             `json:"vout"`                   //vout位置
	Address      string          `json:"address"`                //来源地址
	ScriptPubKey string          `json:"scriptPubKey"`           //公钥
	Amount       decimal.Decimal `json:"amount"`                 //btc金额，例子:0.0003
	RedeemScript string          `json:"redeemScript,omitempty"` //多签赎回脚本，一般单签为空即可
}

type TxInputNew struct {
	Txinputs []TxInput `json:"txinputs"` //intputs

}
