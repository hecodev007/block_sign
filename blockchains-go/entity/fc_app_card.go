package entity

type FcAppCard struct {
	Id        int    `json:"id" xorm:"not null pk autoincr INT(11)"`
	AppId     int    `json:"app_id" xorm:"unique(app_id) INT(11)"`
	CoinId    int    `json:"coin_id" xorm:"unique(app_id) INT(11)"`
	Num       string `json:"num" xorm:"not null default 0.0000000000 comment('总金额') DECIMAL(23,10)"`
	ObtainNum string `json:"obtain_num" xorm:"not null default 0.0000000000 comment('领取金额') DECIMAL(23,10)"`
	OutNum    string `json:"out_num" xorm:"not null default 0.0000000000 comment('退回金额') DECIMAL(23,10)"`
	Total     int    `json:"total" xorm:"not null default 0 comment('总数量') INT(11)"`
	Obtain    int    `json:"obtain" xorm:"not null default 0 comment('领取数量') INT(11)"`
	Out       int    `json:"out" xorm:"not null default 0 comment('退回数量') INT(11)"`
}
