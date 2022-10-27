package entity

import (
	"time"
)

type FcOrderTxsPush struct {
	Id           int64     `json:"id" xorm:"not null pk autoincr BIGINT(11)"`
	OrderTxsId   int64     `json:"order_txs_id"  xorm:" NOT NULL comment('订单交易id')  BIGINT"`
	TxId         string    `json:"tx_id" xorm:" comment('交易哈希')  VARCHAR(256)"`
	Memo         string    `json:"memo" xorm:" comment('交易备注')  VARCHAR(512)"`
	BlockHeight  int64     `json:"block_height" xorm:" comment('交易备注')  BIGINT"`
	Confirmation int       `json:"confirmation" xorm:" comment('确认数')  INT"`
	ConfirmTime  int64     `json:"confirm_time" xorm:" comment('确认时间')  BIGINT"`
	IsIn         int       `json:"is_in" xorm:" comment('是否入账；1：入账，2：出账') INT"`
	Fee          string    `json:"memo" xorm:" comment('交易备注')  VARCHAR(512)"`
	TrxN         int       `json:"trx_n" xorm:" comment('是否入账；1：入账，2：出账') INT"`
	CreateTime   time.Time `json:"create_time" xorm:"not null comment('创建时间') DATETIME"`
}

func (o *FcOrderTxsPush) TableName() string {
	return "fc_order_txs_push"
}

