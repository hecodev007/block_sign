package role_menu

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
)

type Entity struct {
	Db     *orm.CacheDB `json:"-" gorm:"-"`
	RoleId int          `json:"role_id" gorm:"column:role_id"`
	MId    int          `json:"m_id" gorm:"column:m_id"`
	Id     int64        `json:"id" gorm:"column:id; PRIMARY_KEY"`
}

func (e *Entity) TableName() string {
	return "admin_role_menu"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
