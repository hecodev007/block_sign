package model

import "github.com/shopspring/decimal"

type TransferRequest struct {
	Sign            string          `json:"sign"`  //签名，用于身份验证
	Sfrom           string          `json:"sfrom"` //商户标识
	OutOrderId      string          `json:"outOrderId"`
	CoinName        string          `json:"coinName"`
	Amount          decimal.Decimal `json:"amount"`
	ToAddress       string          `json:"toAddress"`
	TokenName       string          `json:"tokenName,omitempty"`
	ContractAddress string          `json:"contractAddress,omitempty"`
	Memo            string          `json:"memo,omitempty"`
	Fee             decimal.Decimal `json:"fee,omitempty"`
	IsForce         bool            `json:"isForce,omitempty"`
}
