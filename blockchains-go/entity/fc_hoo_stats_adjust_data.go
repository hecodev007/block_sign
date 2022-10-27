package entity

import (
	"time"
)

type FcHooStatsAdjustData struct {
	Id         int       `json:"id" xorm:"not null pk autoincr comment('自增ID') INT(11)"`
	ItemType   string    `json:"item_type" xorm:"not null comment('添加数据类型') index(coin_name_type) VARCHAR(32)"`
	CoinName   string    `json:"coin_name" xorm:"not null comment('币种英文名称') index(coin_name_type) VARCHAR(20)"`
	ConDate    time.Time `json:"con_date" xorm:"not null comment('作用日期') index DATETIME"`
	AdjustData string    `json:"adjust_data" xorm:"not null comment('调整数值') DECIMAL(65,20)"`
	Content    string    `json:"content" xorm:"not null comment('备注') VARCHAR(2000)"`
	Addtime    int       `json:"addtime" xorm:"not null comment('创建时间') INT(11)"`
	Lastmodify time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最近修改时间') TIMESTAMP"`
}
