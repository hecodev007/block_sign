package entity

import (
	"time"
)

type FcFixAddress struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Address    string    `json:"address" xorm:"default '' comment('地址') VARCHAR(256)"`
	ChainName  string    `json:"chain_name" xorm:"default '' comment('链名')  VARCHAR(20)"`
	Payed      int       `json:"payed" xorm:"default 0 comment('是否已打手续费;0：否；1：是 ')INT(11)"`
	Amount     string    `json:"amount" xorm:"default 0 comment('金额 ') DECIMAL(20,8)"`
	Status     int       `json:"status" xorm:"default 0 comment('0:准备,1:正在使用,2:禁用') INT(11)"`
	CreateTime time.Time `json:"create_time" xorm:"not null comment('创建时间') DATETIME"`
}
