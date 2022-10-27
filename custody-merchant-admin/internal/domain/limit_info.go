package domain

import (
	"github.com/shopspring/decimal"
)

type LimitTransferInfo struct {
	CoinId       int             `json:"coin_id" description:"币种Id"`
	ServiceId    int             `json:"service_id"  description:"业务线Id"`
	CreateByUser int64           `json:"create_by_user"`
	NumEach      decimal.Decimal `json:"num_each" description:"每笔限额"`
	NumDay       decimal.Decimal `json:"num_day" description:"每日限额"`
	NumWeeks     decimal.Decimal `json:"num_weeks" description:"每周限额"`
	NumMonth     decimal.Decimal `json:"num_month" description:"每月限额"`
}

type LimitWithdrawal struct {
	ServiceId    int             `json:"service_id" description:"业务线Id"`
	NumMinutes   int             `json:"num_minutes" description:"每5分钟的提币次数"`
	NumHours     int             `json:"num_hours" description:"每小时的提币次数"`
	CreateByUser int64           `json:"create_by_user"`
	LineMinutes  decimal.Decimal `json:"line_minutes" description:"每5分钟的提币额度"`
	LineHours    decimal.Decimal `json:"line_hours" description:"每小时的提币额度"`
	Code         string          `json:"code" description:"短信验证码"`
}

type LimitStatus struct {
	Sids             []int `json:"sids"`
	ServiceId        int   `json:"service_id"`
	WithdrawalStatus int   `json:"withdrawal_status"`
	CreateByUser     int64 `json:"create_by_user"`
}

type WithdrawalList struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type OrderLimitNums struct {
	DNums decimal.Decimal
	WNums decimal.Decimal
	MNums decimal.Decimal
}
