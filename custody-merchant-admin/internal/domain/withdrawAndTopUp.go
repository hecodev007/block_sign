package domain

import "github.com/shopspring/decimal"

type WithdrawInfo struct {
	BusinessId  int             `json:"business_id"`
	FromAddress string          `json:"from_address"`
	ToAddress   string          `json:"to_address"`
	Sign        string          `json:"sign"`
	Coin        string          `json:"coin"`
	Chain       string          `json:"chain"`
	ClientId    string          `json:"client_id"`
	Memo        string          `json:"memo"`
	Amount      decimal.Decimal `json:"amount"`
	Nonce       string          `json:"nonce"`
	Ts          int64           `json:"ts"`
}

type WithdrawStruct struct {
	FromAddress string  `json:"from_address" form:"from_address"`
	ToAddress   string  `json:"to_address" form:"to_address"`
	Sign        string  `json:"sign" form:"sign"`
	Coin        string  `json:"coin" form:"coin"`
	Chain       string  `json:"chain" form:"chain"`
	ClientId    string  `json:"client_id" form:"client_id"`
	Memo        string  `json:"memo" form:"memo"`
	Amount      string  `json:"amount" form:"amount"`
	Nonce       string  `json:"nonce" form:"nonce"`
	Ts          float64 `json:"ts" form:"ts"`
}

type WithdrawParams struct {
	FromAddress string          `json:"from_address"`
	ToAddress   string          `json:"to_address"`
	Sign        string          `json:"sign"`
	Coin        string          `json:"coin"`
	Chain       string          `json:"chain"`
	ClientId    string          `json:"client_id"`
	Memo        string          `json:"memo"`
	Amount      decimal.Decimal `json:"amount"`
	Nonce       string          `json:"nonce"`
	Ts          int64           `json:"ts"`
}

func StructToInfo(w *WithdrawStruct) *WithdrawInfo {
	fromString, err := decimal.NewFromString(w.Amount)
	if err != nil {
		return nil
	}
	return &WithdrawInfo{
		FromAddress: w.FromAddress,
		ToAddress:   w.ToAddress,
		Sign:        w.Sign,
		Coin:        w.Coin,
		Chain:       w.Chain,
		ClientId:    w.ClientId,
		Memo:        w.Memo,
		Nonce:       w.Nonce,
		Ts:          int64(w.Ts),
		Amount:      fromString,
	}
}

func ParamsToWithdraw(w *WithdrawStruct) *WithdrawParams {
	fromString, err := decimal.NewFromString(w.Amount)
	if err != nil {
		return nil
	}
	return &WithdrawParams{
		FromAddress: w.FromAddress,
		ToAddress:   w.ToAddress,
		Sign:        w.Sign,
		Coin:        w.Coin,
		Chain:       w.Chain,
		ClientId:    w.ClientId,
		Memo:        w.Memo,
		Nonce:       w.Nonce,
		Ts:          int64(w.Ts),
		Amount:      fromString,
	}
}
