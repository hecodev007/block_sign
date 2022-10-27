package dao

import (
	"custody-merchant-admin/model"
	"custody-merchant-admin/model/orm"
)

type Entity struct {
	Db        *orm.CacheDB `json:"-" gorm:"-"`
	Id        int          `json:"id" gorm:"column:id; PRIMARY_KEY"`
	State     int          `json:"state" gorm:"column:state"`
	Tag       string       `json:"tag" gorm:"column:tag"`
	CodeName  string       `json:"code_name" gorm:"column:code_name"`
	CodeValue string       `json:"code_value" gorm:"column:code_value"`
	Remark    string       `json:"remark" gorm:"column:remark"`
}

func (e *Entity) TableName() string {
	return "phone_code"
}
func NewEntity() *Entity {
	e := Entity{
		Db: orm.Cache(model.DB()),
	}
	return &e
}
