package model

type EthTransferParams struct {
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Amount          string `json:"amount"`
	IsCollect       int    `json:"is_collect"`
	Token           string `json:"token"`
	ContractAddress string `json:"contract_address"`
	Fee             int64  `json:"fee"`

	// 获取Nonce是否使用latest，默认使用pending
	Latest bool `json:"latest"`
}

type EthTransferWithNonceParams struct {
	EthTransferParams

	// 指定nonce值
	Nonce uint64 `json:"nonce"`
}

type EthSignParams struct {
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Amount          string `json:"amount"`
	ContractAddress string `json:"contract_address"`
	Nonce           uint64 `json:"nonce"`
	GasPrice        int64  `json:"gas_price"`
	GasLimit        int64  `json:"gas_limit"`
}

type EthNonceData struct {
	Txid  string `json:"txid"`
	Nonce int64  `json:"nonce"`
}
