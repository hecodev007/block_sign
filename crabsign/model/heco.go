package model

type HecoTransferParams struct {
	ReqBaseParams
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Amount          string `json:"amount"`
	IsCollect       int    `json:"is_collect"`
	Token           string `json:"token"`
	ContractAddress string `json:"contract_address"`
}
