package businessPackage

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"github.com/shopspring/decimal"
	"time"
)

//Entity 业务线套餐表
type Entity struct {
	Db                *orm.CacheDB    `json:"-" gorm:"-"`
	Id                int             `json:"id" gorm:"column:id; PRIMARY_KEY"`
	AccountId         int64           `json:"account_id" gorm:"column:account_id"`           //商户id
	BusinessId        int64           `json:"business_id" gorm:"column:business_id"`         //业务线id
	PackageId         int64           `json:"package_id,omitempty" gorm:"column:package_id"` //套餐id
	TypeName          string          `json:"type_name,omitempty" gorm:"column:type_name"`   //套餐类型
	ModelName         string          `json:"model_name,omitempty" gorm:"column:model_name"`
	OrderType         string          `json:"order_type,omitempty" gorm:"column:order_type"`   //交易类型
	DeductCoin        string          `json:"deduct_coin,omitempty" gorm:"column:deduct_coin"` //扣费币种
	HadUsed           decimal.Decimal `json:"had_used" gorm:"had_used"`                        //已使用数量（流水/月租 记录为金额，地址记录为个数）
	EnterUnit         int             `json:"enter_unit,omitempty" gorm:"column:enter_unit"`
	LimitType         int             `json:"limit_type,omitempty" gorm:"column:limit_type"`
	TypeNums          decimal.Decimal `json:"type_nums,omitempty" gorm:"column:type_nums"`
	TopUpType         int             `json:"top_up_type,omitempty" gorm:"column:top_up_type"`
	TopUpFee          string          `json:"top_up_fee,omitempty" gorm:"column:top_up_fee"`
	WithdrawalType    int             `json:"withdrawal_type,omitempty" gorm:"column:withdrawal_type"`
	WithdrawalFee     string          `json:"withdrawal_fee,omitempty" gorm:"column:withdrawal_fee"`
	ServiceDiscount   decimal.Decimal `json:"service_discount,omitempty" gorm:"column:service_discount"` //当前业务线所用套餐增加此业务线的业务线优惠金额
	ChainNums         int             `json:"chain_nums,omitempty" gorm:"column:chain_nums"`
	ChainDiscountUnit int             `json:"chain_discount_unit,omitempty" gorm:"column:chain_discount_unit"`
	ChainDiscountNums decimal.Decimal `json:"chain_discount_nums,omitempty" gorm:"column:chain_discount_nums"`
	ChainTimeUnit     int             `json:"chain_time_unit,omitempty" gorm:"column:chain_time_unit"`
	CoinNums          int             `json:"coin_nums,omitempty" gorm:"column:coin_nums"`
	CoinDiscountUnit  int             `json:"coin_discount_unit,omitempty" gorm:"column:coin_discount_unit"`
	CoinDiscountNums  decimal.Decimal `json:"coin_discount_nums,omitempty" gorm:"column:coin_discount_nums"`
	CoinTimeUnit      int             `json:"coin_time_unit,omitempty" gorm:"column:coin_time_unit"`
	DeployFee         decimal.Decimal `json:"deploy_fee,omitempty" gorm:"column:deploy_fee"`
	CustodyFee        decimal.Decimal `json:"custody_fee,omitempty" gorm:"column:custody_fee"`
	DepositFee        decimal.Decimal `json:"deposit_fee,omitempty" gorm:"column:deposit_fee"`
	AddrNums          int             `json:"addr_nums,omitempty" gorm:"column:addr_nums"`
	CoverFee          decimal.Decimal `json:"cover_fee,omitempty" gorm:"column:cover_fee"`
	ComboDiscountUnit int             `json:"combo_discount_unit,omitempty" gorm:"column:combo_discount_unit"`
	ComboDiscountNums decimal.Decimal `json:"combo_discount_nums,omitempty" gorm:"column:combo_discount_nums"`
	YearDiscountUnit  int             `json:"year_discount_unit,omitempty" gorm:"column:year_discount_unit"`
	YearDiscountNums  decimal.Decimal `json:"year_discount_nums,omitempty" gorm:"column:year_discount_nums"`
	CreatedAt         time.Time       `json:"created_at" form:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at" form:"updated_at"`
	DeletedAt         *time.Time      `json:"deleted_at" form:"deleted_at"`
}

