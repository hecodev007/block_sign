package transfer

import "github.com/shopspring/decimal"

type ZvcOrderRequest struct {
	OrderRequestHead
	CoinName    string          `json:"coin_name"`    //币种名称
	FromAddress string          `json:"from_address"` //发送地址
	ToAddress   string          `json:"to_address"`   //接收地址
	ToAmount    decimal.Decimal `json:"amount"`       //接收金额
	Memo        string          `json:"memo"`         //memo
}
