package model

type TxInfo struct {
	CoinType     string `json:"coin_type"`
	OuterOrderNo string `json:"outer_order_no"`
	OrderNo      string `json:"order_no"`
	BlockHeight  int    `json:"block_height"`
	Timestamp    int    `json:"timestamp"`
	TxId         string `json:"tx_id"`
	TxType       int    `json:"tx_type"`
	FromAddress  string `json:"from_address"`
	ToAddress    string `json:"to_address"`
	Memo         string `json:"memo"`
	Amount       string `json:"amount"`
	TxFee        string `json:"tx_fee"`
	TxFeeCoin    string `json:"tx_fee_coin"`
	ContrastTime int    `json:"contrast_time"`
}
