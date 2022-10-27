package model

type CoinList struct {
	Father  string `json:"father"`
	Name    string `json:"name"`
	Token   string `json:"token"`
	Decimal int    `json:"decimal"`
}

//CustodyCoinList 托管后台 币参数结构
type CustodyCoinList struct {
	Father   string `json:"father"`
	Name     string `json:"name"`
	Token    string `json:"token"`
	Decimal  int    `json:"decimal"`
	State    int    `json:"state"`
	Confirm    int           `json:"confirm" form:"confirm"`
	FullName string `json:"full_name"`
	PriceUsd string `json:"price_usd"`
}
