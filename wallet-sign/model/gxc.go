package model

type GxcSignParams struct {
	PublicKey string `json:"publicKey"`
	StxHex    string `json:"stxHex"`
	ChainId   string `json:"chainId"`
}
type RespGxcSignParams struct {
	ReqBaseParams
	Hex string `json:"hex"`
}

type GxcTransferParams struct {
	FromAccount string `json:"fromAccount"`
	ToAccount   string `json:"toAccount"`
	Memo        string `json:"memo"`
	Amount      string `json:"amount"`
	Fee         string `json:"fee"`
	PublicKey   string `json:"publicKey"`
}

type RespGxcTransferParams struct {
	Txid string `json:"txid"`
	Fee  string `json:"fee"`
	Memo string `json:"memo"`
}
