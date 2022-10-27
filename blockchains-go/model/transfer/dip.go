package transfer

type DipPaymentRequest struct {
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"` // '接收者地址'
	AmountInt64 string `json:"amount"`
	Fee         int64  `json:"fee,omitempty"` //allow max fee
	Memo        string `json:"memo,omitempty"`
	Gas         int64  `json:"gas"`
	Denom       string `json:"denom"` //币种标示（主要用于区分主链和代币的转账）		dip的是【pdip

}
type DipOrderRequest struct {
	OrderRequestHead
	DipPaymentRequest
}
