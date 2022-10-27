package comboUse

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"custody-merchant-admin/module/log"
	"errors"
)

func (u *Entity) UpDateComboUserDayByCId(aId, cId int64, cTime string, version int, mp map[string]interface{}) (int, error) {

	db := model.DB().Model(&Entity{}).
		Where("combo_user_id=? and package_id=? and DATE_FORMAT(create_time,'%Y-%m-%d') = DATE_FORMAT(?,'%Y-%m-%d') and version =?",
			cId, aId, cTime, version).
		Updates(mp)
	err := model.ModelError(db, global.MsgWarnModelUpdate)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func (u *Entity) CreateComboUserDay(com Entity) (int64, error) {

	// 通过数据的指针来创建
	db := model.DB().Begin()
	db.Create(&com)
	if err := db.Error; err != nil {
		db.Rollback()
		log.Errorf("CreateComboUserDay error: %v", err)
		return 0, err
	}
	db.Commit()
	return db.RowsAffected, nil
}

func (u *Entity) FindComboUserDayByCId(aId, cId int64, ctime string) (*Entity, error) {
	coms := new(Entity)
	if model.FilteredSQLInject(ctime) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	db := model.DB().Model(&Entity{}).Where("package_id =? and combo_user_id = ? and DATE_FORMAT(create_time,'%Y-%m-%d') = DATE_FORMAT(?,'%Y-%m-%d')", aId, cId, ctime).First(coms) // 通过数据的指针来创建

	if coms.Id != 0 {
		return coms, nil
	} else {
		return nil, model.ModelError(db, global.MsgWarnModelNil)
	}
}

func (u *Entity) SumComboUserDayByCId(aId, cId int64) (*Entity, error) {
	coms := new(Entity)
	db := model.DB().Model(&Entity{}).Select("sum(used_addr_day) as used_addr_day").
		Where("combo_user_id = ? and package_id =? ", aId, cId).
		First(coms) // 通过数据的指针来创建

	if coms.Id != 0 {
		return coms, nil
	} else {
		return nil, model.ModelError(db, global.MsgWarnModelNil)
	}
}
