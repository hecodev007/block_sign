package model

type ARSignParams struct {
	LastTx      string `json:"last_tx"`
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Fee         string `json:"fee"`
	Amount      string `json:"amount"`
	//Memo 				string			`json:"memo"`
}

type RespArSignParams struct {
	ReqBaseParams
	Data string `json:"data"`
}

type ARTransferParams struct {
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
	//Memo 			string				`json:"memo"`
}
