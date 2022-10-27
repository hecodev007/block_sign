package atom

import (
	"time"

	"github.com/shopspring/decimal"
)

type FcAddressAmount struct {
	Id            int64           `json:"id" xorm:"pk autoincr BIGINT(20)"`
	CoinId        int             `json:"coin_id" xorm:"not null default 0 INT(10)"`
	CoinType      string          `json:"coin_type" xorm:"not null comment('币种名称') unique(coin_address) VARCHAR(100)"`
	Address       string          `json:"address" xorm:"not null comment('地址') unique(coin_address) VARCHAR(100)"`
	Amount        decimal.Decimal `json:"amount" xorm:"not null default 0.000000000000000000 comment('当前余额') DECIMAL(60,24)"`
	ForzenAmount  string          `json:"forzen_amount" xorm:"not null default 0.000000000000000000 comment('冻结金额') DECIMAL(60,24)"`
	Type          int             `json:"type" xorm:"not null default 0 comment('地址类型 1 冷地址 2 用户地址  3 手续费地址') TINYINT(3)"`
	AppId         int64           `json:"app_id" xorm:"not null default 0 comment('商户id') INT(10)"`
	PendingAmount string          `json:"pending_amount" xorm:"not null default 0.000000000000000000 comment('发送中的金额') DECIMAL(60,24)"`
}

type FcAddressAmount1 struct {
	FcAddressAmount
}

func (t *FcAddressAmount1) TableName() string {
	return "fc_address_amount_1"
}

type FcAddressAmount2 struct {
	FcAddressAmount
}

func (t *FcAddressAmount2) TableName() string {
	return "fc_address_amount_2"
}

type FcTxClearDetail struct {
	Id       int             `json:"id" xorm:"not null pk autoincr INT(11)"`
	TxcId    int             `json:"txc_id" xorm:"not null default 0 comment('交易清分') INT(11)"`
	AgentId  int             `json:"agent_id" xorm:"not null default 0 comment('代理商ID') INT(11)"`
	MchId    int             `json:"mch_id" xorm:"not null default 0 comment('商户ID') INT(11)"`
	CoinType string          `json:"coin_type" xorm:"not null comment('币种名称') unique(trans_check) VARCHAR(15)"`
	TxId     string          `json:"tx_id" xorm:"not null unique(trans_check) VARCHAR(150)"`
	Hash     string          `json:"hash" xorm:"not null comment('块hash') VARCHAR(100)"`
	Dir      int             `json:"dir" xorm:"not null comment('1 入账  2 出账') unique(trans_check) TINYINT(3)"`
	MuxId    string          `json:"mux_id" xorm:"not null comment('btm签名用') VARCHAR(150)"`
	TxN      int             `json:"tx_n" xorm:"not null comment('地址位于交易位置') unique(trans_check) INT(11)"`
	Addr     string          `json:"addr" xorm:"not null comment('地址') unique(trans_check) VARCHAR(100)"`
	Amount   decimal.Decimal `json:"amount" xorm:"not null default 0.000000000000000000 comment('金额') unique(trans_check) DECIMAL(60,24)"`
	AddrType int             `json:"addr_type" xorm:"not null comment('1 用户地址 2 商户公账地址 3 垫资地址 4 商户外部公账地址 5 其他地址') TINYINT(4)"`
	IsSpent  int             `json:"is_spent" xorm:"not null default 0 comment('0 未花费  1 已花费') TINYINT(4)"`
	OrderNo  string          `json:"order_no" xorm:"comment('订单号') VARCHAR(60)"`
	FromTxId string          `json:"from_tx_id" xorm:"comment('花费的tx_id') unique(trans_check) VARCHAR(100)"`
	VoutId   string          `json:"vout_id" xorm:"not null comment('唯一标识btm utxo') VARCHAR(100)"`
	Status   int             `json:"status" xorm:"not null default 1 comment('1 正常  2 作废  3 虚拟地址') TINYINT(4)"`
	IsOver   int             `json:"is_over" xorm:"not null default 0 comment('是否已处理') TINYINT(3)"`
	Remark   string          `json:"remark" xorm:"comment('备注') VARCHAR(255)"`
	CreateAt int             `json:"create_at" xorm:"not null default 0 INT(11)"`
	UpdateAt time.Time       `json:"update_at" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP"`
}

func (t *FcTxClearDetail) TableName() string {
	return "fc_tx_clear_detail"
}
