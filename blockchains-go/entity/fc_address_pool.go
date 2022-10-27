package entity

import (
	"time"
)

type FcAddressPool struct {
	Id                int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	CoinId            int       `json:"coin_id" xorm:"not null unique(idx_coinid_address) INT(11)"`
	Address           string    `json:"address" xorm:"not null unique(idx_coinid_address) VARCHAR(255)"`
	Num               string    `json:"num" xorm:"not null default 0.0000000000 DECIMAL(23,10)"`
	Addtime           time.Time `json:"addtime" xorm:"not null default CURRENT_TIMESTAMP comment('添加时间') TIMESTAMP"`
	AppId             int       `json:"app_id" xorm:"INT(11)"`
	Status            int       `json:"status" xorm:"default 0 comment('是否分配：0未，1是') TINYINT(3)"`
	EntId             int       `json:"ent_id" xorm:"not null INT(11)"`
	ColdCode          int       `json:"cold_code" xorm:"comment('线下冷钱包编号') INT(11)"`
	BatchId           string    `json:"batch_id" xorm:"comment('批量id') VARCHAR(64)"`
	CompatibleAddress string    `json:"compatible_address" xorm:"not null default '' comment('兼容地址') VARCHAR(100)"`
	Key               string    `json:"key" xorm:"not null default '' VARCHAR(150)"`
}
