package entity

import (
	"time"
)

type FcFeeSupplierBtc struct {
	Id              int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	Platform        string    `json:"platform" xorm:"comment('商户名称') unique(coin_name) VARCHAR(20)"`
	CoinName        string    `json:"coin_name" xorm:"not null comment('币种名称') unique(coin_name) VARCHAR(15)"`
	Address         string    `json:"address" xorm:"not null comment('转出手续费的地址') unique(coin_name) VARCHAR(120)"`
	FromConfineNum  string    `json:"from_confine_num" xorm:"not null default 0.0000000000 comment('当币种数量达到这个限制时，才进行转出') DECIMAL(40,10)"`
	FromTransferNum string    `json:"from_transfer_num" xorm:"not null default 0.0000000000 comment('转出的数量') DECIMAL(40,10)"`
	ToConfineNum    string    `json:"to_confine_num" xorm:"not null default 0.0000000000 comment('当币种数量达到这个限制时，才进行转入') DECIMAL(40,10)"`
	ToTransferNum   string    `json:"to_transfer_num" xorm:"not null default 0.0000000000 comment('转入的额度') DECIMAL(40,10)"`
	Lastmodify      time.Time `json:"lastmodify" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
}
