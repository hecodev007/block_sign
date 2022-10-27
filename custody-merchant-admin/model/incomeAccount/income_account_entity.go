package incomeAccount

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"github.com/shopspring/decimal"
	"time"
)

//Entity 收益户记录
type Entity struct {
	Db              *orm.CacheDB    `json:"-" gorm:"-"`
	Id              int64           `json:"id" gorm:"id"`
	MerchantId      int64           `json:"merchant_id" gorm:"column:merchant_id"`
	Version         int64           `json:"version" gorm:"column:version"`
	ServiceId       int             `json:"service_id,omitempty" gorm:"column:service_id"`
	ComboTypeName   string          `json:"combo_type_name,omitempty" gorm:"column:combo_type_name"`
	ComboModelName  string          `json:"combo_model_name,omitempty" gorm:"column:combo_model_name"`
	ComboId         int             `json:"combo_id,omitempty" gorm:"column:combo_id"`
	CoinId          int             `json:"coin_id,omitempty" gorm:"column:coin_id"`
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
	CreatedAt       time.Time       `json:"created_at" gorm:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" gorm:"updated_at"`
	DeletedAt       *time.Time      `json:"deleted_at" gorm:"deleted_at"`
}

func (e *Entity) TableName() string {
	return "income_account"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
