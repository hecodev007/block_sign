package serviceAuditRole

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
)

type Entity struct {
	Db    *orm.CacheDB `json:"-" gorm:"-"`
	Id    int          `json:"id" gorm:"column:id; PRIMARY_KEY"`
	Uid   int64        `gorm:"column:uid" json:"uid"`
	Sid   int          `gorm:"column:sid" json:"sid"`
	Aid   int          `gorm:"column:aid" json:"aid"`
	State int          `gorm:"column:state" json:"state"`
}

func (e *Entity) TableName() string {
	return "service_audit_role"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
