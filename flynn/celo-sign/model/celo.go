package model

type CeloTransferParams struct {
	FromAddress     string `json:"from_address"`
	ToAddress       string `json:"to_address"`
	Amount          string `json:"amount"`
	ContractAddress string `json:"contract_address"`
	IsCollect       int    `json:"is_collect"`
	//Memo 			string				`json:"memo"`
}
