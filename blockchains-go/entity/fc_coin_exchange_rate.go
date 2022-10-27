package entity

import (
	"time"
)

type FcCoinExchangeRate struct {
	Id             int       `json:"id" xorm:"not null pk autoincr comment('自增ID') INT(11)"`
	Name           string    `json:"name" xorm:"not null comment('英文简写') unique(name_date_time) VARCHAR(20)"`
	Createdate     time.Time `json:"createdate" xorm:"not null default '0000-00-00' comment('抓取日期') DATE"`
	CreateDatetime time.Time `json:"create_datetime" xorm:"not null default '0000-00-00 00:00:00' comment('抓取的日期时间, 一日多次对账新增的字段') index unique(name_date_time) DATETIME"`
	Rate           string    `json:"rate" xorm:"not null comment('相对于BTC的汇率') DECIMAL(28,20)"`
	PriceCny       int       `json:"price_cny" xorm:"not null default 0 comment('人民币价格') INT(11)"`
	Unit           string    `json:"unit" xorm:"not null default '' comment('单位, ram用于存放兑eos价格(单位EOS/KB)') VARCHAR(30)"`
	Lastmodify     time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
}
