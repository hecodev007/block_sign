package entity

import (
	"time"
)

type FcAppCoinDetailed struct {
	Id              int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	AppAddressId    int       `json:"app_address_id" xorm:"not null INT(11)"`
	Address         string    `json:"address" xorm:"comment('发起账号') VARCHAR(80)"`
	OppositeAddress string    `json:"opposite_address" xorm:"comment('接收账号') VARCHAR(80)"`
	Type            int       `json:"type" xorm:"comment('1转入2转出') TINYINT(1)"`
	Txid            string    `json:"txid" xorm:"VARCHAR(80)"`
	Balance         string    `json:"balance" xorm:"not null default 0.0000000000 comment('平台余额（由平台通过接口提供）') DECIMAL(23,10)"`
	Before          string    `json:"before" xorm:"not null default 0.0000000000 comment('之前余额') DECIMAL(23,10)"`
	Change          string    `json:"change" xorm:"not null default 0.0000000000 comment('改变金额') DECIMAL(23,10)"`
	After           string    `json:"after" xorm:"not null default 0.0000000000 DECIMAL(23,10)"`
	Fee             string    `json:"fee" xorm:"not null default 0.0000000000 DECIMAL(23,10)"`
	Addtime         time.Time `json:"addtime" xorm:"DATETIME"`
	Code            string    `json:"code" xorm:"comment('交易编号') VARCHAR(20)"`
}
