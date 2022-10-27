package domain

import (
	"github.com/shopspring/decimal"
	"time"
)

type MqWalletInfo struct {
	FromAddress string          `json:"from_address"`
	ToAddress   string          `json:"to_address"`
	TxId        string          `json:"tx_id"`
	SerialNo    string          `json:"serial_no"`
	CoinName    string          `json:"coin_name"`
	Memo        string          `json:"memo"`
	Nums        decimal.Decimal `json:"nums"`
	Height      int             `json:"height"`
	ConfirmNums int             `json:"confirm_nums"`
	ConfirmTime time.Time       `json:"confirm_time"`
	RealNums    decimal.Decimal `json:"real_nums"`
	Destroy     decimal.Decimal `json:"destroy"`
	BurnFee     decimal.Decimal `json:"burn_fee"`
	MinerFee    decimal.Decimal `json:"miner_fee"`
}
