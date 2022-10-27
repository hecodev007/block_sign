package entity

import (
	"time"
)

type FcRecHooData struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	Type       string    `json:"type" xorm:"unique(type) ENUM('balance','gift','plus')"`
	CoinName   string    `json:"coin_name" xorm:"not null unique(type) VARCHAR(15)"`
	Amount     string    `json:"amount" xorm:"not null DECIMAL(40,15)"`
	OtherNum   string    `json:"other_num" xorm:"not null DECIMAL(40,15)"`
	DDate      time.Time `json:"d_date" xorm:"not null default '0000-00-00 00:00:00' comment('统计时间') unique(type) DATETIME"`
	Createtime time.Time `json:"createtime" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP"`
}
