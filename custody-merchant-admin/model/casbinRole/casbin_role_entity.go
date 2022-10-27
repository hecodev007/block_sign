package casbinRole

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"time"
)

type Entity struct {
	Db     *orm.CacheDB `json:"-" gorm:"-"`
	PType  string       `json:"p_type" gorm:"column:p_type" description:"策略类型"`
	UserId string       `json:"user_id" gorm:"column:v0" description:"用户ID"`
	Path   string       `json:"path" gorm:"column:impl" description:"api路径"`
	Method string       `json:"method" gorm:"column:v2" description:"访问方法"`
}

func (c *Entity) TableName() string {
	return "casbin_rule"
}
func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}

type SysRouter struct {
	Db         *orm.CacheDB `json:"-" gorm:"-"`
	Id         int          `json:"id" gorm:"column:id; PRIMARY_KEY"`
	Name       string       `json:"name" gorm:"column:name"`
	Path       string       `json:"path" gorm:"column:path"`
	Method     string       `json:"method" gorm:"column:method"`
	Tag        string       `json:"tag" gorm:"column:tag"`
	CreateTime time.Time    `json:"create_time" gorm:"column:create_time"`
	UpdateTime time.Time    `json:"update_time" gorm:"column:update_time"`
}

func (sr *SysRouter) TableName() string {
	return "sys_router"
}
