package api

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"time"
)

type Entity struct {
	Db         *orm.CacheDB `json:"-" gorm:"-"`
	Type       int          `json:"type" gorm:"column:type"`
	Sort       int          `json:"sort" gorm:"column:sort"`
	State      int          `json:"state" gorm:"column:state"`
	Id         int64        `json:"id" gorm:"column:id; PRIMARY_KEY"`
	Pid        int64        `json:"pid" gorm:"column:pid"`
	Label      string       `json:"label" gorm:"column:label"`
	Icon       string       `json:"icon" gorm:"column:icon"`
	Component  string       `gorm:"column:component" json:"component"` // 组件路径
	Path       string       `json:"path" gorm:"column:path"`
	BtnPath    string       `json:"btn_path" gorm:"column:btn_path"`
	Tag        string       `json:"tag" gorm:"column:tag"`
	MenuType   int          `json:"menu_type" gorm:"column:menu_type"`
	Remark     string       `json:"remark" gorm:"column:remark"`
	ActiveMenu string       `json:"active_menu"`
	Hidden     bool         `json:"hidden"`
	CreateAt   time.Time    `json:"create_at" gorm:"create_at"`
	UpdateAt   time.Time    `json:"update_at" gorm:"update_at"`
	DeletedAt  *time.Time   `json:"deleted_at" gorm:"deleted_at"`
}

func (e *Entity) TableName() string {
	return "admin_menu"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
