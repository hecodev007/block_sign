package model

type NearTransferParams struct {
	ReqBaseParams
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
}
