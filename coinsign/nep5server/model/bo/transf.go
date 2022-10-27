package bo

import "github.com/shopspring/decimal"

type TransfBO struct {
	From           string          `json:"from" binding:"required"`
	To             string          `json:"to" binding:"required"`
	ScriptAddr     string          `json:"scriptAddr" binding:"required"`
	AmountInt      int64           `json:"amountInt" binding:"required,min=1"`
	AmountFloatStr decimal.Decimal `json:"amountFloatStr" binding:"required"`
}
