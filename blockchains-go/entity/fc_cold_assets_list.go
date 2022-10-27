package entity

import (
	"time"
)

type FcColdAssetsList struct {
	Id              int       `json:"id" xorm:"not null pk autoincr comment('自增ID') INT(11)"`
	DateTime        time.Time `json:"date_time" xorm:"not null default '0000-00-00 00:00:00' comment('时间') unique(date_coin_address_type) DATETIME"`
	CoinName        string    `json:"coin_name" xorm:"not null comment('币种名称') unique(date_coin_address_type) VARCHAR(20)"`
	Balance         string    `json:"balance" xorm:"not null default 0.00000000000000000000 comment('余额') DECIMAL(65,20)"`
	Address         string    `json:"address" xorm:"not null default '' comment('余额地址') unique(date_coin_address_type) VARCHAR(255)"`
	Txid            string    `json:"txid" xorm:"not null default '' comment('交易ID') VARCHAR(128)"`
	TradeType       string    `json:"trade_type" xorm:"not null comment('出入账类型；in_recharge:充值入账,in_offplan:计划外入账,in_other:其他入账,out_to_hotwallet:转出到热钱包,out_fee:归集手续费,out_offplan:计划外出账,out_other:其他出账') unique(date_coin_address_type) ENUM('in_offplan','in_other','in_recharge','out_fee','out_offplan','out_other','out_to_hotwallet')"`
	InOutType       string    `json:"in_out_type" xorm:"not null comment('出入账类型, in:入账，out:出账') ENUM('in','out')"`
	Amount          string    `json:"amount" xorm:"not null default 0.00000000000000000000 comment('额度') DECIMAL(65,20)"`
	Fee             string    `json:"fee" xorm:"not null default 0.00000000000000000000 comment('手续费') DECIMAL(50,20)"`
	FeeUnit         string    `json:"fee_unit" xorm:"not null default '' comment('手续费单位') VARCHAR(20)"`
	AddressOpposite string    `json:"address_opposite" xorm:"not null default '' comment('对方地址') VARCHAR(255)"`
	Abstract        string    `json:"abstract" xorm:"not null default '' comment('摘要') VARCHAR(600)"`
	SourceType      string    `json:"source_type" xorm:"not null default 'robot' comment('该记录来源类型，manual-手动录入；robot-程序') ENUM('manual','robot')"`
	DoTime          int       `json:"do_time" xorm:"not null default 0 comment('时间戳') INT(11)"`
	Lastmodify      time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
}
