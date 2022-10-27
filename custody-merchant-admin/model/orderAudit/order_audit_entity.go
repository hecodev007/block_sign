package orderAudit

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"time"
)

type Entity struct {
	Db          *orm.CacheDB `json:"-" gorm:"-"`
	AuditLevel  int          `json:"audit_level" gorm:"column:audit_level"`
	State       int          `json:"state" gorm:"column:state"`
	AuditResult int          `json:"audit_result" gorm:"column:audit_result"`
	Id          int64        `json:"id" gorm:"column:id; PRIMARY_KEY"`
	OrderId     int64        `json:"order_id" gorm:"column:order_id"`
	UserId      int64        `json:"user_id" gorm:"column:user_id"`
	CreateTime  time.Time    `json:"create_time" gorm:"column:create_time"`
	UpdateTime  time.Time    `json:"update_time" gorm:"column:update_time"`
}

func (e *Entity) TableName() string {
	return "order_audit"
}
func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
