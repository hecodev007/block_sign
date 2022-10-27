package transfer

type AtpOrderRequest struct {
	OrderRequestHead
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"` // '接收者地址'
	Value       string `json:"value"`
	Nonce       int64  `json:"nonce"`
	GasPrice    int64  `json:"gas_price"`
	GasLimit    int64  `json:"gas_limit"`
}
