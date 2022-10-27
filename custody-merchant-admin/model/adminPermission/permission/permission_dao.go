package permission

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"custody-merchant-admin/module/log"
)

func (e *Entity) InsertNewUserPermission() (err error) {
	err = e.Db.Table(e.TableName()).Save(e).Error
	if err != nil {
		log.Errorf("SavePackageInfo error: %v", err)
	}
	return
}

// GetPermissionByUid
// 判断系统权限菜单Id
func (e *Entity) GetPermissionByUid(id int64) (*Entity, error) {
	sysRole := new(Entity)
	db := model.DB().Where("uid = ?", id).First(sysRole)
	if sysRole != nil && sysRole.Id != 0 {
		return sysRole, model.ModelError(db, global.MsgWarnModelNil)
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

// GetUserPermission
// 获取管理员访问菜单Id
func (e *Entity) GetUserPermission(uid int64) (*Entity, error) {
	auth := new(Entity)
	db := model.DB().Where("uid =? ", uid).First(auth)
	if auth != nil && auth.Id > 0 {
		return auth, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

// SaveUserPermission
// 获取管理员访问菜单Id
func (e *Entity) SaveUserPermission(p *Entity) error {
	tx := model.DB().Begin()
	if err := tx.Save(p).Error; err != nil {
		log.Errorf("SaveUserPermission error: %v", err)
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
