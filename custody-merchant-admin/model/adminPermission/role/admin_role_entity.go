package role

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
)

type Entity struct {
	Db     *orm.CacheDB `json:"-" gorm:"-"`
	Id     int          `json:"id" gorm:"column:id; PRIMARY_KEY"`
	State  int          `json:"state" gorm:"column:state"`
	Name   string       `json:"name" gorm:"column:name"`
	Tag    string       `json:"tag" gorm:"column:tag"`
	Remark string       `json:"remark" gorm:"column:remark"`
}

func (e *Entity) TableName() string {
	return "admin_role"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
