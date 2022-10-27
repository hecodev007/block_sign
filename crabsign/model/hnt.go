package model

type HntTransferParams struct {
	ReqBaseParams
	FromAddress string `json:"from_address"`
	ToAddress   string `json:"to_address"`
	Amount      string `json:"amount"`
	//FeeInt		string	`json:"fee_int"`
	Nonce int64 `json:"nonce"` //热钱包不需要传这个值
}
