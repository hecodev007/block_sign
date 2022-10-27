package domain

import (
	"github.com/shopspring/decimal"
)

type SearchIncome struct {
	Id         int64  `json:"id"`
	Account    string `json:"account"`
	ComboId    int    `json:"combo_id"`
	MerchantId int64  `json:"merchant_id"`
	ServiceId  int    `json:"service_id"`
	CoinId     int    `json:"coin_id"`
	ChainId    int    `json:"chain_id"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	Limit      int    `json:"limit"  description:"查询条数" example:"10"`
	Offset     int    `json:"offset" description:"查询起始位置" example:"0"`
}

type IncomeList struct {
	Totals   decimal.Decimal `json:"totals"`
	TopUp    decimal.Decimal `json:"top_up"`
	Withdraw decimal.Decimal `json:"withdraw"`
	Combo    decimal.Decimal `json:"combo"`
	TopList  []TopInfo       `json:"top_list"`
	List     []IncomeInfo    `json:"list"`
}

type TopInfo struct {
	CoinName string          `json:"coin_name"`
	CoinId   int             `json:"coin_id"`
	Price    decimal.Decimal `json:"price"`
}
type IncomeStateList struct {
	IncomeName  string          `json:"income_name"`
	IncomePrice decimal.Decimal `json:"income_price"`
	Color       string          `json:"color"`
}

type IncomeInfo struct {
	Serial          int64           `json:"serial"`
	MerchantId      int64           `json:"merchant_id" gorm:"column:merchant_id"`
	UserName        string          `json:"user_name" gorm:"column:user_name"`
	UserPhone       string          `json:"user_phone" gorm:"column:user_phone"`
	UserEmail       string          `json:"user_email" gorm:"column:user_email"`
	ServiceId       int             `json:"service_id,omitempty" gorm:"column:service_id"`
	ServiceName     string          `json:"service_name,omitempty" gorm:"column:service_name"`
	ComboId         int             `json:"combo_id,omitempty" gorm:"column:combo_id"`
	ComboName       string          `json:"combo_name,omitempty" gorm:"column:combo_name"`
	ComboTypeName   string          `json:"combo_type_name" gorm:"column:combo_type_name"`
	ComboModelName  string          `json:"combo_model_name" gorm:"column:combo_model_name"`
	CoinId          int             `json:"coin_id,omitempty" gorm:"column:coin_id"`
	CoinName        string          `json:"coin_name,omitempty" gorm:"column:coin_name"`
	ChainName       string          `json:"chain_name,omitempty" gorm:"column:chain_name"`
	TopUpNums       int             `json:"top_up_nums,omitempty" gorm:"column:top_up_nums"`
	TopUpPrice      decimal.Decimal `json:"top_up_price,omitempty" gorm:"column:top_up_price"`
	ToUpDestroy     decimal.Decimal `json:"top_up_destroy,omitempty" gorm:"column:top_up_destroy"`
	ToUpFee         decimal.Decimal `json:"top_up_fee,omitempty" gorm:"column:top_up_fee"`
	WithdrawNums    int             `json:"withdraw_nums,omitempty" gorm:"column:withdraw_nums"`
	WithdrawPrice   decimal.Decimal `json:"withdraw_price,omitempty" gorm:"column:withdraw_price"`
	WithdrawFee     decimal.Decimal `json:"withdraw_fee,omitempty" gorm:"column:withdraw_fee"`
	WithdrawDestroy decimal.Decimal `json:"withdraw_destroy,omitempty" gorm:"column:withdraw_destroy"`
	MinerFee        decimal.Decimal `json:"miner_fee,omitempty" gorm:"column:miner_fee"`
	TopUpIncome     decimal.Decimal `json:"top_up_income,omitempty" gorm:"column:top_up_income"`
	WithdrawIncome  decimal.Decimal `json:"withdraw_income,omitempty" gorm:"column:withdraw_income"`
	ComboIncome     decimal.Decimal `json:"combo_income,omitempty" gorm:"column:combo_income"`
	TotalIncome     decimal.Decimal `json:"total_income,omitempty" gorm:"column:total_income"`
}
