package entity

import (
	"time"
)

type FcRecAdjustData struct {
	Id         int       `json:"id" xorm:"not null pk autoincr comment('自增ID') INT(11)"`
	CoinName   string    `json:"coin_name" xorm:"not null comment('币种英文名称') index(coin_name_time) VARCHAR(20)"`
	ConTime    time.Time `json:"con_time" xorm:"not null comment('对账日期') index(coin_name_time) DATE"`
	AdjustData string    `json:"adjust_data" xorm:"not null comment('调整数值') DECIMAL(65,20)"`
	Content    string    `json:"content" xorm:"not null comment('调整原因') VARCHAR(2000)"`
	Addtime    int       `json:"addtime" xorm:"not null comment('创建时间') INT(11)"`
	Lastmodify time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最近修改时间') TIMESTAMP"`
}
