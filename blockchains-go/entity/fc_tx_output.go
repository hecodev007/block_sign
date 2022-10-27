package entity

type FcTxOutput struct {
	Id              int64  `json:"id" xorm:"pk autoincr BIGINT(20)"`
	CoinName        string `json:"coin_name" xorm:"not null comment('币种') VARCHAR(20)"`
	BlockHeight     int64  `json:"block_height" xorm:"not null default 0 comment('块高度') BIGINT(20)"`
	FromTxid        string `json:"from_txid" xorm:"not null comment('产出的交易ID') unique(txid_index) VARCHAR(150)"`
	TxIndex         int    `json:"tx_index" xorm:"not null default 0 comment('产出的交易内偏移量') unique(txid_index) INT(10)"`
	FromTxTimestamp int    `json:"from_tx_timestamp" xorm:"not null default 0 comment('产出交易时间戳') INT(11)"`
	FromAddress     string `json:"from_address" xorm:"not null comment('产出地址') VARCHAR(255)"`
	Address         string `json:"address" xorm:"not null comment('所属地址') index VARCHAR(255)"`
	ToAddress       string `json:"to_address" xorm:"not null comment('花费地址') VARCHAR(255)"`
	Value           int64  `json:"value" xorm:"not null default 0 comment('金额') BIGINT(20)"`
	Status          int    `json:"status" xorm:"not null default 0 comment('是否花费 0 未花费  1 冻结  2 已花费') TINYINT(3)"`
	OrderNo         string `json:"order_no" xorm:"comment('发起冻结的外部订单号') index VARCHAR(60)"`
	ToTxid          string `json:"to_txid" xorm:"comment('花费的交易ID') VARCHAR(100)"`
	ToTxTimestamp   int    `json:"to_tx_timestamp" xorm:"not null default 0 comment('花费交易时间戳') INT(11)"`
	AppId           int    `json:"app_id" xorm:"not null default 0 comment('商户ID') index SMALLINT(6)"`
	AddressType     int    `json:"address_type" xorm:"not null default 0 comment('地址类型 0 外部地址 1 归集地址（冷地址）  2 用户地址  3 手续费地址  4 热地址') index TINYINT(3)"`
	CreateAt        int64  `json:"create_at" xorm:"default 0 comment('创建时间') BIGINT(11)"`
	UpdateAt        int64  `json:"update_at" xorm:"default 0 comment('最后修改时间') BIGINT(11)"`
}
