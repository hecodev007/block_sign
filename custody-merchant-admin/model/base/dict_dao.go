package base

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"errors"
)

func (e *DictList) FindDictByType(tag string) ([]DictList, error) {
	var dl []DictList
	if model.FilteredSQLInject(tag) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	db := model.DB().Model(&DictList{}).Joins("left join dict d on d.id = dict_list.dict_id").Where(" d.dict_type = ?", tag).Order("sort asc").Find(&dl)
	return dl, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *DictList) FindDictByTypeValue(tag string, v int) (*DictList, error) {
	dl := new(DictList)
	if model.FilteredSQLInject(tag) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	db := model.DB().Model(dl).Joins("left join dict on dict.id = dict_list.dict_id").
		Where(" dict.dict_type = ? and dict_list.dict_value =?", tag, v).
		First(&dl)
	if dl.Id != 0 {
		return dl, nil
	} else {
		return nil, model.ModelError(db, global.MsgWarnModelNil)
	}
}
