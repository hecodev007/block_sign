package entity

import (
	"time"
)

type FcApiCallback struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	UserId     int       `json:"user_id" xorm:"default 0 unique(coin_name) INT(11)"`
	CoinId     int       `json:"coin_id" xorm:"INT(11)"`
	CoinName   string    `json:"coin_name" xorm:"unique(coin_name) VARCHAR(15)"`
	ApiId      int       `json:"api_id" xorm:"unique(coin_name) INT(11)"`
	Url        string    `json:"url" xorm:"VARCHAR(255)"`
	Createtime int       `json:"createtime" xorm:"INT(11)"`
	Lastmodify time.Time `json:"lastmodify" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
}
