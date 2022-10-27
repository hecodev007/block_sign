package entity

import (
	"time"
)

type FcCoinBalance struct {
	CoinId     int       `json:"coin_id" xorm:"not null pk INT(11)"`
	SurplusNum string    `json:"surplus_num" xorm:"not null default 0.0000000000 comment('剩余数量') DECIMAL(23,10)"`
	Edition    int       `json:"edition" xorm:"not null default 0 comment('版本号') INT(11)"`
	Updatetime time.Time `json:"updatetime" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP"`
}
