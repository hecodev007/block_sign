package domain

import (
	"github.com/shopspring/decimal"
)

type AssetsSelect struct {
	AssetsByTag
	MerchantId   int64  `json:"merchant_id"`
	ServiceId    int    `json:"service_id"`
	ServiceState int    `json:"service_state"`
	CoinState    int    `json:"coin_state"`
	IsTest       int    `json:"is_test"`
	CoinId       int    `json:"coin_id"`
	UnitId       int    `json:"unit_id"`
	Show         int    `json:"show"`
	Limit        int    `json:"limit"  description:"查询条数" example:"10"`
	Offset       int    `json:"offset" description:"查询起始位置" example:"0"`
	CoinName     string `json:"coin_name"`
}

type AssetsInfo struct {
	Id        int64           `json:"id,omitempty"`
	ServiceId int             `json:"service_id"`
	ChainId   int             `json:"chain_id"`
	ChainName string          `json:"chain_name"`
	CoinId    int             `json:"coin_id"`
	CoinName  string          `json:"coin_name"`
	CoinPrice decimal.Decimal `json:"coin_price"`
	Nums      decimal.Decimal `json:"nums"`
	Freeze    decimal.Decimal `json:"freeze"`
	Valuation decimal.Decimal `json:"valuation"`
	Reduced   decimal.Decimal `json:"reduced,omitempty"`
}

type AssetsRingInfo struct {
	CoinName  string          `json:"coin_name"`
	Nums      decimal.Decimal `json:"nums"`
	Price     decimal.Decimal `json:"price"`
	Valuation decimal.Decimal `json:"valuation"`
}

type AssetsRing struct {
	UserId int64 `json:"user_id"`
}

type AssetsByTag struct {
	UserId     int64    `json:"user_id"`
	Tag        string   `json:"tag"`
	SelectTime []string `json:"select_time"`
	StartTime  string   `json:"start_time"`
	EndTime    string   `json:"end_time"`
}

type AssetsTimeInfo struct {
	Scale      string          `json:"scale,omitempty"`
	CreateTime string          `json:"create_time,omitempty"`
	Price      decimal.Decimal `json:"price,omitempty"`
	Freeze     decimal.Decimal `json:"freeze,omitempty"`
	PriceNums  decimal.Decimal `json:"price_nums,omitempty"`
	Nums       decimal.Decimal `json:"nums,omitempty"`
	Valuation  decimal.Decimal `json:"valuation,omitempty"`
}
