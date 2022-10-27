package userPermission

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
)

type Entity struct {
	Db  *orm.CacheDB `json:"-" gorm:"-"`
	Id  int64        `json:"id" gorm:"column:id; PRIMARY_KEY"`
	Uid int64        `gorm:"column:uid" json:"uid,omitempty"`
	Mid string       `gorm:"column:mid" json:"mid,omitempty"`
}

func (e *Entity) TableName() string {
	return "user_permission"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
