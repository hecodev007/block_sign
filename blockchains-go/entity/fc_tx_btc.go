package entity

import (
	"time"
)

type FcTxBtc struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	Platform   string    `json:"platform" xorm:"comment('商户名称') unique(platform) VARCHAR(20)"`
	CoinName   string    `json:"coin_name" xorm:"not null comment('币种名称') unique(platform) VARCHAR(15)"`
	Address    string    `json:"address" xorm:"not null comment('转出手续费的地址') unique VARCHAR(120)"`
	Level      int       `json:"level" xorm:"default 1 comment('级别，从1开始，往上归集') unique(platform) TINYINT(255)"`
	Amount     string    `json:"amount" xorm:"default 0.000000000000000000 comment('归集额度,0表示没有限制') DECIMAL(40,18)"`
	MaxCount   int       `json:"max_count" xorm:"default 0 comment('utxo数量，0等于无限数量') TINYINT(255)"`
	Status     int       `json:"status" xorm:"default 1 comment('0禁用1启用') TINYINT(255)"`
	Lastmodify time.Time `json:"lastmodify" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
}
