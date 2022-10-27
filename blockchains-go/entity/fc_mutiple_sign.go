package entity

import (
	"time"
)

type FcMutipleSign struct {
	Id            int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Address       string    `json:"address" xorm:"comment('多重签名地址') VARCHAR(150)"`
	AuditAddress  string    `json:"audit_address" xorm:"VARCHAR(30)"`
	Updatetime    time.Time `json:"updatetime" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
	CoinId        string    `json:"coin_id" xorm:"VARCHAR(20)"`
	Status        int       `json:"status" xorm:"default 0 TINYINT(4)"`
	ConnetId      int       `json:"connet_id" xorm:"INT(11)"`
	TradeConnetid int       `json:"trade_connetid" xorm:"INT(11)"`
	AppId         int       `json:"app_id" xorm:"not null comment('平台id') INT(11)"`
}
