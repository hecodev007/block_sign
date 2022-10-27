package entity

import (
	"time"
)

type FcAuditAddress struct {
	Id       int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Address  string    `json:"address" xorm:"VARCHAR(255)"`
	ConnetId int       `json:"connet_id" xorm:"index INT(11)"`
	CoinId   int       `json:"coin_id" xorm:"INT(11)"`
	Addtime  time.Time `json:"addtime" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP"`
	Status   int       `json:"status" xorm:"not null default 1 TINYINT(4)"`
	AuditId  int       `json:"audit_id" xorm:"INT(11)"`
	Type     int       `json:"type" xorm:"not null default 0 comment('1已使用0未使用') TINYINT(1)"`
}
