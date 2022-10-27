package transfer

type XlmOrderReq struct {
	OrderRequestHead
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Memo        string `json:"memo"`
	Amount      int64  `json:"amount"`
	Token       string `json:"token"`
	IsRetry     bool   `json:"is_retry"`
}

type XlmOrderHotRequest struct {
	OrderRequestHead
	FromAddress string `json:"from"`
	ToAddress   string `json:"to"`
	Amount      string `json:"value"`
	Memo        string `json:"memo"`
	Fee         string `json:"fee"`
	Token       string `json:"token"`
}
