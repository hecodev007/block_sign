package white

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"errors"
)

func GetWhiteListUse(sId, cId int, addr string) (*WhiteList, error) {

	w := new(WhiteList)
	if model.FilteredSQLInject(addr) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	db := model.DB().Table("white_list").Where("service_id=? and coin_id=? and address=? ", sId, cId, addr).First(w)
	return w, model.ModelError(db, global.MsgWarnModelNil)
}

func GetWhiteUseById(id int64) (*WhiteList, error) {
	var w = new(WhiteList)
	db := model.DB().Where("id = ? ", id).First(&w)
	err := model.ModelError(db, global.MsgWarnModelNil)

	if err != nil {
		return nil, err
	}
	if w.Id != 0 {
		return w, nil
	}
	return nil, err
}

func GetWhiteByAddr(addr string, sid int) (*WhiteList, error) {
	var w = new(WhiteList)
	db := model.DB().Where("address = ? and service_id = ? ", addr, sid).First(&w)
	err := model.ModelError(db, global.MsgWarnModelNil)
	if err != nil {
		return nil, err
	}
	if w.Id != 0 {
		return w, nil
	}
	return nil, err
}

func CloseWhiteUseById(id int64, mp map[string]interface{}) (*WhiteList, error) {
	var w = new(WhiteList)
	db := model.DB().Where("id = ? ", id).Updates(mp)
	err := model.ModelError(db, global.MsgWarnModelNil)

	if err != nil {
		return nil, err
	}
	if w.Id != 0 {
		return w, nil
	}
	return nil, err
}
