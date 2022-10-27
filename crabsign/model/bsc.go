package model

type BscTransferParams struct {
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Amount          string `json:"amount"`
	IsCollect       int    `json:"is_collect"`
	Token           string `json:"token"`
	ContractAddress string `json:"contract_address"`
}

type BscSignParams struct {
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Amount          string `json:"amount"`
	ContractAddress string `json:"contract_address"`
	Nonce           int64  `json:"nonce"`
}

type BscNonceData struct {
	Txid  string `json:"txid"`
	Nonce int64  `json:"nonce"`
}
