package model

type CringTransferParams struct {
	ReqBaseParams
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
	//cold sign need
	Nonce              uint64 `json:"nonce"`
	SpecVersion        uint32 `json:"spec_version"`
	TransactionVersion uint32 `json:"transaction_version"`
}

type CringColdParams struct {
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
type AccountInfo struct {
	Nonce    uint64 `json:"nonce"`
	RefCount uint8  `json:"ref_count"`
	Data     struct {
		Free       uint64 `json:"free"`
		Reserved   uint64 `json:"reserved"`
		MiscFrozen uint64 `json:"misc_frozen"`
		FreeFrozen uint64 `json:"free_frozen"`
	} `json:"data"`
}
