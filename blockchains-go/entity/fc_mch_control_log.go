package entity

type FcMchControlLog struct {
	Id         int    `json:"id" xorm:"not null pk autoincr INT(11)"`
	AppId      int    `json:"app_id" xorm:"not null default 0 comment('商户id') INT(11)"`
	CoinId     int    `json:"coin_id" xorm:"not null default 0 comment('币种id') INT(10)"`
	Level      int    `json:"level" xorm:"not null default 0 comment('风控规则  1 每小时限额  2 每日限额  3 单笔限额') TINYINT(3)"`
	PendTime   int    `json:"pend_time" xorm:"not null default 0 comment('触发时间') INT(10)"`
	UnlockTime int    `json:"unlock_time" xorm:"not null default 0 comment('解锁时间') INT(10)"`
	Operator   string `json:"operator" xorm:"comment('操作人') VARCHAR(100)"`
}
