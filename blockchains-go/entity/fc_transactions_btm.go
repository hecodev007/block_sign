package entity

import (
	"time"
)

type FcTransactionsBtm struct {
	Id                  int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	CoinId              int       `json:"coin_id" xorm:"not null comment('币种ID') INT(11)"`
	Address             string    `json:"address" xorm:"not null comment('地址') index unique(idx_unique) VARCHAR(100)"`
	Amount              string    `json:"amount" xorm:"not null default 0.0000000000 comment('交易量（转出为负数）') DECIMAL(40,10)"`
	Fee                 string    `json:"fee" xorm:"not null default 0.0000000000 comment('手续费（负数）') DECIMAL(40,10)"`
	DoTime              int       `json:"do_time" xorm:"not null default 0 comment('交易时间戳') INT(11)"`
	DoDate              time.Time `json:"do_date" xorm:"comment('交易日期') DATE"`
	Block               int64     `json:"block" xorm:"not null default 0 comment('所在区块高度') BIGINT(20)"`
	Txid                string    `json:"txid" xorm:"not null comment('交易ID(唯一键)') unique(idx_unique) VARCHAR(100)"`
	Flag                int       `json:"flag" xorm:"not null default 0 comment('1转入，2转出') unique(idx_unique) TINYINT(4)"`
	Type                int       `json:"type" xorm:"not null default 0 comment('1冷钱包平台地址2热钱包3普通用户地址') TINYINT(3)"`
	FeeType             int       `json:"fee_type" xorm:"not null default 0 comment('1冷钱包转热钱包2热钱包转热钱包') TINYINT(4)"`
	AddressType         int       `json:"address_type" xorm:"not null default 0 comment('0ETH_erc20地址1比原币地址') TINYINT(4)"`
	BlockHash           string    `json:"block_hash" xorm:"not null default '' VARCHAR(150)"`
	IsSpend             int       `json:"is_spend" xorm:"not null default 0 comment('1已花费0未花费') TINYINT(4)"`
	Utxo                string    `json:"utxo" xorm:"unique(idx_unique) VARCHAR(100)"`
	OppositeAddressType int       `json:"opposite_address_type" xorm:"not null default 0 comment('对方地址类型，0-未知或外部地址；1-冷钱包地址，2-热钱包地址，3-用户地址') TINYINT(2)"`
	Lastmodify          time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
	MuxId               string    `json:"mux_id" xorm:"not null default '' VARCHAR(100)"`
	N                   int       `json:"n" xorm:"not null default 0 INT(11)"`
}
