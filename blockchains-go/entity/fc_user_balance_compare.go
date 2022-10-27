package entity

import (
	"time"
)

type FcUserBalanceCompare struct {
	Id          int       `json:"id" xorm:"not null pk autoincr INT(11)"`
	CoinId      int       `json:"coin_id" xorm:"not null unique(idx_coinid_address) INT(11)"`
	Address     string    `json:"address" xorm:"not null index unique(idx_coinid_address) VARCHAR(80)"`
	CoinNum     string    `json:"coin_num" xorm:"not null default 0.0000000000 comment('币种数量') DECIMAL(23,10)"`
	WaitNum     string    `json:"wait_num" xorm:"not null default 0.0000000000 comment('冻结金额') DECIMAL(23,10)"`
	SharesSum   string    `json:"shares_sum" xorm:"not null default 0.0000000000 comment('升值钱包') DECIMAL(23,10)"`
	InterestSum string    `json:"interest_sum" xorm:"not null default 0.0000000000 comment('利息数') DECIMAL(23,10)"`
	GiftSum     string    `json:"gift_sum" xorm:"not null default 0.0000000000 comment('礼品卡数量') DECIMAL(23,10)"`
	Addtime     int64     `json:"addtime" xorm:"not null comment('添加时间') BIGINT(20)"`
	AppId       int       `json:"app_id" xorm:"unique(idx_coinid_address) INT(11)"`
	EntId       int       `json:"ent_id" xorm:"not null INT(11)"`
	ProcessDate time.Time `json:"process_date" xorm:"not null default '0000-00-00' comment('处理日期') unique(idx_coinid_address) index DATE"`
	Start       string    `json:"start" xorm:"not null default 0.000000 DECIMAL(18,6)"`
	End         string    `json:"end" xorm:"not null default 0.000000 DECIMAL(18,6)"`
}
