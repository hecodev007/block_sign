package unitUsdt

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"errors"
)

// GetUnitUsdt
// 获取Usdt单位汇率访问信息
func (t *UnitUsdt) GetUnitUsdt(name string) (*UnitUsdt, error) {
	u := UnitUsdt{}
	if model.FilteredSQLInject(name) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	db := model.DB().Where("name = ? ", name).First(&u)
	if u.Id != 0 {
		return &u, model.ModelError(db, global.MsgWarnModelNil)
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

// GetUnitUsdtById
// 获取Usdt单位汇率访问信息
func (t *UnitUsdt) GetUnitUsdtById(id int) (*UnitUsdt, error) {
	u := UnitUsdt{}
	db := model.DB().Where("id =?", id).First(&u)
	if u.Id != 0 {
		return &u, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

// GetUnitUsdtList
// 获取Usdt单位汇率访问列表
func (t *UnitUsdt) GetUnitUsdtList() ([]UnitUsdt, error) {
	var u []UnitUsdt
	db := model.DB().Find(&u)
	return u, model.ModelError(db, global.MsgWarnModelNil)
}

// SaveUnitUsdt
// 新增Usdt单位汇率访问
func (u *UnitUsdt) SaveUnitUsdt(unit *UnitUsdt) error {
	tx := model.DB().Begin()
	tx.Create(unit)
	err := model.ModelError(tx, global.MsgWarnModelAdd)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// UpDateUnitUsdt
// 更新Usdt单位汇率访问
func (u *UnitUsdt) UpDateUnitUsdt(unit *UnitUsdt) error {
	db := model.DB().Save(unit)
	return model.ModelError(db, global.MsgWarnModelUpdate)
}

func (u *UnitUsdt) UpdateUnitById(id int, mp map[string]interface{}) error {
	db := model.DB().Table("unit_usdt").Where("id =?", id).Updates(mp)
	return model.ModelError(db, global.MsgWarnModelUpdate)
}
