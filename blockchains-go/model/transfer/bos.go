package transfer

type BosOrderRequest struct {
	OrderRequestHead
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"` // '接收者地址'
	Quantity    string `json:"quantity"`
	Token       string `json:"token"`
	Memo        string `json:"memo"`
	SignPubkey  string `json:"sign_pubkey"`
	Amount      int64  `json:"amount"`
}
