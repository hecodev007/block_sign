package bo

type CreateAddrIn struct {
	Num      uint   `json:"num"`      //数量
	OrderId  string `json:"orderId"`  //订单ID
	Mch      string `json:"mch"`      //商家名
	Coinname string `json:"coinname"` //币种
}

type CreateAddrResult struct {
	Num      uint     `json:"num"`                //数量
	OrderId  string   `json:"orderId"`            //订单ID
	Mch      string   `json:"mch"`                //商家名
	Coinname string   `json:"coinname,omitempty"` //币种
	Addrs    []string `json:"addrs"`
}
