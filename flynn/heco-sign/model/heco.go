package model

type HecoTransferParams struct {
	ReqBaseParams
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Amount          string `json:"amount"`
	IsCollect       int    `json:"is_collect"`
	ContractAddress string `json:"contract_address"`
	Fee             int64  `json:"fee"`
	GasPrice        int64  `json:"gasPrice"`
	GasLimit        int64  `json:"gasLimit"`
}

type HecoSignParams struct {
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Amount          string `json:"amount"`
	ContractAddress string `json:"contract_address"`
	Nonce           int64  `json:"nonce"`
	GasPrice        int64  `json:"gas_price"`
	GasLimit        int64  `json:"gas_limit"`
}

type HecoNonceData struct {
	Txid  string `json:"txid"`
	Nonce int64  `json:"nonce"`
}
