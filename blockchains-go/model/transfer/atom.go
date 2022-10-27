package transfer

type AtomPaymentRequest struct {
	FromAddress string `json:"from_addr"`
	ToAddress   string `json:"to_addr"` // '接收者地址'
	AmountInt64 string `json:"amount"`
	//Fee         int64  `json:"fee,omitempty"` //allow max fee
	Memo    string `json:"memo,omitempty"`
	ChainID string `json:"chain_id,omitempty"`
}
type AtomOrderRequest struct {
	OrderRequestHead
	AtomPaymentRequest
}
