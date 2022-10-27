package entity

import (
	"time"
)

type FcGenerateAddressTask struct {
	Id         int       `json:"id" xorm:"not null pk autoincr comment('自增ID') INT(10)"`
	ApplyId    int       `json:"apply_id" xorm:"not null comment('所属申请ID') INT(10)"`
	Coinname   string    `json:"coinname" xorm:"not null default '' comment('申请币种名称') VARCHAR(16)"`
	Num        int       `json:"num" xorm:"not null comment('本次申请数量') INT(10)"`
	Platform   string    `json:"platform" xorm:"not null comment('商户名称, 如hoo') VARCHAR(32)"`
	Orderid    string    `json:"orderid" xorm:"not null comment('订单编号') unique VARCHAR(64)"`
	Status     int       `json:"status" xorm:"not null default 1 comment('状态, 0-删除, 1-提交申请, 2-处理中, 3-全部完成, 4-失败, 5-部分完成, 6-未知状态') TINYINT(2)"`
	RetryNum   int       `json:"retry_num" xorm:"not null default 0 comment('重试次数') TINYINT(3)"`
	Createtime int       `json:"createtime" xorm:"not null comment('申请时间') INT(11)"`
	BeginTime  int       `json:"begin_time" xorm:"not null default 0 comment('任务开始处理时间') INT(10)"`
	FinishTime int       `json:"finish_time" xorm:"not null default 0 comment('完成时间') INT(10)"`
	Lastmodify time.Time `json:"lastmodify" xorm:"default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
	PlatformId int       `json:"platform_id" xorm:"not null comment('商户ID') INT(11)"`
}
