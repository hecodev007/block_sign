package transfer

type LuncPaymentRequest struct {
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"` // '接收者地址'
	Amount      int64  `json:"amount"`
	Fee         int64  `json:"fee,omitempty"` //allow max fee
	Memo        string `json:"memo,omitempty"`
	Token       string `json:"token"`
}
type LuncOrderRequest struct {
	OrderRequestHead
	Data LuncPaymentRequest `json:"data"`
}
