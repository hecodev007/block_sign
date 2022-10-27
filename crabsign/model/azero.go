package model

type AzeroTransferParams struct {
	ReqBaseParams
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
	//cold sign need
	Nonce              uint64 `json:"nonce"`
	SpecVersion        uint32 `json:"spec_version"`
	TransactionVersion uint32 `json:"transaction_version"`
}

type AzeroColdParams struct {
	ReqBaseParams
	FromAddress        string `json:"from_address"`
	ToAddress          string `json:"to_address"`
	Amount             string `json:"amount"`
	Nonce              uint64 `json:"nonce"`
	SpecVersion        uint32 `json:"spec_version"`
	TransactionVersion uint32 `json:"transaction_version"`
	GenesisHash        string `json:"genesis_hash"`
	BlockHash          string `json:"block_hash"`
	BlockNumber        uint64 `json:"block_number"`
}
