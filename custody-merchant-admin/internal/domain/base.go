package domain

import "github.com/shopspring/decimal"

type ChainInfo struct {
	Id   int64  `json:"id" form:"id"`
	Name string `json:"name" form:"name"`
}

type CoinInfo struct {
	Id       int64           `json:"id" form:"id"`
	ChainId  int64           `json:"chain_id" form:"chain_id"` //
	Name     string          `json:"name" form:"name"`
	PriceUsd decimal.Decimal `json:"price_usd" form:"price_usd"` //
	FullName string          `json:"full_name"  form:"full_name" `
}
