package entity

import (
	"time"
)

type FcFinance struct {
	FId       int       `json:"f_id" xorm:"not null pk autoincr INT(11)"`
	AddressId int       `json:"address_id" xorm:"INT(11)"`
	Type      int       `json:"type" xorm:"comment('1.转入2.转出') TINYINT(4)"`
	CoinId    int       `json:"coin_id" xorm:"INT(11)"`
	Before    string    `json:"before" xorm:"not null default 0.0000000000 comment('之前余额') DECIMAL(23,10)"`
	After     string    `json:"after" xorm:"not null default 0.0000000000 comment('之后余额') DECIMAL(23,10)"`
	Change    string    `json:"change" xorm:"not null default 0.0000000000 comment('操作金额') DECIMAL(23,10)"`
	Addtime   time.Time `json:"addtime" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP"`
	Remark    string    `json:"remark" xorm:"comment('备注') VARCHAR(255)"`
	Service   string    `json:"service" xorm:"comment('业务模块') VARCHAR(11)"`
	ServiceId int       `json:"service_id" xorm:"comment('业务ID') INT(11)"`
	AppId     int       `json:"app_id" xorm:"not null INT(11)"`
	Fee       string    `json:"fee" xorm:"not null default 0.0000000000 DECIMAL(23,10)"`
}
