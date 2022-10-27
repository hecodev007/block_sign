package entity

import "time"

type FcTxBreak struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	OutOrderNo string    `json:"out_order_no" xorm:"comment('订单编号') VARCHAR(256)"`
	TxId       string    `json:"tx_id" xorm:"comment('订单编号') VARCHAR(256)"`
	Chain      string    `json:"chain" xorm:"comment('链') VARCHAR(16)"`
	CoinCode   string    `json:"coin_code" xorm:"comment('币种') VARCHAR(16)"`
	CreateTime time.Time `json:"create_time" xorm:"not null comment('创建时间') DATETIME"`
}


