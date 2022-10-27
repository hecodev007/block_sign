package api

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"time"
)

type Entity struct {
	Db        *orm.CacheDB `json:"-" gorm:"-"`
	Id        int          `json:"id" gorm:"column:id; PRIMARY_KEY"`
	Name      string       `json:"name" gorm:"column:name"`
	Path      string       `json:"path" gorm:"column:path"`
	Method    string       `json:"method" gorm:"column:method"`
	Tag       string       `json:"tag" gorm:"column:tag"`
	CreateAt  time.Time    `json:"create_at" gorm:"create_at"`
	UpdateAt  time.Time    `json:"update_at" gorm:"update_at"`
	DeletedAt *time.Time   `json:"deleted_at" gorm:"deleted_at"`
}

func (e *Entity) TableName() string {
	return "admin_api"
}
func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
