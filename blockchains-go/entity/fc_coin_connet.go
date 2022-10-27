package entity

import (
	"time"
)

type FcCoinConnet struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	CoinId     int       `json:"coin_id" xorm:"INT(11)"`
	Connet     string    `json:"connet" xorm:"VARCHAR(80)"`
	Status     int       `json:"status" xorm:"default 1 TINYINT(4)"`
	Updatetime time.Time `json:"updatetime" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
	Addtime    int64     `json:"addtime" xorm:"BIGINT(20)"`
	Title      string    `json:"title" xorm:"VARCHAR(30)"`
	Type       int       `json:"type" xorm:"comment('1.地址生成链接，2.交易打包链接.3多重签名') TINYINT(4)"`
}
