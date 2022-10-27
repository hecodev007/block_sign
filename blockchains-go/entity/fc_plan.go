package entity

import (
	"time"
)

type FcPlan struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Address    string    `json:"address" xorm:"comment('转出地址') VARCHAR(255)"`
	CoinId     int       `json:"coin_id" xorm:"INT(11)"`
	Num        string    `json:"num" xorm:"not null default 0.0000000000 DECIMAL(23,10)"`
	Content    string    `json:"content" xorm:"VARCHAR(255)"`
	Status     int       `json:"status" xorm:"comment('0:未确认,2：正在确认,1:已确认') TINYINT(1)"`
	Fee        string    `json:"fee" xorm:"not null default 0.0000000000 DECIMAL(23,10)"`
	Nickname   string    `json:"nickname" xorm:"VARCHAR(255)"`
	Addtime    time.Time `json:"addtime" xorm:"DATETIME"`
	Updatetime time.Time `json:"updatetime" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
	Txid       int       `json:"txid" xorm:"INT(11)"`
}
