package entity

import (
	"time"
)

type FcAuditLog struct {
	LogId     int       `json:"log_id" xorm:"not null pk autoincr index INT(11)"`
	UserId    int       `json:"user_id" xorm:"not null index INT(11)"`
	GroupId   int       `json:"group_id" xorm:"not null INT(11)"`
	Ipaddress string    `json:"ipaddress" xorm:"not null VARCHAR(20)"`
	SqlData   string    `json:"sql_data" xorm:"not null comment('执行的数据库语句') TEXT"`
	Ipproxy   string    `json:"ipproxy" xorm:"VARCHAR(20)"`
	Logtime   time.Time `json:"logtime" xorm:"not null default CURRENT_TIMESTAMP comment('创建时间') TIMESTAMP"`
	MId       int       `json:"m_id" xorm:"not null comment('模块ID') index INT(11)"`
	ServiceId int       `json:"service_id" xorm:"comment('服务ID') INT(11)"`
	AuditId   int       `json:"audit_id" xorm:"comment('审核组ID') INT(11)"`
	Level     int       `json:"level" xorm:"comment('审核等级') TINYINT(4)"`
	Type      int       `json:"type" xorm:"default 1 comment('1：审核通过，2：驳回') TINYINT(4)"`
	Username  string    `json:"username" xorm:"VARCHAR(255)"`
}
