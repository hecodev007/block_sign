package _package

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"github.com/shopspring/decimal"
	"time"
)

//Entity 套餐模版表
type Entity struct {
	Db                  *orm.CacheDB    `json:"-" gorm:"-"`
	Id                  int             `json:"id" gorm:"column:id; PRIMARY_KEY"`
	TypeName            string          `json:"type_name,omitempty" gorm:"column:type_name"`
	ModelName           string          `json:"model_name,omitempty" gorm:"column:model_name"`
	EnterUnit           int             `json:"enter_unit,omitempty" gorm:"column:enter_unit"`
	LimitType           int             `json:"limit_type,omitempty" gorm:"column:limit_type"`
	TypeNums            decimal.Decimal `json:"type_nums,omitempty" gorm:"column:type_nums"`
	TopUpType           int             `json:"top_up_type,omitempty" gorm:"column:top_up_type"`
	TopUpFee            string          `json:"top_up_fee,omitempty" gorm:"column:top_up_fee"` //充值收费 百分数/usdt，带单位
	WithdrawalType      int             `json:"withdrawal_type,omitempty" gorm:"column:withdrawal_type"`
	WithdrawalFee       string          `json:"withdrawal_fee,omitempty" gorm:"column:withdrawal_fee"` //提现收费类型 百分数/usdt，带单位
	ServiceNums         int             `json:"service_nums,omitempty" gorm:"column:service_nums"`
	ServiceDiscountUnit int             `json:"service_discount_unit,omitempty" gorm:"column:service_discount_unit"`
	ServiceDiscountNums decimal.Decimal `json:"service_discount_nums,omitempty" gorm:"column:service_discount_nums"`
	ChainNums           int             `json:"chain_nums,omitempty" gorm:"column:chain_nums"`
	ChainDiscountUnit   int             `json:"chain_discount_unit,omitempty" gorm:"column:chain_discount_unit"`
	ChainDiscountNums   decimal.Decimal `json:"chain_discount_nums,omitempty" gorm:"column:chain_discount_nums"`
	ChainTimeUnit       int             `json:"chain_time_unit,omitempty" gorm:"column:chain_time_unit"`
	CoinNums            int             `json:"coin_nums,omitempty" gorm:"column:coin_nums"`
	CoinDiscountUnit    int             `json:"coin_discount_unit,omitempty" gorm:"column:coin_discount_unit"`
	CoinDiscountNums    decimal.Decimal `json:"coin_discount_nums,omitempty" gorm:"column:coin_discount_nums"`
	CoinTimeUnit        int             `json:"coin_time_unit,omitempty" gorm:"column:coin_time_unit"`
	DeployFee           decimal.Decimal `json:"deploy_fee,omitempty" gorm:"column:deploy_fee"`
	CustodyFee          decimal.Decimal `json:"custody_fee,omitempty" gorm:"column:custody_fee"`
	DepositFee          decimal.Decimal `json:"deposit_fee,omitempty" gorm:"column:deposit_fee"`
	AddrNums            int             `json:"addr_nums,omitempty" gorm:"column:addr_nums"`
	CoverFee            decimal.Decimal `json:"cover_fee,omitempty" gorm:"column:cover_fee"`
	ComboDiscountUnit   int             `json:"combo_discount_unit,omitempty" gorm:"column:combo_discount_unit"`
	ComboDiscountNums   decimal.Decimal `json:"combo_discount_nums,omitempty" gorm:"column:combo_discount_nums"`
	YearDiscountUnit    int             `json:"year_discount_unit,omitempty" gorm:"column:year_discount_unit"`
	YearDiscountNums    decimal.Decimal `json:"year_discount_nums,omitempty" gorm:"column:year_discount_nums"`
	CreatedAt           time.Time       `json:"created_at,omitempty" gorm:"created_at,omitempty"`
	UpdatedAt           time.Time       `json:"updated_at,omitempty" gorm:"updated_at,omitempty"`
	DeletedAt           *time.Time      `json:"deleted_at,omitempty" gorm:"deleted_at,omitempty"`
}

func (e *Entity) TableName() string {
	return "admin_package"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}

//PackagePay 套餐收费类型表
type PackagePay struct {
	Db        *orm.CacheDB `json:"-" gorm:"-"`
	Id        int          `json:"id" gorm:"column:id; PRIMARY_KEY"`
	PayType   string       `json:"pay_type,omitempty" gorm:"column:pay_type"`
	PayName   string       `json:"pay_name,omitempty" gorm:"column:pay_name"`
	CreatedAt time.Time    `json:"created_at" gorm:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" gorm:"updated_at"`
	DeletedAt *time.Time   `json:"deleted_at" gorm:"deleted_at"`
}

func (e *PackagePay) PackagePayTableName() string {
	return "admin_package_pay"
}

//PackageTrade 套餐交易类型表
type PackageTrade struct {
	Db        *orm.CacheDB `json:"-" gorm:"-"`
	Id        int          `json:"id" gorm:"column:id; PRIMARY_KEY"`
	TradeType string       `json:"trade_type,omitempty" gorm:"column:trade_type"` //(open-开通，2renew-二次续费，3renew-三次续费。。。。)
	TradeName string       `json:"trade_name,omitempty" gorm:"column:trade_name"`
	CreatedAt time.Time    `json:"created_at" form:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" form:"updated_at"`
	DeletedAt *time.Time   `json:"deleted_at" form:"deleted_at"`
}

func (e *PackageTrade) PackageTradeTableName() string {
	return "admin_package_trade"
}
