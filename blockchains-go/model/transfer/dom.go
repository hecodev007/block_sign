package transfer

type DomOrderRequest struct {
	MchName  string          `json:"mchId"`
	OrderID  string          `json:"orderId"`
	CoinName string          `json:"coinName"`
	Data     DomTransferData `json:"data"`
}

type DomTransferData struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
}
