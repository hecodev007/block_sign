package dao

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"errors"
)

func (e *Entity) FindPhoneCodeAll() ([]Entity, error) {
	var codes []Entity
	db := model.DB().Table("phone_code").Find(&codes)
	if len(codes) != 0 {
		return codes, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) FindPhoneByCode(code string) (*Entity, error) {
	codes := new(Entity)
	if model.FilteredSQLInject(code) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	db := model.DB().Table("phone_code").Where("code_value = ? ", code).First(&codes)
	if codes.Id != 0 {
		return codes, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}
