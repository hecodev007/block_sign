package okt

import "github.com/shopspring/decimal"

type ResponseBlock struct {
	Jsonrpc string `json:"jsonrpc"`
	Id int `json:"id"`
	Result struct{
		Response struct{
			Data string `json:"data"`
			LastBlockHeight decimal.Decimal `json:"last_block_height"`
			LastBlockAppHash string `json:"last_block_app_hash"`
		} `json:"response"`
	} `json:"result"`
}