package entity

import (
	"time"
)

type FcAppCoinVer struct {
	Id       int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	AppName  string    `json:"app_name" xorm:"VARCHAR(15)"`
	CoinName string    `json:"coin_name" xorm:"VARCHAR(15)"`
	Address  string    `json:"address" xorm:"VARCHAR(255)"`
	Num1     string    `json:"num1" xorm:"not null default 0.0000000000 comment('FNS余额') DECIMAL(23,10)"`
	Num2     string    `json:"num2" xorm:"not null default 0.0000000000 comment('平台余额') DECIMAL(23,10)"`
	VerTime  time.Time `json:"ver_time" xorm:"DATETIME"`
}
