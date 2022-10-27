package comboUse

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"github.com/shopspring/decimal"
	"time"
)

type Entity struct {
	Db          *orm.CacheDB    `json:"-" gorm:"-"`
	Id          int64           `json:"id" gorm:"column:id; PRIMARY_KEY"`
	ComboUserId int64           `json:"combo_user_id" gorm:"column:combo_user_id"`
	PackageId   int64           `json:"package_id" gorm:"column:package_id"`
	UsedAddrDay int64           `json:"used_addr_day" gorm:"column:used_addr_day"`
	UsedLineDay decimal.Decimal `json:"used_line_day" gorm:"column:used_line_day"`
	CreateTime  time.Time       `json:"create_time" gorm:"column:create_time"`
	Version     int             `json:"version" gorm:"column:version"`
}

func (u *Entity) TableName() string {
	return "combo_user_day"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
