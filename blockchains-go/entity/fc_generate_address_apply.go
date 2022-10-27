package entity

import (
	"time"
)

type FcGenerateAddressApply struct {
	Id         int       `json:"id" xorm:"not null pk autoincr comment('自增ID') INT(11)"`
	Username   string    `json:"username" xorm:"not null default '' comment('登录后台的账号') VARCHAR(32)"`
	CoinName   string    `json:"coin_name" xorm:"not null default '' comment('申请币种名称') VARCHAR(255)"`
	ApplyNum   string    `json:"apply_num" xorm:"not null default '' comment('申请地址数量') VARCHAR(255)"`
	Platform   string    `json:"platform" xorm:"not null default '' comment('申请商户，如hoo') VARCHAR(255)"`
	OutOrderid string    `json:"out_orderid" xorm:"not null default '' comment('合作方订单ID') index VARCHAR(64)"`
	Status     int       `json:"status" xorm:"not null default 1 comment('状态, 0-删除, 1-提交申请, 2-处理中, 3-全部完成, 4-失败, 5-部分完成, 6-未知状态') TINYINT(2)"`
	CallBack   string    `json:"call_back" xorm:"not null default '' comment('回调url') VARCHAR(255)"`
	Createtime int       `json:"createtime" xorm:"not null comment('申请时间') INT(11)"`
	Lastmodify time.Time `json:"lastmodify" xorm:"default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
	PlatformId int       `json:"platform_id" xorm:"not null default 0 comment('商户ID') INT(11)"`
	IsBuild    int       `json:"is_build" xorm:"not null default 1 comment('是否预生成地址 1 预生成 2 用户申请') TINYINT(4)"`
}
