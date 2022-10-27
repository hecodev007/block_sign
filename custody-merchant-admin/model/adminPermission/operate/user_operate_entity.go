package operate

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
	"time"
)

type Entity struct {
	Db        *orm.CacheDB `json:"-" gorm:"-"`
	Id        int64        `json:"id" gorm:"column:id; PRIMARY_KEY"`
	UserId    int64        `json:"user_id" gorm:"column:user_id"`
	UserName  string       `json:"user_name" gorm:"column:user_name"`
	CreatedBy string       `json:"created_by" gorm:"column:created_by"`
	Content   string       `json:"content" gorm:"column:content"`
	Platform  string       `json:"platform" gorm:"column:platform"`
	CreatedAt time.Time    `json:"created_at" gorm:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" gorm:"updated_at"`
	DeletedAt *time.Time   `json:"deleted_at" gorm:"deleted_at"`
}

func (e *Entity) TableName() string {
	return "user_operate"
}

func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
