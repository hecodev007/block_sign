package entity

type FcMchProfile struct {
	Id          int    `json:"id" xorm:"not null pk autoincr INT(10)"`
	MchId       int    `json:"mch_id" xorm:"not null INT(11)"`
	Code        string `json:"code" xorm:"comment('商户代码') VARCHAR(50)"`
	FirstName   string `json:"first_name" xorm:"comment('名') VARCHAR(255)"`
	LastName    string `json:"last_name" xorm:"comment('姓') VARCHAR(255)"`
	Email       string `json:"email" xorm:"not null comment('商户邮箱') unique VARCHAR(50)"`
	Mobile      string `json:"mobile" xorm:"comment('商户手机') VARCHAR(20)"`
	Company     string `json:"company" xorm:"comment('企业全称') VARCHAR(255)"`
	CompanyUrl  string `json:"company_url" xorm:"comment('企业URL') VARCHAR(255)"`
	CompanyType string `json:"company_type" xorm:"comment('企业类型  1 交易所 2 对冲/套利基金 3 资产管理 4 场外交易 5 私人银行 6 其他') VARCHAR(255)"`
	CompanyImg  string `json:"company_img" xorm:"comment('公司照片URL') VARCHAR(255)"`
	State       string `json:"state" xorm:"comment('国家') VARCHAR(20)"`
	MchType     int    `json:"mch_type" xorm:"default 0 comment('注册类型  1 企业   2  个人') TINYINT(3)"`
	Position    string `json:"position" xorm:"comment('职位') VARCHAR(100)"`
	Large       string `json:"large" xorm:"comment('资金体量') VARCHAR(100)"`
	City        string `json:"city" xorm:"comment('城市') VARCHAR(20)"`
	Address     string `json:"address" xorm:"comment('地址') VARCHAR(100)"`
	Remark      string `json:"remark" xorm:"comment('备注') VARCHAR(255)"`
	Avatarurl   string `json:"avatarurl" xorm:"comment('头像URL') VARCHAR(255)"`
	AuditRemark string `json:"audit_remark" xorm:"comment('审核备注') VARCHAR(255)"`
	CreateAt    int    `json:"create_at" xorm:"INT(11)"`
	UpdateAt    int    `json:"update_at" xorm:"INT(11)"`
}
