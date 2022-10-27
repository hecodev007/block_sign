package entity

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"time"
	"xorm.io/builder"
)

type FcMch struct {
	Id            int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	AgentId       int       `json:"agent_id" xorm:"default 0 comment('代理商ID') INT(10)"`
	Platform      string    `json:"platform" xorm:"comment('商户简称') unique VARCHAR(50)"`
	LoginName     string    `json:"login_name" xorm:"comment('登录名') unique VARCHAR(150)"`
	Password      string    `json:"password" xorm:"comment('密码') VARCHAR(255)"`
	Salt          string    `json:"salt" xorm:"comment('盐') VARCHAR(255)"`
	ApiKey        string    `json:"api_key" xorm:"comment('API KEY') VARCHAR(255)"`
	ApiSecret     string    `json:"api_secret" xorm:"not null comment('商户api_secret') VARCHAR(255)"`
	ApiPublicKey  string    `json:"api_public_key" xorm:"not null comment('商户api公钥') VARCHAR(255)"`
	GoogleSecret  string    `json:"google_secret" xorm:"comment('google安全密匙') VARCHAR(32)"`
	Status        int       `json:"status" xorm:"not null default 1 comment('1未审核2已审核3已过期') TINYINT(3)"`
	IsFrozen      int       `json:"is_frozen" xorm:"not null default 1 comment('0冻结1正常') TINYINT(3)"`
	LastLoginIp   string    `json:"last_login_ip" xorm:"comment('最后登录IP') VARCHAR(20)"`
	LastLoginTime time.Time `json:"last_login_time" xorm:"default '0000-00-00 00:00:00' comment('最后登录时间') DATETIME"`
	CreateAt      int       `json:"create_at" xorm:"not null default 0 INT(11)"`
	UpdateAt      int       `json:"update_at" xorm:"not null default 0 INT(11)"`
}

func (o FcMch) Find(cond builder.Cond) ([]*FcMch, error) {
	res := make([]*FcMch, 0)
	if err := db.Conn.Where(cond).Desc("id").Find(&res); err != nil {
		return nil, err
	}
	return res, nil
}
