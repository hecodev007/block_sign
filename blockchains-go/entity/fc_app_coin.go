package entity

import (
	"time"
)

type FcAppCoin struct {
	Id           int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	AppAddressId int       `json:"app_address_id" xorm:"unique INT(11)"`
	Num          string    `json:"num" xorm:"not null default 0.0000000000 comment('当前余额') DECIMAL(23,10)"`
	Updatetime   time.Time `json:"updatetime" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
}
