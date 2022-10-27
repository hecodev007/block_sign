package entity

import (
	"time"
)

type FcMchOrderGoods struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	OrderId    int       `json:"order_id" xorm:"not null default 0 comment('订单id') INT(10)"`
	OrderNo    string    `json:"order_no" xorm:"not null comment('订单ID') VARCHAR(32)"`
	MchId      int       `json:"mch_id" xorm:"not null comment('商户ID') INT(11)"`
	Type       int       `json:"type" xorm:"not null comment('订单类型 1 购买服务  2 加购地址 3 提现') TINYINT(255)"`
	CoinId     int       `json:"coin_id" xorm:"INT(11)"`
	CoinName   string    `json:"coin_name" xorm:"not null comment('币种服务') VARCHAR(255)"`
	Money      string    `json:"money" xorm:"not null default 0.00000000 comment('总额') DECIMAL(20,8)"`
	Price      string    `json:"price" xorm:"not null default 0.00000000 comment('单价') DECIMAL(20,8)"`
	Num        int       `json:"num" xorm:"not null default 1 comment('购买数量') INT(11)"`
	Term       int       `json:"term" xorm:"default 0 comment('购买币种服务年限') INT(11)"`
	Status     int       `json:"status" xorm:"not null default 0 comment('0待处理1已完成2处理失败') TINYINT(4)"`
	Createtime int       `json:"createtime" xorm:"not null INT(11)"`
	Lastmodify time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP"`
}
