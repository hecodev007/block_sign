package entity

import (
	"time"
)

type FcYouzi struct {
	Id       int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	Username string    `json:"username" xorm:"unique(username) VARCHAR(80)"`
	Amount   string    `json:"amount" xorm:"not null DECIMAL(40,15)"`
	TDate    time.Time `json:"t_date" xorm:"not null DATE"`
	Addtime  int       `json:"addtime" xorm:"not null INT(11)"`
	Type     int       `json:"type" xorm:"not null default 0 comment('0柚子理财1pos矿池') unique(username) TINYINT(4)"`
	DoDate   time.Time `json:"do_date" xorm:"not null default '0000-00-00 00:00:00' unique(username) DATETIME"`
}
