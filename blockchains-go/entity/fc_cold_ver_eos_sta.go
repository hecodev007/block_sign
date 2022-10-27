package entity

import (
	"time"
)

type FcColdVerEosSta struct {
	Id          int64     `json:"id" xorm:"pk autoincr BIGINT(11)"`
	AddressType int       `json:"address_type" xorm:"not null default 0 comment('地址类型,1-冷钱包账户;2-热钱包账户;3-用户账户；4-计划外账户;5-柚子银行') TINYINT(2)"`
	Address     string    `json:"address" xorm:"not null index unique(tdate_address) VARCHAR(80)"`
	Num         string    `json:"num" xorm:"not null default 0.0000000000 comment('余额') DECIMAL(28,10)"`
	Staked      string    `json:"staked" xorm:"not null default 0.00000000000000000000 comment('抵押额度') DECIMAL(65,20)"`
	Cpu         string    `json:"cpu" xorm:"not null default 0.0000000000 comment('cpu价值') DECIMAL(28,10)"`
	CpuRes      string    `json:"cpu_res" xorm:"not null default 0.0000000000 comment('实际cpu(ms)') DECIMAL(20,10)"`
	Ram         string    `json:"ram" xorm:"not null default 0.0000000000 comment('内存价值') DECIMAL(20,10)"`
	RamRes      string    `json:"ram_res" xorm:"not null default 0.0000000000 comment('实际内存(byte)') DECIMAL(20,10)"`
	Net         string    `json:"net" xorm:"not null default 0.0000000000 comment('带宽价值') DECIMAL(20,10)"`
	NetRes      string    `json:"net_res" xorm:"not null default 0.0000000000 comment('实际带宽(byte)') DECIMAL(20,10)"`
	TDate       time.Time `json:"t_date" xorm:"not null DATE"`
	Time        int       `json:"time" xorm:"not null INT(11)"`
	DoDate      time.Time `json:"do_date" xorm:"not null default '0000-00-00 00:00:00' unique(tdate_address) DATETIME"`
}
