package entity

type FcGroup struct {
	Id          int    `json:"id" xorm:"not null pk autoincr comment('用户组id,自增主键') MEDIUMINT(8)"`
	Module      string `json:"module" xorm:"not null default '' comment('用户组所属模块') VARCHAR(20)"`
	Type        int    `json:"type" xorm:"not null default 2 comment('组类型') TINYINT(4)"`
	Title       string `json:"title" xorm:"not null default '' comment('用户组中文名称') CHAR(20)"`
	Description string `json:"description" xorm:"not null default '' comment('描述信息') VARCHAR(80)"`
	Status      int    `json:"status" xorm:"not null default 1 comment('用户组状态：为1正常，为0禁用,-1为删除') TINYINT(1)"`
	Rules       string `json:"rules" xorm:"not null default '' comment('用户组拥有的规则id，多个规则 , 隔开') VARCHAR(500)"`
	Privilege   string `json:"privilege" xorm:"comment('权限信息') TEXT"`
}
