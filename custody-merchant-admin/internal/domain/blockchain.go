package domain

import "github.com/shopspring/decimal"

//BCBaseInfo 钱包返回币info
type BCBaseInfo struct {
	Code    int         `json:"code"`
	Status  int         `json:"status"`
	Msg     string      `json:"msg"`
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}

//BCCoinInfo 钱包返回币info
type BCCoinInfo struct {
	Father   string `json:"father" form:"father"`
	Name     string `json:"name" form:"name"`
	Token    string `json:"token" form:"token"`
	Decimal  int64  `json:"decimal" form:"decimal"`
	State    int64  `json:"state" form:"state"`
	Confirm  int64  `json:"confirm" form:"confirm"`
	FullName string `json:"full_name" form:"full_name"`
	PriceUsd string `json:"price_usd" form:"price_usd"`
}

//BCCoinInfo 钱包返回币info
type BCMchReq struct {
	ClientId   string `json:"client_id" form:"client_id"`
	Name       string `json:"name" form:"name"`
	Phone      string `json:"phone" form:"phone"`
	Email      string `json:"email" form:"email"`
	CompanyImg string `json:"company_img" form:"company_img"`
}

type BCMchInfo struct {
	ClientId string `json:"client_id" form:"client_id"`
	Secret   string `json:"secret" form:"secret"`
}

type BCWithDrawReq struct {
	ApiKey          string          `json:"api_key" form:"api_key"`   //商户的clientID
	CallBack        string          `json:"callBack" form:"callBack"` //回调地址，已有默认值"/custody/blockchain/callback"
	OutOrderId      string          `json:"outOrderId" form:"outOrderId"`
	CoinName        string          `json:"coinName" form:"coinName"`
	Amount          decimal.Decimal `json:"amount" form:"amount"`
	ToAddress       string          `json:"toAddress" form:"toAddress"`
	TokenName       string          `json:"tokenName" form:"tokenName"`
	ContractAddress string          `json:"contractAddress" form:"contractAddress"` //合约地址
	Memo            string          `json:"memo" form:"memo"`

	//Fee             string `json:"fee" form:"fee"`                         //不知含义
	//IsForce         string `json:"isorce" form:"isForce"`                  //不知含义
}

type InComeBack struct {
	Amount          float64 `json:"amount" form:"amount"`
	Apinonce        string  `json:"apinonce" form:"apinonce"`
	Apisign         string  `json:"apisign" form:"apisign"`
	Apits           string  `json:"apits" form:"apits"`
	BlockHeight     int64   `json:"block_height" form:"block_height"`
	ClientId        string  `json:"client_id" form:"client_id"`
	Coin            string  `json:"coin" form:"coin"`
	CoinType        string  `json:"coin_type" form:"coin_type"`
	ConfirmTime     int64   `json:"confirm_time" form:"confirm_time"`
	Confirmations   int64   `json:"confirmations" form:"confirmations"`
	ContractAddress string  `json:"contract_address" form:"contract_address"`
	Fee             float64 `json:"fee" form:"fee"` //手续费
	FromAddress     string  `json:"form_address" form:"form_address"`
	FromTrxId       string  `json:"form_trx_id" form:"form_trx_id"`
	IsIn            int     `json:"is_in" form:"is_in"` //是否转入 1 转入 2 转出
	Memo            string  `json:"memo" form:"memo"`
	Timestamp       int     `json:"timestamp" form:"timestamp"`
	ToAddress       string  `json:"to_address" form:"to_address"`
	ToRawAddress    string  `json:"to_raw_address" form:"to_raw_address"`
	TransactionId   string  `json:"transaction_id" form:"transaction_id"`
	TrxN            int     `json:"trx_n" form:"trx_n"`
	Txid            string  `json:"txid" form:"txid"`
	UserSubId       int     `json:"user_sub_id" form:"user_sub_id"`
	MemoEncrypt     string  `json:"memo_encrypt" form:"memo_encrypt"`
	Msg             string  `json:"msg" form:"msg"`
	IsRisk          bool    `json:"is_risk" form:"is_risk"`
	RiskLevel       int     `json:"risk_level" form:"risk_level"`
	RiskMsg         string  `json:"risk_msg" form:"risk_msg"`

	OutOrderId   string `json:"outOrderId" form:"outOrderId"`
	OuterOrderNo string `json:"outer_order_no" form:"outer_order_no"`
}
