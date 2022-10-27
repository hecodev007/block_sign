package entity

import (
	"time"
)

type FcRecHoo struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	Pid        int       `json:"pid" xorm:"not null default 0 comment('info_id') INT(11)"`
	Code       string    `json:"code" xorm:"not null VARCHAR(30)"`
	Type       string    `json:"type" xorm:"not null comment('balance:余额，plus 币加，gift 礼品卡，ram 内存') VARCHAR(15)"`
	CoinName   string    `json:"coin_name" xorm:"not null comment('币种英文名称') VARCHAR(11)"`
	Address    string    `json:"address" xorm:"not null VARCHAR(80)"`
	Amount     string    `json:"amount" xorm:"not null default 0.000000000000000 comment('数量') DECIMAL(40,15)"`
	Freeze     string    `json:"freeze" xorm:"not null default 0.000000000000000 comment('冻结') DECIMAL(40,15)"`
	ConTime    time.Time `json:"con_time" xorm:"not null comment('对帐时间') DATE"`
	Addtime    int       `json:"addtime" xorm:"not null comment('添加时间') INT(11)"`
	RepeatType string    `json:"repeat_type" xorm:"comment('balance:余额，plus 币加，gift 礼品卡，ram 内存') VARCHAR(15)"`
	Page       int       `json:"page" xorm:"INT(11)"`
}
