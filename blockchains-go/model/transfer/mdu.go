package transfer

import "github.com/shopspring/decimal"

// mdu订单请求
type MduOrderRequest struct {
	ApplyId       int64           `json:"apply_id,omitempty"`
	ApplyCoinId   int64           `json:"apply_coin_id,omitempty"`
	OuterOrderNo  string          `json:"outer_order_no,omitempty"`
	OrderNo       string          `json:"order_no,omitempty"`
	MchName       string          `json:"mch_name,omitempty"`
	CoinName      string          `json:"coin_name,omitempty"`
	FromAddress   string          `json:"from_address"` //发送地址
	ToAddress     string          `json:"to_address"`   //接收地址
	ToAmountFloat decimal.Decimal `json:"amount"`       //接收金额
	Token         string          `json:"token,omitempty"`
	Decimal       int             `json:"decimal,omitempty"`
	Memo          string          `json:"memo,omitempty"`
}
