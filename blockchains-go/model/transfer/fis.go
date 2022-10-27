package transfer

type FisOrderRequest struct {
	OrderRequestHead
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"` // '接收者地址'
	Amount      string `json:"amount"`
}
