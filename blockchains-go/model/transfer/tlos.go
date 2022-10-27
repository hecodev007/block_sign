package transfer

type TlosOrderRequest struct {
	OrderRequestHead
	Data *TlosOrderData `json:"data"`
}
type TlosOrderData struct {
	FromAddress string `json:"from_address,omitempty"`
	ToAddress   string `json:"to_address,omitempty"`
	Token       string `json:"token,omitempty"`
	Quantity    string `json:"quantity,omitempty"`
	Memo        string `json:"memo,omitempty"`
	SignPubKey  string `json:"sign_pubkey,omitempty"`
	BlockId     string `json:"block_id,omitempty"`
	Amount      int64  `json:"amount,omitempty"`
}
