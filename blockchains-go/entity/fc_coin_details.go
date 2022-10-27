package entity

import (
	"time"
)

type FcCoinDetails struct {
	Id           int       `json:"id" xorm:"not null pk INT(11)"`
	AppId        int       `json:"app_id" xorm:"not null index(app_id) INT(11)"`
	Address      string    `json:"address" xorm:"not null index(app_id) VARCHAR(255)"`
	CoinId       int       `json:"coin_id" xorm:"not null index(app_id) INT(11)"`
	Amount       string    `json:"amount" xorm:"not null default 0.0000000000 comment('改变数量') DECIMAL(23,10)"`
	Type         int       `json:"type" xorm:"not null comment('0转入1转出') TINYINT(1)"`
	Addtime      string    `json:"addtime" xorm:"not null VARCHAR(50)"`
	Status       int       `json:"status" xorm:"comment('0:未确认,2：正在确认,1:已确认') TINYINT(1)"`
	Confirm      int       `json:"confirm" xorm:"INT(11)"`
	Block        int       `json:"block" xorm:"INT(11)"`
	Txid         int       `json:"txid" xorm:"INT(11)"`
	Fee          string    `json:"fee" xorm:"not null default 0.0000000000 DECIMAL(23,10)"`
	Updatetime   time.Time `json:"updatetime" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
	CurrentPrice string    `json:"current_price" xorm:"not null default 0.0000000000 DECIMAL(23,10)"`
	ActualAmount string    `json:"actual_amount" xorm:"not null default 0.0000000000 DECIMAL(23,10)"`
}
