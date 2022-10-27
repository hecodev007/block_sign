package entity

import (
	"time"
)

type FcRecFee struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	Type       int       `json:"type" xorm:"not null default 0 comment('1冷钱包出账手续费2热钱包转出手续费3hoo转出手续费4hoo即时兑换5push市场手续费统计6huobi交易手续费统计7ram交易费用8EOS小程序收入') unique(type) TINYINT(4)"`
	CoinName   string    `json:"coin_name" xorm:"not null unique(type) VARCHAR(15)"`
	Amount     string    `json:"amount" xorm:"not null DECIMAL(40,18)"`
	DData      time.Time `json:"d_data" xorm:"not null unique(type) DATETIME"`
	Createtime time.Time `json:"createtime" xorm:"not null default CURRENT_TIMESTAMP TIMESTAMP"`
	DTime      int       `json:"d_time" xorm:"not null INT(11)"`
}
