package entity

type FcMchControl struct {
	Id        int     `json:"id" xorm:"not null pk autoincr INT(11)"`
	AppId     int     `json:"app_id" xorm:"not null default 0 comment('商户id') INT(11)"`
	CoinId    int     `json:"coin_id" xorm:"not null default 0 comment('币种id') INT(10)"`
	Status    int     `json:"status" xorm:"not null default 0 comment('状态 0 正常  1 冻结') TINYINT(3)"`
	HourLimit string  `json:"hour_limit" xorm:"not null default 0.000000000000000000 comment('每小时限额') DECIMAL(40,18)"`
	Single    string  `json:"single" xorm:"not null default 0.000000000000000000 comment('单笔限额') DECIMAL(40,18)"`
	DayLimit  string  `json:"day_limit" xorm:"not null default 0.000000000000000000 comment('每日限额') DECIMAL(40,18)"`
	Locklen   float32 `json:"locklen" xorm:"not null default 0.00 comment('冻结时长 以小时为单位') FLOAT(10,2)"`
	Locktime  int     `json:"locktime" xorm:"not null default 0 comment('冻结时间') INT(11)"`
}
