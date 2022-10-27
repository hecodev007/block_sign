package model

//业务标识

type MchInfo struct {
	MchId    string `json:"mchId,omitempty"`
	OrderId  string `json:"orderId,omitempty"`
	CoinName string `json:"coinName,omitempty"`
}

type MchInfoReq struct {
	MchId        string `json:"mch_id,omitempty"`
	OrderId      string `json:"order_no,omitempty"`
	OuterOrderNo string `json:"outer_order_no,omitempty"`
	CoinName     string `json:"coin_name,omitempty"`
}
