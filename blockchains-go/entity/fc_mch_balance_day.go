package entity

type FcMchBalanceDay struct {
	Id        int    `json:"id" xorm:"not null pk autoincr comment('id') INT(11)"`
	AppId     int    `json:"app_id" xorm:"not null default 0 comment('商户id') INT(10)"`
	Balance   string `json:"balance" xorm:"not null default 0.000000000000000000 comment('余额') DECIMAL(40,18)"`
	OutAmount string `json:"out_amount" xorm:"not null default 0.000000000000000000 comment('支出金额') DECIMAL(40,18)"`
	Date      string `json:"date" xorm:"comment('日期') VARCHAR(12)"`
	AddTime   int    `json:"add_time" xorm:"not null default 0 comment('加入时间') INT(11)"`
}
