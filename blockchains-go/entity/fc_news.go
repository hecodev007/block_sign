package entity

import (
	"time"
)

type FcNews struct {
	Id      int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Api     string    `json:"api" xorm:"not null VARCHAR(255)"`
	Title   string    `json:"title" xorm:"VARCHAR(255)"`
	Content string    `json:"content" xorm:"VARCHAR(1000)"`
	Status  int       `json:"status" xorm:"default 0 comment('0未读1已读') TINYINT(1)"`
	Addtime time.Time `json:"addtime" xorm:"DATETIME"`
}
