package entity

import (
	"time"
)

type FcRecHooStatsData struct {
	Id         int       `json:"id" xorm:"not null pk autoincr comment('自增ID') INT(10)"`
	CoinName   string    `json:"coin_name" xorm:"not null comment('币种英文名称') unique(coin_type_day) VARCHAR(15)"`
	Type       string    `json:"type" xorm:"not null comment('gift-礼品卡,plus-币+,HOO接口-用户余额,wallet_cold-冷钱包,ram-内存,wallet_app用户热钱包余额') unique(coin_type_day) ENUM('balance','gift','hubi_balance','plus','ram','wallet_app','wallet_cold')"`
	Amount     string    `json:"amount" xorm:"not null comment('合计数量') DECIMAL(65,20)"`
	OtherNum   string    `json:"other_num" xorm:"not null default 0.00000000000000000000 comment('冻结，手续费等数据') DECIMAL(65,20)"`
	StatsDate  time.Time `json:"stats_date" xorm:"not null default '0000-00-00' comment('统计日期') index DATE"`
	StatsTime  time.Time `json:"stats_time" xorm:"not null default '0000-00-00 00:00:00' comment('统计时间') DATETIME"`
	Lastmodify time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最近修改时间') TIMESTAMP"`
	DoDate     time.Time `json:"do_date" xorm:"not null default '0000-00-00 00:00:00' unique(coin_type_day) DATETIME"`
	DoTime     int       `json:"do_time" xorm:"not null INT(11)"`
}
