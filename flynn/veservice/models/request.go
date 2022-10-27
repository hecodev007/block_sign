package models

type Object struct {
	Id      string
	Jsonrpc string
	Method  string
	Params  []interface{}
}

type SignTawTx struct {
	RawTx    string
	Inputs   string
	PrivKeys string
}

type EthRawTx struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Nonce       uint64 `json:"nonce"`
	Amount      int64  `json:"amount"`
	GasLimit    uint64 `json:"gasLimit"`
	GasPrice    int64  `json:"gasPrice"`
	Data        string `json:"data"`
	BlockNumber int64  `json:"blockNumber"`
}

type EthRawData struct {
	Index int    `json:"index"`
	Hex   string `json:"hex"`
}

type Erc20RawTx struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Nonce       uint64 `json:"nonce"`
	Amount      int64  `json:"amount"`
	GasLimit    uint64 `json:"gasLimit"`
	GasPrice    int64  `json:"gasPrice"`
	Data        string `json:"data"`
	BlockNumber int64  `json:"blockNumber"`
}

type HCUtxoParams struct {
	Txid    string  `json:"txid"`
	Voutn   uint32  `json:"vout"`
	Amount  float64 `json:"amount"`
	Tree    int8    `json:"tree"`
	Address string  `json:"address"`
}
