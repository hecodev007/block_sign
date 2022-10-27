package transfer

type IostOrderRequest struct {
	OrderNo  string `json:"order_no,omitempty"`
	MchName  string `json:"mch_name,omitempty"`
	CoinName string `json:"coin_name,omitempty"`

	FromAccount string `json:"from_account"`
	FromAddress string `json:"from_address"`
	ToAccount   string `json:"to_account"` // '接收者地址'
	Value       string `json:"value"`
	Memo        string `json:"memo"`
	Fee         string `json:"fee"`
	Token       string `json:"token"`
}
