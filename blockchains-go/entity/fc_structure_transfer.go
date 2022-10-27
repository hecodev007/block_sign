package entity

import (
	"time"
)

type FcStructureTransfer struct {
	Id          int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	ApplyId     int       `json:"apply_id" xorm:"not null comment('申请id') index(apply_id) INT(11)"`
	ApplyCoinId int       `json:"apply_coin_id" xorm:"not null comment('申请币种信息id') index(apply_id) INT(11)"`
	OrderCodeId string    `json:"order_code_id" xorm:"comment('order_id下面交易的编号') VARCHAR(120)"`
	CoinName    string    `json:"coin_name" xorm:"not null comment('币种') VARCHAR(15)"`
	Content     string    `json:"content" xorm:"not null default '' comment('说明') VARCHAR(50)"`
	ErrorSum    int       `json:"error_sum" xorm:"not null default 0 comment('广播失败次数') INT(11)"`
	Status      int       `json:"status" xorm:"not null comment('0:构建完成1:构建失败2:签名成功3:签名失败4:签名失败待重试5:广播成功6:广播失败7:广播失败待重试') TINYINT(2)"`
	IsRetry     int       `json:"is_retry" xorm:"not null default 0 comment('0非重试1重试') TINYINT(2)"`
	Createtime  int       `json:"createtime" xorm:"not null INT(11)"`
	Lastmodify  time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
	IsDing      int       `json:"is_ding" xorm:"not null default 0 TINYINT(1)"`
}
