package model

import "github.com/shopspring/decimal"
import "github.com/Dipper-Labs/go-sdk/client/types"

type DipNodeInfo struct {
	NodeInfo NodeInfo `json:"node_info"`
}

type NodeInfo struct {
	Network string `json:"network"`
}

type DipTransferParams struct {
	ReqBaseParams
	FromAddress string `json:"from_address"` //发送地址
	ToAddress   string `json:"to_address"`   //接收地址
	Amount      string `json:"amount"`       //接收金额
	Denom       string `json:"denom"`        //币种标示（主要用于区分主链和代币的转账）		dip的是【pdip】
	Gas         uint64 `json:"gas"`
	Fee         int64  `json:"fee"`
	Memo        string `json:"memo"` //memo

}

type DipSignParams struct {
	ReqBaseParams
	FromAddress string          `json:"fromaddress"` //发送地址
	ToAddress   string          `json:"toaddress"`   //接收地址
	ToAmount    decimal.Decimal `json:"toamount"`    //接收金额
	Memo        string          `json:"memo"`        //memo
	ChainID     string          `json:"chain_id"`
}

type (
	AccountValue struct {
		Address       string       `json:"address"`
		Coins         []types.Coin `json:"coins"`
		AccountNumber string       `json:"account_number"`
		Sequence      string       `json:"sequence"`
	}

	AccountResult struct {
		Type  string       `json:"type"`
		Value AccountValue `json:"value"`
	}

	DipAccountBody struct {
		Height string        `json:"height"`
		Result AccountResult `json:"result"`
	}
)
