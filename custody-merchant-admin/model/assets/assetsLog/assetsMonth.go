package assetsLog

import (
	"github.com/shopspring/decimal"
	"time"
)

type AssetsMonth struct {
	CoinId       int             `gorm:"column:coin_id" json:"coin_id,omitempty"`
	ServiceId    int             `gorm:"column:service_id" json:"service_id,omitempty"`
	Id           int64           `json:"id" gorm:"column:id; PRIMARY_KEY"`
	Nums         decimal.Decimal `gorm:"column:nums" json:"nums,omitempty"`
	ChainAddress string          `gorm:"column:chain_address" json:"chain_address,omitempty"`
	Valuation    decimal.Decimal `gorm:"column:valuation" json:"valuation,omitempty"`
	Freeze       decimal.Decimal `gorm:"column:freeze" json:"freeze,omitempty"`
	CreateTime   time.Time       `gorm:"column:create_time" json:"create_time,omitempty"`
}

type AsMonth struct {
	CreateTime string          `gorm:"column:create_time" json:"create_time,omitempty"`
	Nums       decimal.Decimal `gorm:"column:nums" json:"nums,omitempty"`
	Valuation  decimal.Decimal `gorm:"column:valuation" json:"valuation,omitempty"`
	Freeze     decimal.Decimal `gorm:"column:freeze" json:"freeze,omitempty"`
}

func (u *AssetsMonth) TableName() string {
	return "assets_month"
}
