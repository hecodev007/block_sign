package entity

import (
	"time"
)

type FcAppCardStart struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	AppId      int       `json:"app_id" xorm:"INT(11)"`
	CoinId     int       `json:"coin_id" xorm:"unique(trade_id) INT(11)"`
	Address    string    `json:"address" xorm:"VARCHAR(80)"`
	Num        string    `json:"num" xorm:"not null default 0.0000000000 comment('总金额') DECIMAL(23,10)"`
	ObtainNum  string    `json:"obtain_num" xorm:"not null default 0.0000000000 comment('领取金额') DECIMAL(23,10)"`
	OutNum     string    `json:"out_num" xorm:"not null default 0.0000000000 comment('退回金额') DECIMAL(23,10)"`
	Total      int       `json:"total" xorm:"not null default 0 comment('总数量') INT(11)"`
	Obtain     int       `json:"obtain" xorm:"not null default 0 comment('领取数量') INT(11)"`
	Out        int       `json:"out" xorm:"not null default 0 comment('退回数量') INT(11)"`
	Code       string    `json:"code" xorm:"not null comment('范奈斯交易编号') unique VARCHAR(50)"`
	TradeId    string    `json:"trade_id" xorm:"not null comment('平台交易编号') unique(trade_id) VARCHAR(255)"`
	CreateTime time.Time `json:"create_time" xorm:"DATETIME"`
	TradeTime  string    `json:"trade_time" xorm:"DECIMAL(18,6)"`
	Type       int       `json:"type" xorm:"not null comment('1个人') TINYINT(1)"`
	IsExamice  int       `json:"is_examice" xorm:"not null default 0 comment('0不审核1审核') TINYINT(1)"`
	Price      string    `json:"price" xorm:"not null default 0.0000000000 comment('每次领取金额') DECIMAL(23,10)"`
}
