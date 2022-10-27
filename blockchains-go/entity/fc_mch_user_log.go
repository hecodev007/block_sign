package entity

import (
	"time"
)

type FcMchUserLog struct {
	Id            int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	MchId         int       `json:"mch_id" xorm:"not null INT(11)"`
	LoginName     string    `json:"login_name" xorm:"not null VARCHAR(60)"`
	Company       string    `json:"company" xorm:"not null VARCHAR(100)"`
	Creattime     time.Time `json:"creattime" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP"`
	LastLoginIp   int       `json:"last_login_ip" xorm:"not null comment('最后登录IP') INT(11)"`
	LastLoginTime int       `json:"last_login_time" xorm:"not null comment('最后登录时间') INT(11)"`
	Module        string    `json:"module" xorm:"comment('操作模块') VARCHAR(20)"`
	ModuleDetail  string    `json:"module_detail" xorm:"comment('操作详情') VARCHAR(20)"`
}
