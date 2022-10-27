package financeAssets

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"github.com/shopspring/decimal"
	"time"
)

//Entity 托管后台财务钱包表
type Entity struct {
	Db         *orm.CacheDB    `json:"-" gorm:"-"`
	Id         int64           `json:"id" gorm:"id"`
	Nums       decimal.Decimal `json:"nums" gorm:"nums"`
	Freeze     decimal.Decimal `json:"freeze" gorm:"freeze"`     //冻结金额
	CoinId     int64           `json:"coin_id" gorm:"coin_id"`   //币id
	Coin       string          `json:"coin" gorm:"coin"`         //主链币
	SubCoin    string          `json:"sub_coin" gorm:"sub_coin"` //代币
	Token      string          `json:"token" gorm:"token"`       //合约
	Address    string          `json:"address" gorm:"address"`   //地址
	BusinessId int64           `json:"business_id" gorm:"business_id"`
	AccountId  int64           `json:"account_id" gorm:"account_id"`
	Remark     string          `json:"remark" gorm:"remark"`
	CreatedAt  time.Time       `json:"created_at" gorm:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at" gorm:"updated_at"`
	DeletedAt  *time.Time      `json:"deleted_at" gorm:"deleted_at"`
}

func (e *Entity) TableName() string {
	return "admin_finance_assets"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
