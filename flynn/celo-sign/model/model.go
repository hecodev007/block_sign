package model

//========================create address==================//
type ReqBaseParams struct {
	OrderId  string `json:"orderId,omitempty"`
	MchId    string `json:"mchId,omitempty"`
	CoinName string `json:"coinName,omitempty"`
}

type ReqCreateAddressParams struct {
	Num int `json:"num"`
	ReqBaseParams
}

type RespCreateAddressParams struct {
	ReqCreateAddressParams
	Address []string `json:"address"`
}

//==========================================================//

type ReqSignParams struct {
	ReqBaseParams
	Data interface{} `json:"data"`
}

type RespSignParams struct {
	ReqBaseParams
	Result string `json:"result"`
}

//=============================================================//
