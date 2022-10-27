package bo

//生成地址数量
//{"num":10,"orderId":"123456","mchId":"hoo","coinName":"bch"}
type CreateAddrParam struct {
	Count int    `json:"count"`
	Mch   string `json:"mch"`
}
