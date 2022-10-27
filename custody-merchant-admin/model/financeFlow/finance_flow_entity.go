package financeFlow

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"github.com/shopspring/decimal"
	"time"
)

//Entity 托管后台财务钱包流水表
type Entity struct {
	Db          *orm.CacheDB    `json:"-" gorm:"-"`
	Id          int64           `json:"id" gorm:"id"`
	OrderId     string          `json:"order_id" gorm:"order_id"`         //合约地址
	FlowType    string          `json:"flow_type" gorm:"flow_type"`       //类型 in-收入，out-支出
	FromAddress string          `json:"from_address" gorm:"from_address"` //源地址
	ToAddress   string          `json:"to_address" gorm:"to_address"`     //转出地址
	Nums        decimal.Decimal `json:"nums" gorm:"nums"`
	CoinId      int64           `json:"coin_id" gorm:"coin_id"`     //币id
	CoinName    string          `json:"coin_name" gorm:"coin_name"` //币
	Token       string          `json:"token" gorm:"token"`         //合约地址
	BusinessId  int64           `json:"business_id" gorm:"business_id"`
	AccountId   int64           `json:"account_id" gorm:"account_id"`
	Remark      string          `json:"remark" gorm:"remark"`
	CreatedAt   time.Time       `json:"created_at" gorm:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" gorm:"updated_at"`
	DeletedAt   *time.Time      `json:"deleted_at" gorm:"deleted_at"`
}

func (e *Entity) TableName() string {
	return "admin_finance_assets_flow"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
