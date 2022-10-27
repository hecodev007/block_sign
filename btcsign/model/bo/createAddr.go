package bo

//{"num":10,"orderId":"123456","mchId":"hoo","coinName":"btc"}

type CreateAddrParam struct {
	Count int    `json:"count"`
	Mch   string `json:"mch"`
}
