package entity

import (
	"time"
)

type FcAddressCoin struct {
	Id      int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	CoinId  int       `json:"coin_id" xorm:"not null unique(idx_coinid_address) INT(11)"`
	Address string    `json:"address" xorm:"not null index unique(idx_coinid_address) VARCHAR(255)"`
	Num     string    `json:"num" xorm:"not null default 0.0000000000 comment('用户金额') DECIMAL(23,10)"`
	Addtime time.Time `json:"addtime" xorm:"not null default CURRENT_TIMESTAMP comment('添加时间') TIMESTAMP"`
	AppId   int       `json:"app_id" xorm:"index INT(11)"`
	Status  int       `json:"status" xorm:"default 0 TINYINT(3)"`
	EntId   int       `json:"ent_id" xorm:"not null INT(11)"`
}
