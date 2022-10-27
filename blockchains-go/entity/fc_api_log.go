package entity

import (
	"time"
)

type FcApiLog struct {
	LogId     int       `json:"log_id" xorm:"not null pk autoincr INT(11)"`
	AppId     int       `json:"app_id" xorm:"not null index INT(11)"`
	Ipaddress string    `json:"ipaddress" xorm:"not null VARCHAR(20)"`
	SqlData   string    `json:"sql_data" xorm:"not null comment('执行的数据库语句') TEXT"`
	Ipproxy   string    `json:"ipproxy" xorm:"VARCHAR(20)"`
	Moudle    string    `json:"moudle" xorm:"not null comment('发生模块') VARCHAR(20)"`
	Action    string    `json:"action" xorm:"not null comment('发生的控制器') VARCHAR(20)"`
	Logtime   time.Time `json:"logtime" xorm:"not null default CURRENT_TIMESTAMP comment('创建时间') TIMESTAMP"`
	MId       int       `json:"m_id" xorm:"not null default 0 comment('模块ID') index INT(11)"`
	GroupId   int       `json:"group_id" xorm:"not null INT(11)"`
	Params    string    `json:"params" xorm:"TEXT"`
}
