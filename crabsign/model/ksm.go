package model

type KsmTransferParams struct {
	ReqBaseParams
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
	Nonce       uint64 `json:"nonce"` //cold sign need
}

type KsmReqNodeParams struct {
	FromAddress string `json:"fromAddr"`
	ToAddress   string `json:"toAddr"`
	Amount      string `json:"toAmount"`
	FromSeed    string `json:"fromSeed"`
}

type KsmRespNodeParams struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

type KsmColdParams struct {
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
