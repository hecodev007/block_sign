package entity

import (
	"time"
)

type FcHooRtExchange struct {
	Id           int       `json:"id" xorm:"not null pk autoincr comment('自增ID') INT(11)"`
	HooId        int64     `json:"hoo_id" xorm:"not null comment('hoo那边的id') unique(hooid_channel) BIGINT(20)"`
	Channel      int       `json:"channel" xorm:"not null default 0 comment('兑换通道,（1火币，2OK等）') unique(hooid_channel) TINYINT(2)"`
	Symbol       string    `json:"symbol" xorm:"not null comment('交易对（btceos等）') VARCHAR(20)"`
	Price        string    `json:"price" xorm:"not null default 0.00000000000000000000 comment('交易价格') DECIMAL(40,20)"`
	Amount       string    `json:"amount" xorm:"not null default 0.00000000000000000000 comment('交易数量') DECIMAL(40,20)"`
	ActualAmount string    `json:"actual_amount" xorm:"not null default 0.00000000000000000000 comment('实际兑换数量') DECIMAL(40,20)"`
	Fee          string    `json:"fee" xorm:"not null default 0.00000000000000000000 comment('手续费') DECIMAL(40,20)"`
	ExAmount     string    `json:"ex_amount" xorm:"not null default 0.00000000000000000000 comment('交易所兑换数量') DECIMAL(40,20)"`
	ExFee        string    `json:"ex_fee" xorm:"not null default 0.00000000000000000000 comment('交易所手续费') DECIMAL(40,20)"`
	ExFinishAt   string    `json:"ex_finish_at" xorm:"not null default '' comment('交易所结束时间') VARCHAR(20)"`
	TradeNo      string    `json:"trade_no" xorm:"not null default '' comment('交易所订单号') VARCHAR(20)"`
	Status       int       `json:"status" xorm:"not null default 0 comment('状态（#1提交 2交易中 3交易提交 4交易失败 5失效 6交易所完成）') TINYINT(2)"`
	Response     string    `json:"response" xorm:"not null default '' comment('交易所返回信息') VARCHAR(2000)"`
	IsBuy        int       `json:"is_buy" xorm:"not null default 0 comment('是否买入（1买入，0卖出）') TINYINT(2)"`
	CreateAt     int       `json:"create_at" xorm:"not null default 0 comment('创建时间-传过来的数据') INT(11)"`
	Createtime   time.Time `json:"createtime" xorm:"not null default '0000-00-00 00:00:00' comment('创建时间-原始数据转成日期格式，便于查看') DATETIME"`
	Lastmodify   time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最近修改时间') TIMESTAMP"`
}
