package transfer

type Wd_wdOrderRequest struct {
	OrderRequestHead
	FromAddress string `json:"from_addr"`
	ToAddress   string `json:"to_addr"` // '接收者地址'
	Amount      string `json:"amount"`
	GasLimit    int64  `json:"gas_limit"`
	GasPremium  int64  `json:"gas_premium"`
	GasFeeCap   int64  `json:"gas_fee_cap"`
	Nonce       int64  `json:"nonce"`
}
