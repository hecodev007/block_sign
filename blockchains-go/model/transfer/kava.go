package transfer

type KavaPaymentRequest struct {
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"` // '接收者地址'
	Amount      int64  `json:"amount"`
	Fee         int64  `json:"fee,omitempty"` //allow max fee
	Memo        string `json:"memo,omitempty"`
}
type KavaOrderRequest struct {
	OrderRequestHead
	Data KavaPaymentRequest `json:"data"`
}
