package entity

import (
	"time"
)

type FcMchLoginLog struct {
	Id            int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	MchId         int       `json:"mch_id" xorm:"not null comment('关联用户表ID') INT(11)"`
	LoginName     string    `json:"login_name" xorm:"not null comment('登陆账号') VARCHAR(100)"`
	Company       string    `json:"company" xorm:"not null comment('用户名') VARCHAR(60)"`
	Creattime     time.Time `json:"creattime" xorm:"not null default CURRENT_TIMESTAMP comment('创建时间') TIMESTAMP"`
	LastLoginIp   int       `json:"last_login_ip" xorm:"not null comment('最后登录IP') INT(11)"`
	LastLoginTime int       `json:"last_login_time" xorm:"not null comment('最后登录时间') INT(11)"`
	Avatarurl     string    `json:"avatarurl" xorm:"comment('用户头像url') VARCHAR(100)"`
}
