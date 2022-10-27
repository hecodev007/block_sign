package assets

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"github.com/shopspring/decimal"
)

type Assets struct {
	Db            *orm.CacheDB    `json:"-" gorm:"-"`
	CoinId        int             `gorm:"column:coin_id" json:"coin_id"`
	ServiceId     int             `gorm:"column:service_id" json:"service_id"`
	Id            int64           `gorm:"column:id; PRIMARY_KEY" json:"id"`
	CoinName      string          `gorm:"column:coin_name" json:"coin_name"`
	Nums          decimal.Decimal `gorm:"column:nums" json:"nums"`
	Freeze        decimal.Decimal `gorm:"column:freeze" json:"freeze"`
	FinanceFreeze decimal.Decimal `gorm:"column:finance_freeze" json:"finance_freeze"`
	Version       int64           `gorm:"column:version" json:"version"`
}

type AssetsList struct {
	ServiceId int             `gorm:"column:service_id" json:"service_id,omitempty"`
	ChainId   int             `gorm:"column:chain_id" json:"chain_id,omitempty"`
	ChainName string          `gorm:"column:chain_name" json:"chain_name,omitempty"`
	CoinId    int             `gorm:"column:coin_id" json:"coin_id,omitempty"`
	CoinName  string          `gorm:"column:coin_name" json:"coin_name,omitempty"`
	Nums      decimal.Decimal `gorm:"column:nums" json:"nums,omitempty"`
	Valuation decimal.Decimal `gorm:"column:valuation" json:"valuation,omitempty"`
	Freeze    decimal.Decimal `gorm:"column:freeze" json:"freeze,omitempty"`
}

type FinanceServiceAsset struct {
	ServiceId     int             `gorm:"column:service_id" json:"service_id"`
	AccountId     int64           `gorm:"column:account_id" json:"account_id"`
	ServiceName   string          `gorm:"column:service_name" json:"service_name"`
	UserName      string          `gorm:"column:user_name" json:"user_name"`
	CoinName      string          `gorm:"column:coin_name" json:"coin_name"`
	FinanceFreeze decimal.Decimal `gorm:"column:finance_freeze" json:"finance_freeze"`
}

func (ass *Assets) TableName() string {
	return "assets"
}

func NewEntity() *Assets {
	e := Assets{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
