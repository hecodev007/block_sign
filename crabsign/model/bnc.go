package model

type BncTransferParams struct {
	ReqBaseParams
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
	//cold sign need
	Nonce              uint64 `json:"nonce"`
	SpecVersion        uint32 `json:"spec_version"`
	TransactionVersion uint32 `json:"transaction_version"`
}
