package userAddr

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"time"
)

type Entity struct {
	Db           *orm.CacheDB `json:"-" gorm:"-"`
	Id           int64        `json:"id" gorm:"column:id; PRIMARY_KEY"`
	MerchantUser string       `gorm:"column:merchant_user" json:"merchant_user"`
	MerchantId   int64        `gorm:"column:merchant_id" json:"merchant_id"`
	ServiceId    int64        `gorm:"column:service_id" json:"service_id"`
	CoinId       int          `gorm:"column:coin_id" json:"coin_id"`
	Address      string       `gorm:"column:address" json:"address"`
	State        int          `gorm:"column:state" json:"state"`
	ClinetId     string       `gorm:"column:clinet_id" json:"clinet_id"`
	SecureKey    string       `gorm:"column:secure_key" json:"secure_key"`
	ChainId      int          `gorm:"column:chain_id" json:"chain_id"`
	CreatedAt    time.Time    `json:"created_at" gorm:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" gorm:"updated_at"`
	DeletedAt    *time.Time   `json:"deleted_at" gorm:"deleted_at"`
}

func (e *Entity) TableName() string {
	return "user_addr"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
