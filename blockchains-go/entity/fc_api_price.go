package entity

import (
	"time"
)

type FcApiPrice struct {
	Id         int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	CoinId     int       `json:"coin_id" xorm:"comment('币种id') INT(11)"`
	CoinName   string    `json:"coin_name" xorm:"not null comment('币种名称') VARCHAR(15)"`
	Price      string    `json:"price" xorm:"not null comment('服务费用') DECIMAL(20,10)"`
	Rate       string    `json:"rate" xorm:"not null comment('流水费率') DECIMAL(20,10)"`
	Term       int       `json:"term" xorm:"not null default 0 comment('0表示不限制，其他表示年限') INT(11)"`
	Createtime int       `json:"createtime" xorm:"not null default 0 comment('创建时间') INT(11)"`
	Lastmodify time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
	Status     int       `json:"status" xorm:"not null default 0 comment('是否启用：0不启用，1启用') TINYINT(4)"`
}
