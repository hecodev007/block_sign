package entity

import (
	"time"
)

type FcStructureTransferFollow struct {
	Id          int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	ApplyCoinId int       `json:"apply_coin_id" xorm:"not null comment('申请id') index(apply_follow_id) INT(11)"`
	TransferId  int       `json:"transfer_id" xorm:"not null comment('交易id') index(apply_follow_id) INT(11)"`
	CoinName    string    `json:"coin_name" xorm:"not null comment('币种名称') VARCHAR(15)"`
	Type        string    `json:"type" xorm:"not null comment('utxo:消耗的utxo,structure:构建数据create:构建结果sign:签名参数push:广播参数') ENUM('create','push','sign','structure','utxo')"`
	JsonData    string    `json:"json_data" xorm:"not null comment('返回值') TEXT"`
	Status      int       `json:"status" xorm:"not null default 0 comment('0正在进行的程序批次1执行完的批次程序') TINYINT(2)"`
	Createtime  int       `json:"createtime" xorm:"not null comment('创建时间') INT(11)"`
	Lastmodify  time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
	IsDing      int       `json:"is_ding" xorm:"not null default 0 TINYINT(1)"`
}
