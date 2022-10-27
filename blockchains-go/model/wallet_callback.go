package model

type WalltCallBack struct {
	ApplyId      int64  `json:"applyId"`
	OrderNo      string `json:"orderNo"`
	OuterOrderNo string `json:"outerOrderNo"`
	MchName      string `json:"mchName"`
	Status       int    `json:"status"`
	Message      string `json:"message"`
	Worker       string `json:"worker"`
}
