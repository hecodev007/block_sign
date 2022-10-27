package entity

import (
	"time"
)

type FcFncIntstHis struct {
	Id              int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	Coin            string    `json:"coin" xorm:"not null default '0' unique(idx_coin_date) VARCHAR(20)"`
	LastDate        time.Time `json:"last_date" xorm:"not null comment('结算日期') unique(idx_coin_date) DATE"`
	Uid             int       `json:"uid" xorm:"not null default 0 comment('用户ID') INT(11)"`
	LastInterestSum string    `json:"last_interest_sum" xorm:"default 0.0000000000 comment('上一期的利息累计数') DECIMAL(23,10)"`
	Interest        string    `json:"interest" xorm:"not null default 0.0000000000 comment('本期利息总计') DECIMAL(23,10)"`
	InterestSum     string    `json:"interest_sum" xorm:"not null default 0.0000000000 comment('历史累积利息总计') DECIMAL(23,10)"`
	HooSum          string    `json:"hoo_sum" xorm:"default 0.0000000000 comment('虎符计息的利息累积数') DECIMAL(23,10)"`
}
