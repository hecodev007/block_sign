package vo

type CreateAddrVO struct {
	OrderId  string   `json:"orderId"`  //订单ID
	MchId    string   `json:"mchId" `   //商家名
	CoinName string   `json:"coinName"` //币种
	Num      uint     `json:"num"`      //数量
	Addrs    []string `json:"addrs"`    //addrs
}
