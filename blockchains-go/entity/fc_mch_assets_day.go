package entity

type FcMchAssetsDay struct {
	Id       int    `json:"id" xorm:"not null pk autoincr INT(11)"`
	AppId    int    `json:"app_id" xorm:"not null default 0 comment('商户id') unique(mch_coin_date) INT(10)"`
	CoinId   int    `json:"coin_id" xorm:"not null default 0 comment('币种id') INT(10)"`
	CoinName string `json:"coin_name" xorm:"not null comment('币种名称') unique(mch_coin_date) VARCHAR(12)"`
	In       string `json:"in" xorm:"not null default 0.000000000000000000 comment('日收入') DECIMAL(40,18)"`
	Out      string `json:"out" xorm:"not null default 0.000000000000000000 comment('日支出') DECIMAL(40,18)"`
	InNum    int    `json:"in_num" xorm:"not null default 0 comment('收入笔数') INT(10)"`
	OutNum   int    `json:"out_num" xorm:"not null default 0 comment('支出笔数') INT(10)"`
	Date     string `json:"date" xorm:"comment('日期') unique(mch_coin_date) VARCHAR(20)"`
}
