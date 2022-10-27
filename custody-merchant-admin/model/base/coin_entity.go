package base

import (
	"github.com/shopspring/decimal"
	"time"
)

type CoinInfo struct {
	Id         int64           `json:"id" gorm:"column:id; PRIMARY_KEY"`
	State      int             `json:"state" gorm:"column:state"`
	ChainId    int             `json:"chain_id"  gorm:"column:chain_id"`
	Confirm    int64           `json:"confirm"  gorm:"column:confirm"`
	Name       string          `json:"name"  gorm:"column:name"`
	Token      string          `json:"token"  gorm:"column:token"`
	PriceUsd   decimal.Decimal `json:"price_usd"  gorm:"column:price_usd"`
	FullName   string          `json:"full_name" gorm:"column:full_name"`
	CreateTime time.Time       `json:"create_time"  gorm:"column:create_time"`
	UpdateTime time.Time       `json:"update_time"  gorm:"column:update_time"`
	DeletedAt  time.Time       `json:"deleted_at"  gorm:"column:deleted_at"`
}

func (u *CoinInfo) TableName() string {
	return "coin_info"
}
