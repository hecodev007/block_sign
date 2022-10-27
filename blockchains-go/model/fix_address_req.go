package model

import "github.com/shopspring/decimal"

type FixAddressReq struct {
	Days        int             `json:"days"`
	FeeAmount   decimal.Decimal `json:"feeAmount"`
	TotalAmount decimal.Decimal `json:"totalAmount"`
}

type OrderTxLinkReq struct {
	TxId    string `json:"txId"`
	OrderNo string `json:"orderNo"`
}
