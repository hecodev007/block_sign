package entity

import (
	"github.com/group-coldwallet/blockchains-go/db"
	"time"
)

type OrderPriorityStatus int

const (
	OrderPriorityStatusProcessing OrderPriorityStatus = 1 //正在处理
	OrderPriorityStatusCompleted  OrderPriorityStatus = 2 //订单已完成
)

type FcOrderPriority struct {
	Id           int                 `json:"id" xorm:"not null pk autoincr comment('自增ID') INT(11)"`
	ApplyId      int                 `json:"apply_id" xorm:"not null comment('对应fc_transfers_apply表的 id') INT(11)"`
	OuterOrderNo string              `json:"outer_order_no" xorm:"not null default '' comment('订单外部编号') VARCHAR(64)"`
	ChainName    string              `json:"chain_name" xorm:"not null default '' comment('链名') VARCHAR(16)"`
	CoinCode     string              `json:"coin_code" xorm:"not null default '' comment('币名') VARCHAR(16)"`
	Status       OrderPriorityStatus `json:"status" xorm:"not null default 1 comment('1:正在处理，2:订单已完成') INT(11)"`
	MchId        int                 `json:"mch_id" xorm:"not null default 0 comment('商户id') INT(11)"`
	CreateTime   time.Time           `json:"create_time" xorm:"not null comment('创建时间') DATETIME"`
}

func (o *FcOrderPriority) Add() (int64, error) {
	return db.Conn.InsertOne(o)
}