func (e *Entity) TableName() string {
	return "service_combo"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}

type MchPackageDB struct {
	Name              string          `json:"name" gorm:"column:name"`
	Coin              string          `json:"coin" gorm:"column:coin"`
	SubCoin           string          `json:"sub_coin" gorm:"column:sub_coin"`
	AccountId         int64           `json:"account_id" gorm:"column:account_id"`           //商户id
	BusinessId        int64           `json:"business_id" gorm:"column:business_id"`         //业务线id
	PackageId         int64           `json:"package_id,omitempty" gorm:"column:package_id"` //套餐id
	TypeName          string          `json:"type_name,omitempty" gorm:"column:type_name"`   //套餐类型
	ModelName         string          `json:"model_name,omitempty" gorm:"column:model_name"`
	DeductCoin        string          `json:"deduct_coin" gorm:"column:deduct_coin"` //扣费币种
	EnterUnit         int             `json:"enter_unit,omitempty" gorm:"column:enter_unit"`
	LimitType         int             `json:"limit_type,omitempty" gorm:"column:limit_type"`
	TypeNums          decimal.Decimal `json:"type_nums,omitempty" gorm:"column:type_nums"`
	TopUpType         int             `json:"top_up_type,omitempty" gorm:"column:top_up_type"`
	TopUpFee          string          `json:"top_up_fee,omitempty" gorm:"column:top_up_fee"`
	WithdrawalType    int             `json:"withdrawal_type,omitempty" gorm:"column:withdrawal_type"`
	WithdrawalFee     string          `json:"withdrawal_fee,omitempty" gorm:"column:withdrawal_fee"`
	ServiceDiscount   decimal.Decimal `json:"service_discount,omitempty" gorm:"column:service_discount"` //当前业务线所用套餐增加此业务线的业务线优惠金额
	ChainNums         int             `json:"chain_nums,omitempty" gorm:"column:chain_nums"`
	ChainDiscountUnit int             `json:"chain_discount_unit,omitempty" gorm:"column:chain_discount_unit"`
	ChainDiscountNums decimal.Decimal `json:"chain_discount_nums,omitempty" gorm:"column:chain_discount_nums"`
	ChainTimeUnit     int             `json:"chain_time_unit,omitempty" gorm:"column:chain_time_unit"`
	CoinNums          int             `json:"coin_nums,omitempty" gorm:"column:coin_nums"`
	CoinDiscountUnit  int             `json:"coin_discount_unit,omitempty" gorm:"column:coin_discount_unit"`
	CoinDiscountNums  decimal.Decimal `json:"coin_discount_nums,omitempty" gorm:"column:coin_discount_nums"`
	CoinTimeUnit      int             `json:"coin_time_unit,omitempty" gorm:"column:coin_time_unit"`
	DeployFee         decimal.Decimal `json:"deploy_fee,omitempty" gorm:"column:deploy_fee"`
	CustodyFee        decimal.Decimal `json:"custody_fee,omitempty" gorm:"column:custody_fee"`
	DepositFee        decimal.Decimal `json:"deposit_fee,omitempty" gorm:"column:deposit_fee"`
	AddrNums          int             `json:"addr_nums,omitempty" gorm:"column:addr_nums"`
	CoverFee          decimal.Decimal `json:"cover_fee,omitempty" gorm:"column:cover_fee"`
	ComboDiscountUnit int             `json:"combo_discount_unit,omitempty" gorm:"column:combo_discount_unit"`
	ComboDiscountNums decimal.Decimal `json:"combo_discount_nums,omitempty" gorm:"column:combo_discount_nums"`
	YearDiscountUnit  int             `json:"year_discount_unit,omitempty" gorm:"column:year_discount_unit"`
	YearDiscountNums  decimal.Decimal `json:"year_discount_nums,omitempty" gorm:"column:year_discount_nums"`
}
