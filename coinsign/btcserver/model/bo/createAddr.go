package bo

//{"num":10,"orderId":"123456","mchId":"hoo","coinName":"btc"}

type CreateAddrParam struct {
	Num      int    `json:"num"`
	OrderId  string `json:"orderId"`
	MchId    string `json:"mchId"`
	CoinName string `json:"coinName"`
}
