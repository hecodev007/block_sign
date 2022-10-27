package entity

import (
	"time"
)

type FcSms struct {
	MId        int       `json:"m_id" xorm:"not null pk autoincr INT(11)"`
	Name       string    `json:"name" xorm:"VARCHAR(25)"`
	Operator   string    `json:"operator" xorm:"VARCHAR(50)"`
	User       string    `json:"user" xorm:"VARCHAR(50)"`
	Pwd        string    `json:"pwd" xorm:"VARCHAR(255)"`
	Key        string    `json:"key" xorm:"VARCHAR(50)"`
	Status     int       `json:"status" xorm:"default 0 TINYINT(4)"`
	Addtime    int64     `json:"addtime" xorm:"BIGINT(20)"`
	Updatetime time.Time `json:"updatetime" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
}
