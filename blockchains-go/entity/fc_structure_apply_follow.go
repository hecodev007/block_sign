package entity

import (
	"time"
)

type FcStructureApplyFollow struct {
	Id          int       `json:"id" xorm:"not null pk autoincr INT(10)"`
	ApplyId     int       `json:"apply_id" xorm:"index INT(11)"`
	CoinName    string    `json:"coin_name" xorm:"comment('币种名称') VARCHAR(255)"`
	FromAddress string    `json:"from_address" xorm:"comment('出账地址') TEXT"`
	ToAddress   string    `json:"to_address" xorm:"comment('接收地址') TEXT"`
	ToAmount    string    `json:"to_amount" xorm:"comment('接收金额') VARCHAR(255)"`
	Content     string    `json:"content" xorm:"comment('备注') TEXT"`
	Status      int       `json:"status" xorm:"default 0 comment('0未执行1执行中2部分成功3全部成功4全部失败') TINYINT(255)"`
	Createtime  int       `json:"createtime" xorm:"comment('创建时间') INT(11)"`
	Lastmodify  time.Time `json:"lastmodify" xorm:"default CURRENT_TIMESTAMP comment('最后更新时间') TIMESTAMP"`
}
