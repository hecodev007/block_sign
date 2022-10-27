package model

type TransferResponse struct {
	Status  int64  `json:"status"`
	Message string `json:"msg"`
	Success bool   `json:"success"`
}
