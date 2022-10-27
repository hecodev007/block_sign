package transfer

// Waxp订单请求
type WaxpOrderRequest struct {
	ApplyId      int64  `json:"apply_id,omitempty"`
	ApplyCoinId  int64  `json:"apply_coin_id,omitempty"`
	OuterOrderNo string `json:"outer_order_no,omitempty"`
	OrderNo      string `json:"order_no,omitempty"`
	MchName      string `json:"mch_name,omitempty"`
	CoinName     string `json:"coin_name,omitempty"`
	FromAddress  string `json:"from_address"`    //发送地址
	ToAddress    string `json:"to_address"`      //接收地址
	Token        string `json:"token,omitempty"` //code
	Decimal      int    `json:"decimal,omitempty"`
	Memo         string `json:"memo,omitempty"`
	Quantity     string `json:"quantity,omitempty"`
	SignPubkey   string `json:"sign_pubkey,omitempty"`
}

// Wax订单请求
type WaxpHotOrderRequest struct {
	OrderRequestHead
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"` // '接收者地址'
	Quantity    string `json:"quantity"`
	Token       string `json:"token"`
	Memo        string `json:"memo"`
	SignPubkey  string `json:"sign_pubkey"`
	Amount      int64  `json:"amount"`
}
