package fullYear

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"time"
)

//Entity 满年判断表
type Entity struct {
	Db         *orm.CacheDB `json:"-" gorm:"-"`
	Id         int64        `json:"id" gorm:"id"`
	AccountId  int64        `json:"account_id" gorm:"account_id"`   //商户id
	PackageId  int64        `json:"package_id" gorm:"package_id"`   //套餐id
	BusinessId int64        `json:"business_id" gorm:"business_id"` //业务线id
	LatestTime time.Time    `json:"latest_time" gorm:"latest_time"` //最近时间。合同开始时间/用户注册时间/最新订单支付时间 最小值
	CreatedAt  time.Time    `json:"created_at" gorm:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at" gorm:"updated_at"`
	DeletedAt  *time.Time   `json:"deleted_at" gorm:"deleted_at"`
}

func (e *Entity) TableName() string {
	return "admin_full_year"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
