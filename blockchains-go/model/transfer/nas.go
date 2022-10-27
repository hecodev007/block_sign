package transfer

type NasOrderRequest struct {
	OrderRequestHead
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"` // '接收者地址'
	Amount      string `json:"amount"`
	Token       string `json:"token,omitempty"`
	Fee         int64  `json:"fee,omitempty"`
}
