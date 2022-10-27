package entity

import (
	"time"
)

type FcTxTransaction struct {
	Id             int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	TxcId          int       `json:"txc_id" xorm:"not null default 0 comment('清分表id') INT(11)"`
	TxcDetailId    int       `json:"txc_detail_id" xorm:"not null default 0 comment('清分明细表id') INT(11)"`
	AgentId        int       `json:"agent_id" xorm:"not null default 0 comment('代理商id') INT(11)"`
	MchId          int       `json:"mch_id" xorm:"not null default 0 INT(11)"`
	TxType         int       `json:"tx_type" xorm:"not null default 0 comment('1 入账 2 出账 3 归集入账 4 归集出账 5 手续费入账 6 手续费出账 7 余额入账  8 余额归集') TINYINT(3)"`
	OuterOrderNo   string    `json:"outer_order_no" xorm:"not null comment('外部订单号') VARCHAR(255)"`
	OrderNo        string    `json:"order_no" xorm:"not null comment('内部订单号') VARCHAR(255)"`
	CoinType       string    `json:"coin_type" xorm:"not null comment('币种名称') VARCHAR(15)"`
	SettleId       int       `json:"settle_id" xorm:"not null default 0 comment('服务费清分规则id') INT(11)"`
	AgentProfit    int       `json:"agent_profit" xorm:"not null default 0 comment('代理商收益') INT(11)"`
	PlatformProfit int       `json:"platform_profit" xorm:"not null default 0 comment('平台收益') INT(11)"`
	MchServiceFee  int       `json:"mch_service_fee" xorm:"not null default 0 comment('商户服务收费') INT(11)"`
	MchServiceCoin string    `json:"mch_service_coin" xorm:"comment('服务费币种') VARCHAR(15)"`
	BlockHeight    int       `json:"block_height" xorm:"not null default 0 INT(11)"`
	Timestamp      int       `json:"timestamp" xorm:"not null default 0 comment('交易时间戳') INT(11)"`
	TxId           string    `json:"tx_id" xorm:"not null comment('txid') VARCHAR(150)"`
	Amount         string    `json:"amount" xorm:"default 0.000000000000000000 DECIMAL(40,18)"`
	TxFee          string    `json:"tx_fee" xorm:"DECIMAL(40,18)"`
	TxFeeCoin      string    `json:"tx_fee_coin" xorm:"not null comment('矿工费币种') VARCHAR(15)"`
	FromAddress    string    `json:"from_address" xorm:"comment('出账地址') VARCHAR(255)"`
	ToAddress      string    `json:"to_address" xorm:"comment('入账地址') VARCHAR(255)"`
	Memo           string    `json:"memo" xorm:"comment('链上交易备注') VARCHAR(255)"`
	Confirmations  int       `json:"confirmations" xorm:"not null default 0 comment('确认数') INT(10)"`
	NotifyStatus   int       `json:"notify_status" xorm:"not null default 0 comment('通知状态，是否已通知商户') TINYINT(3)"`
	Status         int       `json:"status" xorm:"not null default 1 comment('交易状态 1 正常  2 作废  3 虚拟补单') TINYINT(3)"`
	Remark         string    `json:"remark" xorm:"comment('备注') VARCHAR(255)"`
	CreateAt       int       `json:"create_at" xorm:"not null default 0 comment('插入时间戳') INT(11)"`
	UpdateAt       time.Time `json:"update_at" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
	ContrastTime   int       `json:"contrast_time" xorm:"not null default 0 comment('对账时间') INT(11)"`
}
