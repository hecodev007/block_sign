package transfer

type XdagOrderRequest struct {
	OrderNo  string `json:"order_no,omitempty"`
	MchName  string `json:"mch_name,omitempty"`
	CoinName string `json:"coin_name,omitempty"`

	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"` // '接收者地址'
	Value       string `json:"value"`
	Amount      string `json:"amount"`
	Memo        string `json:"memo"`
	Fee         string `json:"fee"`
	Token       string `json:"token"`
}
