package entity

import (
	"time"
)

type FcFncIntst struct {
	Id          int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Uid         int       `json:"uid" xorm:"comment('用户ID') INT(11)"`
	Coin        string    `json:"coin" xorm:"not null default '0' VARCHAR(20)"`
	DealDate    time.Time `json:"deal_date" xorm:"comment('计息日期') DATE"`
	CreateTime  time.Time `json:"create_time" xorm:"comment('利息生成时间') DATETIME"`
	BaseAmount  string    `json:"base_amount" xorm:"default 0.0000000000 comment('本金') DECIMAL(23,10)"`
	Rate        string    `json:"rate" xorm:"default 0.0000000000 comment('利率') DECIMAL(23,10)"`
	Interest    string    `json:"interest" xorm:"default 0.0000000000 comment('利息') DECIMAL(23,10)"`
	InterestSum string    `json:"interest_sum" xorm:"default 0.0000000000 comment('利息累计') DECIMAL(23,10)"`
}
