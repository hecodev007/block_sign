package entity

import (
	"time"
)

type FcDeposit struct {
	Id           int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	AddressId    int       `json:"address_id" xorm:"index INT(11)"`
	CoinId       int       `json:"coin_id" xorm:"INT(11)"`
	Txid         string    `json:"txid" xorm:"VARCHAR(200)"`
	Amount       string    `json:"amount" xorm:"not null default 0.0000000000 DECIMAL(23,10)"`
	Fee          string    `json:"fee" xorm:"not null default 0.0000000000 DECIMAL(23,10)"`
	ActualAmount string    `json:"actual_amount" xorm:"not null default 0.0000000000 DECIMAL(23,10)"`
	Updatetime   time.Time `json:"updatetime" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
	Addtime      int64     `json:"addtime" xorm:"BIGINT(20)"`
	Status       int       `json:"status" xorm:"default 0 comment('0:未确认,2：正在确认,1:已确认') TINYINT(4)"`
	Confirm      int       `json:"confirm" xorm:"comment('确认数') INT(11)"`
	Address      string    `json:"address" xorm:"index VARCHAR(255)"`
	Block        int       `json:"block" xorm:"INT(11)"`
	CheckStatus  int       `json:"check_status" xorm:"default 0 comment('是否已使用') TINYINT(4)"`
}
