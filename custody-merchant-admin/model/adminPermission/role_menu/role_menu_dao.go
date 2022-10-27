package role_menu

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"custody-merchant-admin/module/log"
)

func (e *Entity) CreateRoleMenus(rm Entity) (int, error) {
	db := model.DB().Begin()
	db.Create(&rm)
	err := model.ModelError(db, global.MsgWarnModelAdd)
	if err != nil {
		db.Rollback()
		log.Errorf("CreateRoleMenus error: %v", err)
		return 0, err
	}
	db.Commit()
	return 1, nil
}

func (e *Entity) GetMenuAll() (*[]Entity, error) {
	var rList []Entity
	db := model.DB().Find(&rList)
	return &rList, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) GetMenuByPId(pid int64) ([]Entity, error) {

	var menu []Entity
	db := model.DB()
	db.Where("pid = ?", pid).Find(&menu)
	if len(menu) > 0 {
		return menu, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}
