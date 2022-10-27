package entity

type FcMchMoney struct {
	Id           int    `json:"id" xorm:"not null pk autoincr INT(11)"`
	AppId        int    `json:"app_id" xorm:"not null default 0 comment('商户id') INT(10)"`
	Amount       string `json:"amount" xorm:"not null default 0.000000000000000000 comment('可用余额') DECIMAL(40,18)"`
	AmountFreeze string `json:"amount_freeze" xorm:"not null default 0.000000000000000000 comment('冻结金额') DECIMAL(40,18)"`
	Address      string `json:"address" xorm:"comment('商户余额地址') VARCHAR(200)"`
	Status       int    `json:"status" xorm:"default 1 comment('0冻结
1正常') TINYINT(4)"`
}
