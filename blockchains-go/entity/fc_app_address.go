package entity

import (
	"time"
)

type FcAppAddress struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	AppId      int       `json:"app_id" xorm:"unique(idx_appid_coinid) INT(11)"`
	AppAddress string    `json:"app_address" xorm:"VARCHAR(160)"`
	Updatetime time.Time `json:"updatetime" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
	CoinId     int       `json:"coin_id" xorm:"unique(idx_appid_coinid) INT(11)"`
	Num        string    `json:"num" xorm:"not null default 0.0000000000 comment('分配余额') DECIMAL(23,10)"`
	SurplusNum string    `json:"surplus_num" xorm:"not null default 0.0000000000 comment('当前余额') DECIMAL(23,10)"`
	Type       int       `json:"type" xorm:"comment('1普通，2巨额') TINYINT(4)"`
	Status     int       `json:"status" xorm:"default 1 TINYINT(4)"`
}
