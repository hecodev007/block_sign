package serviceChains

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"time"
)

type Entity struct {
	Db           *orm.CacheDB `json:"-" gorm:"-"`
	Id           int          `json:"id" gorm:"column:id; PRIMARY_KEY"`
	MerchantId   int64        `json:"merchant_id" gorm:"column:merchant_id"`
	Account      string       `json:"account" gorm:"column:account"`
	ServiceId    int          `json:"service_id" gorm:"column:service_id"`
	CoinId       int          `json:"coin_id" gorm:"column:coin_id" `
	CoinName     string       `gorm:"column:coin_name" json:"coin_name,omitempty"`
	ChainAddr    string       `json:"chain_addr" gorm:"column:chain_addr"`
	Reason       string       `json:"reason,omitempty" gorm:"column:reason"`
	IsWithdrawal int          `json:"is_withdrawal" gorm:"column:is_withdrawal"`
	IsGetAddr    int          `json:"is_get_addr" gorm:"column:is_get_addr"`
	State        int          `gorm:"column:state" json:"state"`
	CreatedAt    time.Time    `json:"created_at" gorm:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" gorm:"updated_at"`
	DeletedAt    *time.Time   `json:"deleted_at" gorm:"deleted_at"`
}

func (e *Entity) TableName() string {
	return "service_chains"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
