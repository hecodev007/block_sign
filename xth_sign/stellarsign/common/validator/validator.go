package validator

import (
	"github.com/eoscanada/eos-go"
	"github.com/shopspring/decimal"
)

//这个是通用的请求参数，不要改，要改另开struct
type CreateAddressParams struct {
	Num      int    `json:"num" binding:"min=1,max=50000"`
	OrderId  string `json:"order_no" binding:"required"`
	MchName  string `json:"mch_name" binding:"required"`
	CoinName string `json:"coin_name" binding:"required"`
}

type CreateAddressReturns struct {
	Code    int                       `json:"code"`
	Message string                    `json:"message"`
	Data    CreateAddressReturns_data `json:"data"`
}

type CreateAddressReturns_data struct {
	CreateAddressParams
	Address []string `json:"address"`
}

type SignHeader struct {
	MchId    string `json:"mch_no" `
	MchName  string `json:"mch_name" binding:"required"`
	OrderId  string `json:"order_no" `
	CoinName string `json:"coin_name" `
}

type SignParams struct {
	SignHeader
	SignParams_Data
}
type SignParams_Data struct {
	FromAddress string          `json:"from_address" binding:"required"`
	ToAddress   string          `json:"to_address" binding:"required"`
	Token       string          `json:"token"`                     //telos主币是：“”
	Value       decimal.Decimal `json:"amount" binding:"required"` //
	Memo        string          `json:"memo"`
	Seed        string
}

type SignReturns_data struct {
	eos.PackedTransaction
	TxHash interface{} `json:"txid"`
}

type SignReturns struct {
	SignHeader
	Code    int    `json:"code"`
	Message string `json:"message"`
	Rawtx   string `json:"rawtx"`
	Data    string `json:"data"`
}

type TransferReturns struct {
	SignHeader
	Code    int    `json:"code"`
	Message string `json:"message"`
	Rawtx   string `json:"rawtx"`
	Data    string `json:"txid"` //txid
}

type GetBalanceParams struct {
	CoinName string `json:"coin_name"`
	Token    string `json:"token"`
	Address  string `json:"address" binding:"required"`
}

type TrustLineParams struct {
	MchName string `json:"mch_name" binding:"required"`
	Token   string `json:"token" binding:"required"`
	Address string `json:"address" binding:"required"`
	Seed    string
}
type GetBalanceReturns struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"` //数值
}
type ValidAddressParams struct {
	Address string `json:"address"`
}
type ValidAddressReturns struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    bool   `json:"data"` //数值
}
