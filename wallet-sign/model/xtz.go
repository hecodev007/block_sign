package model

type XtzTransferParams struct {
	ReqBaseParams
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
	GasLimit    string `json:"Gas_limit"`
	Fee         string `json:"fee"`
}
