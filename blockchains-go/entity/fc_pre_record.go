package entity

import (
	"time"
)

type FcPreRecord struct {
	Id        int64     `json:"id" xorm:"pk autoincr BIGINT(20)"`
	CoinName  string    `json:"coin_name" xorm:"not null VARCHAR(20)"`
	Timestamp int       `json:"timestamp" xorm:"INT(10)"`
	TxId      string    `json:"tx_id" xorm:"index VARCHAR(100)"`
	JsonData  string    `json:"json_data" xorm:"LONGTEXT"`
	AddTime   int       `json:"add_time" xorm:"INT(10)"`
	UpdateAt  time.Time `json:"update_at" xorm:"not null default CURRENT_TIMESTAMP DATETIME"`
}
