package transfer

type KlayPaymentRequest struct {
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"` // '接收者地址'
	Amount      string `json:"amount"`
	Fee         string `json:"fee,omitempty"` //allow max fee
	Token       string `json:"token,omitempty"`
	FeePayer    string `json:"fee_payer"`
}
type KlayOrderRequest struct {
	OrderRequestHead
	Data KlayPaymentRequest `json:"data"`
}
