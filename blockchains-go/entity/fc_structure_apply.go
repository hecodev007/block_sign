package entity

import (
	"time"
)

type FcStructureApply struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Username   string    `json:"username" xorm:"comment('登陆账号名') VARCHAR(255)"`
	Department string    `json:"department" xorm:"comment('申请部门') VARCHAR(255)"`
	Applicant  string    `json:"applicant" xorm:"comment('申请人') VARCHAR(255)"`
	Operator   string    `json:"operator" xorm:"comment('操作人') VARCHAR(255)"`
	Type       string    `json:"type" xorm:"comment('出账，归集') ENUM('cz','gj')"`
	Purpose    string    `json:"purpose" xorm:"comment('出库用途') TEXT"`
	Status     int       `json:"status" xorm:"default 0 comment('交易状态0未审核1审核通过2审核驳回') TINYINT(255)"`
	Createtime int       `json:"createtime" xorm:"INT(11)"`
	Lastmodify time.Time `json:"lastmodify" xorm:"default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
	Memo       string    `json:"memo" xorm:"not null default '' comment('eos memo') VARCHAR(100)"`
}
