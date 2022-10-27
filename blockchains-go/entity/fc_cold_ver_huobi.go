package entity

import (
	"time"
)

type FcColdVerHuobi struct {
	Id             int       `json:"id" xorm:"not null pk autoincr comment('自增ID') INT(11)"`
	AccountId      int       `json:"account_id" xorm:"not null default 0 comment('账户ID') unique(day_account_id_currency) INT(11)"`
	AccountType    string    `json:"account_type" xorm:"not null default '' comment('账户类型，spot：现货账户') unique(day_account_id_currency) VARCHAR(15)"`
	State          string    `json:"state" xorm:"not null default '' comment('账户状态，working：正常 lock：账户被锁定') unique(day_account_id_currency) VARCHAR(15)"`
	Currency       string    `json:"currency" xorm:"not null default '' comment('币种英文名称') unique(day_account_id_currency) VARCHAR(15)"`
	Type           string    `json:"type" xorm:"not null default '' comment('类型，trade: 交易余额，frozen: 冻结余额') unique(day_account_id_currency) VARCHAR(15)"`
	Balance        string    `json:"balance" xorm:"not null default 0.00000000000000000000 comment('余额') DECIMAL(65,20)"`
	Createdate     time.Time `json:"createdate" xorm:"not null default '0000-00-00' comment('创建时间') DATE"`
	CreateDatetime time.Time `json:"create_datetime" xorm:"not null default '0000-00-00 00:00:00' comment('抓取的日期时间, 一日多次对账新增的字段') index unique(day_account_id_currency) DATETIME"`
	Lastmodify     time.Time `json:"lastmodify" xorm:"not null default CURRENT_TIMESTAMP comment('最后修改时间') TIMESTAMP"`
}
