package userAddr

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
)

func (e *Entity) FindAddressByUId(mid int64, uid string) ([]Entity, error) {
	lst := []Entity{}
	db := model.DB().Where("merchant_id=? and merchant_user=?", mid, uid).Find(&lst)
	return lst, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) FindAddressByAddr(addr string) error {

	db := model.DB().Table(e.TableName()).Where("address=?", addr).First(e)
	return model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) CreateUserAddress() error {
	db := model.DB().Begin()
	db.Table(e.TableName()).Omit("deleted_at", "updated_at").Create(e)
	if db.Error != nil {
		db.Rollback()
		return model.ModelError(db, global.MsgWarnModelAdd)
	}
	db.Commit()
	return model.ModelError(db, global.MsgWarnModelNil)
}
