package bo

//生成地址数量
//{"num":10,"orderId":"123456","mchId":"hoo","coinName":"bch"}
type CreateAddrParam struct {
	Num      int    `json:"num"`
	OrderId  string `json:"orderId"`
	MchId    string `json:"mchId"`
	CoinName string `json:"coinName"`
}
