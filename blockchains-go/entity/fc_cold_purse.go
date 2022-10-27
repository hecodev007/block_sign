package entity

import (
	"time"
)

type FcColdPurse struct {
	Id           int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	AppId        int       `json:"app_id" xorm:"not null default 0 unique(appid_coinid) INT(11)"`
	CoinId       int       `json:"coin_id" xorm:"not null default 0 unique(appid_coinid) INT(11)"`
	Num          string    `json:"num" xorm:"not null default 0.0000000000 comment('分配余额') DECIMAL(23,10)"`
	Status       int       `json:"status" xorm:"default 1 TINYINT(4)"`
	Address      string    `json:"address" xorm:"TEXT"`
	Key          string    `json:"key" xorm:"TEXT"`
	Updatetime   time.Time `json:"updatetime" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
	SmallAddress string    `json:"small_address" xorm:"not null default '' comment('找零地址') VARCHAR(255)"`
}
