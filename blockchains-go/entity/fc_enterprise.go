package entity

import (
	"time"
)

type FcEnterprise struct {
	Id        int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Title     string    `json:"title" xorm:"comment('名称') VARCHAR(255)"`
	Phone     string    `json:"phone" xorm:"comment('联系电话') VARCHAR(20)"`
	StartTime string    `json:"start_time" xorm:"comment('合作开始时间') VARCHAR(20)"`
	EndTime   string    `json:"end_time" xorm:"comment('合作结束时间') VARCHAR(20)"`
	Content   string    `json:"content" xorm:"comment('备注') VARCHAR(255)"`
	Place     string    `json:"place" xorm:"comment('公司地址') VARCHAR(255)"`
	Addtime   time.Time `json:"addtime" xorm:"comment('添加时间') DATETIME"`
	Status    int       `json:"status" xorm:"default 1 comment('0禁用1启用') TINYINT(1)"`
}
