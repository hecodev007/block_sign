package model

//业务标识

type MchInfo struct {
	MchId    string `json:"mchId,omitempty"`
	OrderId  string `json:"orderId,omitempty"`
	CoinName string `json:"coinName,omitempty"`
}
