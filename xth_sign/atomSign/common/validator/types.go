package validator

import (
	"github.com/shopspring/decimal"
)
//这个是通用的请求参数，不要改，要改另开struct
type CreateAddressParams struct {
	Num      int    `json:"num" binding:"min=1,max=50000"`
	OrderNo  string `json:"order_no" binding:"required"`
	MchName  string `json:"mch_name" binding:"required"`
	CoinName string `json:"coin_name" binding:"required"`
}

//固定的几个参数
type Header struct {
	OrderNo  string `json:"order_no"`
	//MchId  int64 `json:"mch_id" binding:"required"`
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

type SignParams struct {
	Header
	SignParams_data
}

type SignParams_data struct {
	FromAddr      string `json:"from_addr"`
	ToAddr        string `json:"to_addr"`
	Amount_str        decimal.Decimal  `json:"amount"`
	Amount int64
	AccountNumber uint64 `json:"account_number"`
	ChainID       string `json:"chain_id"`
	Sequence      uint64 `json:"sequence"`
	Memo          string `json:"memo"`
	Gas           uint64 `json:"gas"`
	Fee           int64  `json:"fee"`
}
type SignReturns struct {
	Header
	Code    int64       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	TxHash  string      `json:"txid"`
}
