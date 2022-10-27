package entity

import (
	"time"
)

type FcRecFinancial struct {
	Id           int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	Type         int       `json:"type" xorm:"not null comment('1范奈斯2虎符') unique(code) TINYINT(4)"`
	Code         string    `json:"code" xorm:"not null unique(code) VARCHAR(50)"`
	Addtime      int       `json:"addtime" xorm:"not null comment('添加时间') INT(11)"`
	ColdNum      string    `json:"cold_num" xorm:"not null default 0.0000000000 comment('冷钱包余额') DECIMAL(20,10)"`
	AppNum       string    `json:"app_num" xorm:"not null default 0.0000000000 comment('热钱包余额') DECIMAL(20,10)"`
	FeeNum       string    `json:"fee_num" xorm:"not null default 0.0000000000 comment('手续费') DECIMAL(20,10)"`
	HuobiHistory string    `json:"huobi_history" xorm:"not null default 0.0000000000 comment('火币期初余额') DECIMAL(28,10)"`
	HuobiCurrent string    `json:"huobi_current" xorm:"not null default 0.0000000000 comment('火币期末余额') DECIMAL(28,10)"`
	UserNum      string    `json:"user_num" xorm:"not null default 0.0000000000 comment('用户余额') DECIMAL(20,10)"`
	IncNum       string    `json:"inc_num" xorm:"not null default 0.0000000000 comment('币+余额') DECIMAL(20,10)"`
	IncFee       string    `json:"inc_fee" xorm:"not null default 0.0000000000 comment('币+利息') DECIMAL(20,10)"`
	GiftNum      string    `json:"gift_num" xorm:"not null default 0.0000000000 comment('礼品卡余额') DECIMAL(20,10)"`
	ConTime      time.Time `json:"con_time" xorm:"not null comment('对帐时间') DATE"`
	Content      string    `json:"content" xorm:"not null default '' comment('备注, 废弃字段') VARCHAR(15)"`
	CoinName     string    `json:"coin_name" xorm:"not null comment('币种英文名称') unique(code) VARCHAR(11)"`
}
