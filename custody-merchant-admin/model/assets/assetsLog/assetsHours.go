package assetsLog

import (
	"github.com/shopspring/decimal"
	"time"
)

type AssetsHours struct {
	ServiceId    int             `gorm:"column:service_id" json:"service_id"`
	CoinId       int             `gorm:"column:coin_id" json:"coin_id"`
	Id           int64           `json:"id" gorm:"column:id; PRIMARY_KEY"`
	CoinName     string          `gorm:"column:coin_name" json:"coin_name"`
	ChainAddress string          `gorm:"column:chain_address" json:"chain_address,omitempty"`
	Nums         decimal.Decimal `gorm:"column:nums" json:"nums"`
	Valuation    decimal.Decimal `gorm:"column:valuation" json:"valuation"`
	Freeze       decimal.Decimal `gorm:"column:freeze" json:"freeze,omitempty"`
	CreateTime   time.Time       `gorm:"column:create_time" json:"create_time"`
}

type AsHours struct {
	CreateTime string          `gorm:"column:create_time" json:"create_time"`
	Nums       decimal.Decimal `gorm:"column:nums" json:"nums"`
	Valuation  decimal.Decimal `gorm:"column:valuation" json:"valuation"`
	Freeze     decimal.Decimal `gorm:"column:freeze" json:"freeze,omitempty"`
}

func (u *AssetsHours) TableName() string {
	return "assets_hours"
}
