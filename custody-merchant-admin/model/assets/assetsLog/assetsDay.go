package assetsLog

import (
	"github.com/shopspring/decimal"
	"time"
)

type AssetsDay struct {
	CoinId       int             `gorm:"column:coin_id" json:"coin_id,omitempty"`
	ServiceId    int             `gorm:"column:service_id" json:"service_id,omitempty"`
	Id           int64           `json:"id" gorm:"column:id; PRIMARY_KEY"`
	CoinName     string          `gorm:"column:coin_name" json:"coin_name,omitempty"`
	ChainAddress string          `gorm:"column:chain_address" json:"chain_address,omitempty"` //blockchain钱包地址
	Nums         decimal.Decimal `gorm:"column:nums" json:"nums,omitempty"`
	Valuation    decimal.Decimal `gorm:"column:valuation" json:"valuation,omitempty"`
	Freeze       decimal.Decimal `gorm:"column:freeze" json:"freeze,omitempty"`
	CreateTime   time.Time       `gorm:"column:create_time" json:"create_time,omitempty"`
}

type AsTime struct {
	CoinId     int             `gorm:"column:coin_id" json:"coin_id,omitempty"`
	ServiceId  int             `gorm:"column:service_id" json:"service_id,omitempty"`
	CreateTime string          `gorm:"column:create_time" json:"create_time,omitempty"`
	Nums       decimal.Decimal `gorm:"column:nums" json:"nums,omitempty"`
	Price      decimal.Decimal `gorm:"column:price" json:"price,omitempty"`
	Valuation  decimal.Decimal `gorm:"column:valuation" json:"valuation,omitempty"`
	Freeze     decimal.Decimal `gorm:"column:freeze" json:"freeze,omitempty"`
}

func (u *AssetsDay) TableName() string {
	return "assets_day"
}
