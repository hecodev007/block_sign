package model

type CdsTransferParams struct {
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
	IsCollect   int    `json:"is_collect"`
	//Memo 			string				`json:"memo"`
}
