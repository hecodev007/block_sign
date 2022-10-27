package entity

import (
	"time"
)

type FcAccess struct {
	Id               int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Uuid             string    `json:"uuid" xorm:"comment('用户唯一标识') VARCHAR(100)"`
	Username         string    `json:"username" xorm:"comment('用户名') VARCHAR(60)"`
	Password         string    `json:"password" xorm:"comment('密码') VARCHAR(100)"`
	Salt             string    `json:"salt" xorm:"VARCHAR(5)"`
	FirstName        string    `json:"first_name" xorm:"comment('姓氏') VARCHAR(60)"`
	LastName         string    `json:"last_name" xorm:"comment('名字') VARCHAR(60)"`
	Email            string    `json:"email" xorm:"comment('邮箱') VARCHAR(60)"`
	Mobile           string    `json:"mobile" xorm:"comment('电话号码') VARCHAR(100)"`
	Company          string    `json:"company" xorm:"comment('企业名称') VARCHAR(100)"`
	CompanyUrl       string    `json:"company_url" xorm:"comment('企业网址') VARCHAR(100)"`
	CompanyType      int       `json:"company_type" xorm:"not null default 0 comment('企业类型  1 交易所 2 对冲/套利基金 3 资产管理 4 场外交易 5 私人银行 6 其他') TINYINT(3)"`
	Code             string    `json:"code" xorm:"comment('组织机构代码') VARCHAR(60)"`
	State            string    `json:"state" xorm:"comment('所在国家') VARCHAR(50)"`
	City             string    `json:"city" xorm:"comment('所在城市') VARCHAR(50)"`
	Address          string    `json:"address" xorm:"comment('企业地址') VARCHAR(255)"`
	Remark           string    `json:"remark" xorm:"comment('备注') VARCHAR(255)"`
	ApiKey           string    `json:"api_key" xorm:"comment('商户应用标识符') unique VARCHAR(100)"`
	PlatformAddr     string    `json:"platform_addr" xorm:"not null comment('商户名称') VARCHAR(255)"`
	PrivateKey       string    `json:"private_key" xorm:"comment('商户私钥') TEXT"`
	PublicKey        string    `json:"public_key" xorm:"comment('商户公钥') TEXT"`
	Creattime        time.Time `json:"creattime" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP"`
	Updatetime       int       `json:"updatetime" xorm:"default 0 INT(11)"`
	AuditState       int       `json:"audit_state" xorm:"not null default 0 comment('审核状态 0未审核 1审核通过 2审核失败') TINYINT(3)"`
	CompanyImg       string    `json:"company_img" xorm:"comment('图片路径') VARCHAR(255)"`
	AppName          string    `json:"app_name" xorm:"VARCHAR(25)"`
	WithdrawCallback string    `json:"withdraw_callback" xorm:"VARCHAR(255)"`
	SitIp            string    `json:"sit_ip" xorm:"not null VARCHAR(255)"`
	Domain           string    `json:"domain" xorm:"not null comment('客户端url') VARCHAR(60)"`
	AccStatus        int       `json:"acc_status" xorm:"not null default 1 TINYINT(4)"`
	Platform         string    `json:"platform" xorm:"not null default '' comment('商户名称') VARCHAR(50)"`
	EntId            int       `json:"ent_id" xorm:"default 0 comment('企业ID') INT(11)"`
	Lasttime         int64     `json:"lasttime" xorm:"BIGINT(20)"`
	UpdateNum        int       `json:"update_num" xorm:"not null default 0 comment('更新次数') INT(11)"`
	Status           int       `json:"status" xorm:"default 1 TINYINT(4)"`
	LastLoginIp      int       `json:"last_login_ip" xorm:"not null comment('最后登录IP') INT(11)"`
	LastLoginTime    int       `json:"last_login_time" xorm:"not null comment('最后登录时间') INT(11)"`
	GroupId          int       `json:"group_id" xorm:"default 0 INT(11)"`
	Avatarurl        string    `json:"avatarurl" xorm:"comment('用户头像url') VARCHAR(100)"`
	UserStatus       int       `json:"user_status" xorm:"default 1 comment('是否审核 1未审核 2已审核 3审核失败') TINYINT(4)"`
	StatusContent    string    `json:"status_content" xorm:"comment('审核备注') VARCHAR(100)"`
}
