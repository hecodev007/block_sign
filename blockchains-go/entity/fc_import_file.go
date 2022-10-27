package entity

import (
	"time"
)

type FcImportFile struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Username   string    `json:"username" xorm:"not null VARCHAR(20)"`
	AppId      int       `json:"app_id" xorm:"INT(11)"`
	CoinId     int       `json:"coin_id" xorm:"INT(11)"`
	FileName   string    `json:"file_name" xorm:"VARCHAR(50)"`
	Count      int       `json:"count" xorm:"INT(11)"`
	CreateTime time.Time `json:"create_time" xorm:"DATETIME"`
}
