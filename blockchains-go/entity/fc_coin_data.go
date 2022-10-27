package entity

import (
	"time"
)

type FcCoinData struct {
	Id          int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	CoinId      int       `json:"coin_id" xorm:"INT(11)"`
	Name        string    `json:"name" xorm:"VARCHAR(20)"`
	BlockHeight int64     `json:"block_height" xorm:"default 1 BIGINT(20)"`
	Updatetime  time.Time `json:"updatetime" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
}
