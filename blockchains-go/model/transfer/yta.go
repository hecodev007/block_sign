package transfer

type YtaOrderRequest struct {
	OrderRequestHead
	Data YtaOrderData `json:"data"`
}
type YtaOrderData struct {
	SignPubkey  string `json:"sign_pubkey"`
	Token       string `json:"token"`
	FromAddress string `json:"from_address"`
	Memo        string `json:"memo"`
	Quantity    string `json:"quantity"`
	ToAddress   string `json:"to_address"`
	BlockId     string `json:"block_id"`
}
