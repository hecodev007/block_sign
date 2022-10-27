package entity

import (
	"time"
)

type FcGlyEmail struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	Email      string    `json:"email" xorm:"VARCHAR(100)"`
	Title      string    `json:"title" xorm:"VARCHAR(100)"`
	Content    string    `json:"content" xorm:"TEXT"`
	Status     int       `json:"status" xorm:"default 1 comment('1未发送2已发送') TINYINT(1)"`
	Type       int       `json:"type" xorm:"default 1 comment('消息类型  1 邮件  2  短信') TINYINT(3)"`
	TemplateId int       `json:"template_id" xorm:"default 0 comment('模板id') INT(11)"`
	ErrorNum   int       `json:"error_num" xorm:"not null default 0 TINYINT(1)"`
	Createtime int       `json:"createtime" xorm:"default 0 INT(11)"`
	Lastmodify time.Time `json:"lastmodify" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
}
