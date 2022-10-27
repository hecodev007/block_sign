package entity

import (
	"time"
)

type FcTransfersApplyCoin struct {
	Id         int       `json:"id" xorm:"not null pk autoincr comment('自增ID') INT(11)"`
	ApplyId    int       `json:"apply_id" xorm:"not null comment('申请ID，关联apply表的ID') index INT(11)"`
	CoinName   string    `json:"coin_name" xorm:"not null comment('币种名称') VARCHAR(16)"`
	ToAmount   string    `json:"to_amount" xorm:"not null comment('接收金额') DECIMAL(50,20)"`
	OrderId    string    `json:"order_id" xorm:"not null default '' comment('订单ID, 跟业务对接用') VARCHAR(100)"`
	Content    string    `json:"content" xorm:"not null default '' comment('备注，扩展字段') VARCHAR(64)"`
	Status     int       `json:"status" xorm:"not null default 0 comment('0未执行1执行中2部分成功3全部成功4全部失败(字段暂时未用)') TINYINT(3)"`
	Createtime int       `json:"createtime" xorm:"not null comment('创建时间') INT(11)"`
	Lastmodify time.Time `json:"lastmodify" xorm:"default CURRENT_TIMESTAMP comment('最后更新时间') TIMESTAMP"`
	Memo       string    `json:"memo" xorm:"not null default '' comment('eos memo') VARCHAR(100)"`
	Eostoken   string    `json:"eostoken" xorm:"not null default '' comment('eostoken') VARCHAR(50)"`
	Eoskey     string    `json:"eoskey" xorm:"not null default '' comment('eoskey') VARCHAR(50)"`
}
