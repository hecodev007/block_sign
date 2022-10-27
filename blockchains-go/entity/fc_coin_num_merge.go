package entity

import (
	"time"
)

type FcCoinNumMerge struct {
	Id          int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Platform    string    `json:"platform" xorm:"not null comment('商户名称, 如hoo') index(platform) VARCHAR(32)"`
	CoinName    string    `json:"coin_name" xorm:"not null comment('币种名称') index(platform) VARCHAR(15)"`
	Address     string    `json:"address" xorm:"not null comment('归集地址(一般都是填写冷钱包地址)') index VARCHAR(255)"`
	RestrictNum string    `json:"restrict_num" xorm:"not null default 0.0000000000000000 comment('归集设置，即当用户余额达到这个限制时，开始归集') DECIMAL(40,16)"`
	Conversion  int       `json:"conversion" xorm:"default 1 comment('换算位数，比如BTC的换算位数是8') INT(11)"`
	Createtime  int       `json:"createtime" xorm:"not null comment('创建时间') INT(11)"`
	Lastmodify  time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
}
