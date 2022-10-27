package entity

import (
	"time"
)

type FcColdVerHsr struct {
	Id          int64     `json:"id" xorm:"pk autoincr BIGINT(11)"`
	Address     string    `json:"address" xorm:"not null index unique(tdate_address) VARCHAR(80)"`
	Num         string    `json:"num" xorm:"not null default 0.0000000000 comment('余额') DECIMAL(40,10)"`
	AddressType int       `json:"address_type" xorm:"not null comment('地址类型：1冷钱包平台地址2热钱包3普通用户地址') TINYINT(4)"`
	CoinId      int       `json:"coin_id" xorm:"default 0 INT(10)"`
	TDate       time.Time `json:"t_date" xorm:"not null unique(tdate_address) DATE"`
}
