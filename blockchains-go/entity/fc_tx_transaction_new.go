package entity

import (
	"time"
)

type FcTxTransactionNew struct {
	AgentId        int       `json:"agent_id" xorm:"not null default 0 comment('代理商id') INT(11)"`
	AgentProfit    int       `json:"agent_profit" xorm:"not null default 0 comment('代理商收益') INT(11)"`
	Amount         string    `json:"amount" xorm:"default 0.000000000000000000 DECIMAL(60,24)"`
	BlockHeight    int64     `json:"block_height" xorm:"not null default 0 INT(11)"`
	Coin           string    `json:"coin" xorm:"not null comment('主链币名称') VARCHAR(25)"`
	CoinType       string    `json:"coin_type" xorm:"not null comment('币种名称') VARCHAR(15)"`
	Confirmations  int64     `json:"confirmations" xorm:"not null default 0 comment('确认数') INT(10)"`
	ContrastTime   int64     `json:"contrast_time" xorm:"not null default 0 comment('对账时间') INT(11)"`
	CreateAt       int64     `json:"create_at" xorm:"not null default 0 comment('插入时间戳') INT(11)"`
	FromAddress    string    `json:"from_address" xorm:"comment('出账地址') VARCHAR(255)"`
	Id             int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	IsVerify       int       `json:"is_verify" xorm:"not null default 0 comment('是否为验证过的交易，0：未验证，1：交易所请求验证成功，2：公链云推送验证成功，-1商户信息错误，其他：验证失败及次数') TINYINT(4)"`
	MchId          int       `json:"mch_id" xorm:"not null default 0 INT(11)"`
	MchServiceCoin string    `json:"mch_service_coin" xorm:"comment('服务费币种') VARCHAR(10)"`
	MchServiceFee  int       `json:"mch_service_fee" xorm:"not null default 0 comment('商户服务收费') INT(11)"`
	Memo           string    `json:"memo" xorm:"comment('链上交易备注') VARCHAR(255)"`
	NotifyStatus   int       `json:"notify_status" xorm:"not null default 0 comment('通知状态，是否已通知商户') TINYINT(3)"`
	OrderNo        string    `json:"order_no" xorm:"not null comment('内部订单号') VARCHAR(255)"`
	OuterOrderNo   string    `json:"outer_order_no" xorm:"not null comment('外部订单号') VARCHAR(255)"`
	PlatformProfit int       `json:"platform_profit" xorm:"not null default 0 comment('平台收益') INT(11)"`
	Remark         string    `json:"remark" xorm:"comment('备注') VARCHAR(255)"`
	SettleId       int       `json:"settle_id" xorm:"not null default 0 comment('服务费清分规则id') INT(11)"`
	Status         int       `json:"status" xorm:"not null default 1 comment('交易状态 1 正常  2 作废  3 虚拟补单') TINYINT(3)"`
	Timestamp      int64     `json:"timestamp" xorm:"not null default 0 comment('交易时间戳') INT(11)"`
	ToAddress      string    `json:"to_address" xorm:"comment('入账地址') VARCHAR(255)"`
	TxFee          string    `json:"tx_fee" xorm:"DECIMAL(60,24)"`
	TxFeeCoin      string    `json:"tx_fee_coin" xorm:"not null comment('矿工费币种') VARCHAR(10)"`
	TxId           string    `json:"tx_id" xorm:"not null comment('txid') VARCHAR(100)"`
	TxType         int       `json:"tx_type" xorm:"not null default 0 comment('1 入账 2 出账 3 归集入账 4 归集出账 5 手续费入账 6 手续费出账 7 双币链btc入账，8双币链btc出账') TINYINT(3)"`
	TxcDetailId    int       `json:"txc_detail_id" xorm:"not null default 0 comment('清分明细表id') INT(11)"`
	TxcId          int       `json:"txc_id" xorm:"not null default 0 comment('清分表id') INT(11)"`
	UpdateAt       time.Time `json:"update_at" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
	TxN            int       `json:"tx_n" xorm:"not null default 0 comment('交易序号') INT(11)"`
	SeqNo          string    `json:"seq_no" xorm:"comment('流水号') VARCHAR(128)"`
}
