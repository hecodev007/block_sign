package api

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
)

// GetSysAPI
// 获取管理员访问信息
func (e *Entity) GetSysAPI(name, path, method string) (*Entity, error) {
	auth := Entity{}
	db := model.DB().Where("name =? and path=? and method=?", name, path, method).First(&auth)
	if auth.Id != 0 {
		return &auth, model.ModelError(db, global.MsgWarnModelNil)
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

// FindSysAPIByTag
// 获取管理员访问信息
func (e *Entity) FindSysAPIByTag(tag string) ([]Entity, error) {
	var auth []Entity
	db := model.DB().Where("tag = ? ", tag).Find(&auth)
	return auth, model.ModelError(db, global.MsgWarnModelNil)
}

// GetSysAPIById
// 获取管理员访问信息
func (e *Entity) GetSysAPIById(ids string) ([]Entity, error) {
	var (
		auth []Entity
	)
	if ids != "" {
		db := model.DB().Where("id in (?)", ids).Find(&auth)
		return auth, model.ModelError(db, global.MsgWarnModelNil)
	}
	return nil, nil
}

// GetSysAPIByTag
// 获取访问信息
func (e *Entity) GetSysAPIByTag(tag string) (*Entity, error) {
	var auth = new(Entity)
	db := model.DB().Where("tag = ?", tag).First(auth)
	if auth != nil && auth.Id > 0 {
		return auth, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

// GetSysAPIByRole
// 根据角色获取管理员访问信息
func (e *Entity) GetSysAPIByRole(role int) ([]Entity, error) {
	var auth []Entity
	db := model.DB().Where("role =?", role).Find(&auth)
	return auth, model.ModelError(db, global.MsgWarnModelNil)
}

// GetSysAPIList
// 获取管理员访问列表
func (e *Entity) GetSysAPIList() ([]Entity, error) {
	var auths []Entity
	db := model.DB().Find(&auths)
	return auths, model.ModelError(db, global.MsgWarnModelNil)
}

// SaveSysAPI
// 新增管理员访问
func (e *Entity) SaveSysAPI(sr *Entity) error {
	tx := model.DB().Begin()
	tx.Create(sr)
	err := model.ModelError(tx, global.MsgWarnModelAdd)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// UpdateSysAPI
// 更新管理员访问
func (e *Entity) UpdateSysAPI(sr *Entity) error {
	db := model.DB().Save(sr)
	err := model.ModelError(db, global.MsgWarnModelUpdate)
	if err != nil {
		return err
	}
	return nil
}
