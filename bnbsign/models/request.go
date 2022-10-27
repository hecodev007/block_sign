package models

type Object struct {
	Id      string
	Jsonrpc string
	Method  string
	Params  []interface{}
}

type GetNewAddress struct {
	Id     string `json:"id"`
	Result string `json:"result"`
	Error  string `json:"error"`
}

type csvAddress struct {
	EncryptWif string `csv:"wif"`
	Address    string `csv:"address"`
}

type csvAddress2 struct {
	WifKey  string `csv:"wifkey"`
	Address string `csv:"address"`
}

type csvAddress3 struct {
	Address string `csv:"address"`
}

type HCUtxoParams struct {
	Txid    string  `json:"txid"`
	Voutn   uint32  `json:"vout"`
	Amount  float64 `json:"amount"`
	Tree    int8    `json:"tree"`
	Address string  `json:"address"`
}

type HcVout struct {
	Address string  `json:"address"`
	Amount  float64 `json:"amount"`
}