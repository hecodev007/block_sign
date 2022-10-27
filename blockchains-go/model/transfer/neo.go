package transfer

type NeoOrderRequest struct {
	OrderRequestHead
	//FromAddress string `json:"from_address"`
	//ToAddress   string `json:"to_address"` // '接收者地址'
	//Amount      string `json:"amount"`
	TxIns  []NeoTxIn  `json:"tx_ins"`
	TxOuts []NeoTxOut `json:"tx_outs"`
}

type NeoTxIn struct {
	FromAddr   string `json:"from_addr"`
	FromTxid   string `json:"from_txid"`
	FromIndex  int    `json:"from_index"`
	FromAmount int64  `json:"from_amount"`
}

type NeoTxOut struct {
	ToAddr   string `json:"to_addr"`
	ToAmount int64  `json:"to_amount"`
}

type NeoUtxo struct {
	Address string       `json:"address"`
	Balance []NeoBalance `json:"balance"`
}

type NeoBalance struct {
	Amount      string       `json:"amount"`
	Asset       string       `json:"asset"`
	AssetHash   string       `json:"asset_hash"`
	AssetSymbol string       `json:"asset_symbol"`
	Unspent     []NeoUnspent `json:"unspent"`
}
type NeoUnspent struct {
	Value string `json:"value"`
	Txid  string `json:"txid"`
	N     int    `json:"n"`
}
