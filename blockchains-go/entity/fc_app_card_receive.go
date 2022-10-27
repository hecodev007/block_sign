package entity

import (
	"time"
)

type FcAppCardReceive struct {
	Id          int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	AppId       int       `json:"app_id" xorm:"INT(11)"`
	CoinId      int       `json:"coin_id" xorm:"index(coin_id) INT(11)"`
	Address     string    `json:"address" xorm:"VARCHAR(80)"`
	Num         string    `json:"num" xorm:"not null default 0.0000000000 comment('领取金额') DECIMAL(23,10)"`
	Obtain      int       `json:"obtain" xorm:"not null default 0 comment('领取数量') INT(11)"`
	Code        string    `json:"code" xorm:"not null comment('范奈斯交易编号') VARCHAR(50)"`
	TradeId     string    `json:"trade_id" xorm:"not null comment('平台交易编号') index(coin_id) VARCHAR(255)"`
	CreateTime  time.Time `json:"create_time" xorm:"DATETIME"`
	TradeTime   string    `json:"trade_time" xorm:"DECIMAL(18,6)"`
	Type        int       `json:"type" xorm:"comment('1个人') TINYINT(1)"`
	CardStartId int       `json:"card_start_id" xorm:"INT(11)"`
}
