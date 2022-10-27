package model

type Result struct{
	Code int `json:"code"`
	Msg string `json:"message"`
}

type TransferParams struct{
	From string `json:"from"`
	Amount string `json:"amount"`
}

