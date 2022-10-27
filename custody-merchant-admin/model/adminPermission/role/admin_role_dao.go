package role

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
)

// GetAdminRole
// 获取系统角色信息
func (e *Entity) GetAdminRole(id int) (*Entity, error) {
	adminRole := Entity{}
	db := model.DB().Where("id = ? ", id).First(&adminRole)
	if adminRole.Id != 0 {
		return &adminRole, model.ModelError(db, global.MsgWarnModelNil)
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

// GetAdminRoleTag
// 获取系统角色信息
func (e *Entity) GetAdminRoleTag(tag string) (*Entity, error) {
	adminRole := Entity{}
	db := model.DB().Where("tag = ? ", tag).First(&adminRole)
	if adminRole.Id != 0 {
		return &adminRole, model.ModelError(db, global.MsgWarnModelNil)
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

// GetAdminRoleByIds
// 获取系统角色信息
func (e *Entity) GetAdminRoleByIds(ids []int) ([]Entity, error) {
	var adminRole []Entity
	db := model.DB().Where("id in (?) ", ids).Find(&adminRole)
	return adminRole, model.ModelError(db, global.MsgWarnModelNil)
}

// GetAdminRoleIsAdmin
// 判断系统角色
func (e *Entity) GetAdminRoleIsAdmin(id int) (*Entity, error) {

	adminRole := new(Entity)
	db := model.DB().Where("id = ?", id).First(adminRole)

	if adminRole.Id != 0 {
		return adminRole, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

// GetAdminRoleList
// 查询所有的角色
func (e *Entity) GetAdminRoleList() ([]Entity, error) {
	var adminRole []Entity
	db := model.DB().Order("id asc").Find(&adminRole)
	return adminRole, model.ModelError(db, global.MsgWarnModelNil)
}

// GetAdminRoleNoSuperList
// 查询所有的角色,超级管理员除外
func (e *Entity) GetAdminRoleNoSuperList(rid int) ([]Entity, error) {
	var adminRole []Entity
	db := model.DB().Table("admin_role").Where("id > ?", rid).Order("id asc").Find(&adminRole)
	return adminRole, model.ModelError(db, global.MsgWarnModelNil)
}

// GetAdminRoleIsSuperAdmin
// 判断系统角色
func (e *Entity) GetAdminRoleIsSuperAdmin(id int, tag string) (bool, error) {

	adminRole := new(Entity)
	db := model.DB().Where("id =? and tag =? ", id, tag).First(adminRole)
	if adminRole != nil && adminRole.Id != 0 {
		return true, nil
	}
	return false, model.ModelError(db, global.MsgWarnModelNil)

}

// GetAdminRoleAllByRid
// 根据用户角色获取系统角色
func (e *Entity) GetAdminRoleAllByRid(rid int) ([]Entity, error) {

	var adminRole []Entity
	db := model.DB().Where("id > ? ", rid).Find(&adminRole)
	if len(adminRole) != 0 {
		return adminRole, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)

}
