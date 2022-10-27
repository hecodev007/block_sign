package model

import "github.com/shopspring/decimal"

type CocosTransferParams struct {
	ReqBaseParams
	FromAddress string          `json:"fromaddress"` //发送地址
	ToAddress   string          `json:"toaddress"`   //接收地址
	ToAmount    decimal.Decimal `json:"toamount"`    //接收金额
	Memo        string          `json:"memo"`        //memo
	// write by flynn 2020-10-14
	AssetSymbol  string `json:"asset_symbol"`
	AssetId      string `json:"asset_id"` // asset_id
	AssetDecimal int32  `json:"asset_decimal"`
}
