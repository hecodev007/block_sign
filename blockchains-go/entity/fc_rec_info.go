package entity

import (
	"time"
)

type FcRecInfo struct {
	Id       int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	Code     string    `json:"code" xorm:"not null comment('编号(用户名)') VARCHAR(50)"`
	FileName string    `json:"file_name" xorm:"not null comment('文件名(显示用，无实际意义)') VARCHAR(30)"`
	Addtime  int       `json:"addtime" xorm:"not null comment('添加时间') INT(11)"`
	ConTime  time.Time `json:"con_time" xorm:"not null comment('对帐时间') DATE"`
	Content  string    `json:"content" xorm:"not null comment('备注') TEXT"`
	Path     string    `json:"path" xorm:"not null VARCHAR(255)"`
	Username string    `json:"username" xorm:"not null VARCHAR(20)"`
	Status   int       `json:"status" xorm:"not null default 0 comment('0未对账1已对账') TINYINT(4)"`
	Suffix   string    `json:"suffix" xorm:"not null comment('后缀') VARCHAR(15)"`
}
