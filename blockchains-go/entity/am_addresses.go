package entity

import "time"

type Addresses struct {
	Id             int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Address        string    `json:"address" xorm:"varchar(255)"`
	CoinType       string    `json:"coin_type" xorm:"varchar(20)"`
	AddrType       string    `json:"addr_type" xorm:"varchar(20)"`
	Status         string    `json:"status" xorm:"varchar(20)"`
	ComeFrom       string    `json:"come_from" xorm:"varchar(20)"`
	UserId         int       `json:"user_id" xorm:"INT(11)"`
	RegisterPushed int       `json:"register_pushed" xorm:"INT(11)"`
	DeletePushed   int       `json:"delete_pushed" xorm:"INT(11)"`
	CreatedAt      time.Time `json:"created_at" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
	UpdatedAt      time.Time `json:"updated_at" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
	DeletedAt      time.Time `json:"deleted_at" xorm:"default CURRENT_TIMESTAMP TIMESTAMP"`
}
