package entity

import (
	"time"
)

type FcMchOrder struct {
	Id          int       `json:"id" xorm:"not null pk autoincr comment('id') INT(11)"`
	OrderNo     string    `json:"order_no" xorm:"not null comment('订单号') VARCHAR(32)"`
	MchId       int       `json:"mch_id" xorm:"not null default 0 comment('商户id') index(borrow_id) INT(10)"`
	OrderAmount string    `json:"order_amount" xorm:"not null default 0.000000000000000000 comment('订单金额') DECIMAL(40,18)"`
	Type        int       `json:"type" xorm:"not null default 0 comment('订单类型 1 购买服务  2 加购地址 3 提现') TINYINT(3)"`
	Status      int       `json:"status" xorm:"not null default 0 comment('订单状态 0 未开始 1 准备中 2 已完成 3取消订单') index(borrow_id) index TINYINT(3)"`
	PayStatus   int       `json:"pay_status" xorm:"not null default 0 comment('支付状态 0 未支付  1 已支付') TINYINT(3)"`
	Editiion    int       `json:"editiion" xorm:"not null default 0 INT(11)"`
	CreateAt    int       `json:"create_at" xorm:"not null default 0 comment('下单时间') INT(11)"`
	PayAt       int       `json:"pay_at" xorm:"not null default 0 comment('支付时间') INT(11)"`
	Source      int       `json:"source" xorm:"not null default 0 comment('0用户下单1后台下单') TINYINT(4)"`
	UpdateAt    time.Time `json:"update_at" xorm:"default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
}
