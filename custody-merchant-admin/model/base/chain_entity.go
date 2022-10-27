package base

import (
	"github.com/shopspring/decimal"
	"time"
)

type ChainInfo struct {
	Id         int             `json:"id" gorm:"column:id; PRIMARY_KEY"`
	State      int             `json:"state" gorm:"column:state"`
	Name       string          `json:"name" gorm:"column:name"`
	PriceUsd   decimal.Decimal `json:"price_usd"  gorm:"column:price_usd"`
	CreateTime time.Time       `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time       `json:"update_time" gorm:"column:update_time"`
}

func (u *ChainInfo) TableName() string {
	return "chain_info"
}
