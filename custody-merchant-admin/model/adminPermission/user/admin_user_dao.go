package api

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/model"
	"custody-merchant-admin/module/log"
	"errors"
	"fmt"
	"strings"
)

// SaveUser
// 新增用户
func (e *Entity) SaveUser(u *Entity) (int64, error) {
	tx := model.DB().Begin()
	if err := tx.Omit("deleted_at", "updated_at", "login_time").Create(u).Error; err != nil {
		log.Errorf("SaveUser error: %v", err)
		tx.Rollback()
		return u.Id, err
	}
	tx.Commit()
	return u.Id, nil
}

// UpdateUser
// 更新用户
func (e *Entity) UpdateUser(u *Entity) error {
	tx := model.DB().Begin()
	if err := tx.Omit("create_at", "login_time").Save(u).Error; err != nil {
		log.Errorf("SaveUser error: %v", err)
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// UpdatePersonalUser
// 更新用户
func (e *Entity) UpdatePersonalUser(id int64, mp map[string]interface{}) error {
	if err := model.DB().Table("admin_user").Where("id=?", id).Updates(mp).Error; err != nil {
		log.Errorf("UpdatePersonalUser error: %v", err)
		return err
	}
	return nil
}

// GetUserById
// 通过用户Id查询用户信息
func (e *Entity) GetUserById(id int64) (*Entity, error) {
	u := new(Entity)
	db := model.DB().Where("id = ? and state != 2", id).First(u)
	if u != nil && u.Id != 0 {
		return u, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

// GetUserByRId
// 通过用户RId查询用户信息
func (e *Entity) GetUserByRId(pid int64, rid int) ([]Entity, error) {
	var u []Entity
	sql := " select * from admin_user where admin_user.pid = ? and admin_user.roles = ? and admin_user.state != 2"
	db := model.DB().Raw(sql, pid, rid).Find(&u)
	return u, model.ModelError(db, global.MsgWarnModelNil)
}

// HaveUserId
// 通过用户Id查询用户信息
func (e *Entity) HaveUserId(id int64) (*Entity, error) {
	u := new(Entity)
	db := model.DB().Where("id = ?", id).First(u)
	if u != nil && u.Id != 0 {
		return u, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) GetSaltByPhoneAndEmail(phone, email string) (*Entity, error) {
	u := Entity{}
	db := model.DB()
	if model.FilteredSQLInject(phone, email) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	db.Where("state = 0 ")
	if phone != "" {
		db.Where(" and phone = ? ", phone)
	}
	if email != "" {
		db.Where(" and email = ? ", email)
	}
	d := db.First(&u)
	if u.Id != 0 {
		return &u, model.ModelError(d, global.MsgWarnModelNil)
	}
	return nil, model.ModelError(d, global.MsgWarnModelNil)
}

// GetUserByPhoneAndPwd
// 通过用户手机号、密码查询用户信息
func (e *Entity) GetUserByPhoneAndPwd(phone, pwd string) (*Entity, error) {
	u := new(Entity)
	if model.FilteredSQLInject(phone, pwd) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	db := model.DB().Where("phone = ? and password = ? and state = 0", phone, pwd).First(u)
	//Cache(db).First(&u)
	if u != nil && u.Id != 0 {
		return u, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelErrorAccount)
}

// GetUserByEmailAndPwd
// 通过用户邮箱、密码查询用户信息
func (e *Entity) GetUserByEmailAndPwd(email, pwd string) (*Entity, error) {
	u := new(Entity)
	if model.FilteredSQLInject(email, pwd) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	l := model.DB().Where("email = ? and password = ? and state = 0", email, pwd).First(u)
	if u != nil && u.Id != 0 {
		return u, nil
	}
	return nil, model.ModelError(l, global.MsgWarnModelErrorAccount)
}

// GetUserByEmail
// 通过用户邮箱查询用户信息
func (e *Entity) GetUserByEmail(email string) (*Entity, error) {
	u := Entity{}
	if model.FilteredSQLInject(email) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	l := model.DB().Where("email = ? and state != 2", email).First(&u)
	if u.Id != 0 {
		return &u, nil
	}
	return nil, model.ModelError(l, global.MsgWarnModelNil)
}

// GetUserByPhone
// 通过用户手机查询用户信息
func (e *Entity) GetUserByPhone(phone string) (*Entity, error) {
	u := Entity{}
	if model.FilteredSQLInject(phone) {
		return nil, errors.New(global.MsgWarnSqlInject)
	}
	l := model.DB().Where("phone = ? and state != 2", phone).First(&u)
	return &u, model.ModelError(l, global.MsgWarnModelNil)
}

// UpdatePwdById
// 通过Id更新密码
func (e *Entity) UpdatePwdById(id int64, pwd, salt string) error {
	if model.FilteredSQLInject(pwd, salt) {
		return fmt.Errorf(global.MsgWarnSqlInject)
	}
	db := model.DB().Table(e.TableName()).Where("id = ? and state = 0", id).Updates(map[string]interface{}{"password": pwd, "salt": salt})
	return model.ModelError(db, global.MsgWarnModelUpdate)
}

// UpdatePwdByPhone
// 通过Phone更新密码
func (e *Entity) UpdatePwdByPhone(phone string, mp map[string]interface{}) error {
	if model.FilteredSQLInject(phone) {
		return fmt.Errorf(global.MsgWarnSqlInject)
	}
	db := model.DB().Model(&Entity{}).Where("phone = ? and state = 0", phone).Updates(mp)
	return model.ModelError(db, global.MsgWarnModelUpdate)
}

// UpdatePwdByEmail
// 通过邮箱更新密码
func (e *Entity) UpdatePwdByEmail(email string, mp map[string]interface{}) error {
	if model.FilteredSQLInject(email) {
		return fmt.Errorf(global.MsgWarnSqlInject)
	}
	db := model.DB().Model(&Entity{}).Where("email = ? and state = 0", email).Updates(mp)
	return model.ModelError(db, global.MsgWarnModelUpdate)
}

// UpdatePwdByRoles
// 更新用户角色信息
func (e *Entity) UpdatePwdByRoles(id int64, roles string) error {
	if model.FilteredSQLInject(roles) {
		return fmt.Errorf(global.MsgWarnSqlInject)
	}
	db := model.DB().Model(&Entity{}).Where("id = ? and state = 0", id).Update("roles", roles)
	return model.ModelError(db, global.MsgWarnModelUpdate)
}

// FindAdminUserInfoList
// 查询用户信息列表
func (e *Entity) FindAdminUserInfoList(uid int64, userSelect *domain.SelectUserInfo) ([]Entity, int64, error) {
	var (
		usi   []Entity
		count int64
	)
	if model.FilteredSQLInject(userSelect.Name) {
		return nil, 0, errors.New(global.MsgWarnSqlInject)
	}
	db := model.DB().Table("admin_user").Where("state !=2")
	if userSelect.Name != "" {
		db.Where(" name=? ", userSelect.Name)
	}
	if userSelect.Account != "" {
		if strings.Contains(userSelect.Account, "@") {
			db.Where("email=? ", userSelect.Account)
		} else {
			db.Where("phone=? ", userSelect.Account)
		}
	}
	if uid != 0 {
		db.Where(" (pid= ? or id = ? )", uid, uid)
	}
	db.Offset(userSelect.Offset).Limit(userSelect.Limit).Find(&usi).Offset(-1).Limit(-1).Count(&count)
	return usi, count, model.ModelError(db, global.MsgWarnModelNil)
}

// GetUserByIds
// 通过批量用户Id查询用户信息
func (e *Entity) GetUserByIds(ids []int64) ([]Entity, error) {
	var u []Entity
	db := model.DB().Where("id in (?) and state = 0", ids).Find(&u)
	if len(u) != 0 {
		return u, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

// UpdateUserById
// 通过用户Id更新用户
func (e *Entity) UpdateUserById(id int64, up map[string]interface{}) (int64, error) {

	db := model.DB().Model(&Entity{}).Where("id = ? ", id).Updates(up)
	return db.RowsAffected, model.ModelError(db, global.MsgWarnModelUpdate)
}

func (e *Entity) GetAdminPersonal(id int64) (*Entity, error) {
	u := new(Entity)
	db := model.DB().Table("admin_user").Where("id = ? ", id).First(u)
	if u != nil && u.Id > 0 {
		return u, nil
	} else {
		return nil, model.ModelError(db, global.MsgWarnModelNil)
	}
}

// HaveAdminUserByPIdAndUId
// 通过用户Id查询用户信息
func (e *Entity) HaveAdminUserByPIdAndUId(id, pid int64) (*Entity, error) {
	u := new(Entity)
	db := model.DB().Where("id = ? and pid = ? and state !=2", id, pid).First(u)
	if u != nil && u.Id != 0 {
		return u, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) GetAdminUserUId(uid int64) (*Entity, error) {
	u := new(Entity)
	db := model.DB().Table("admin_user").Where("uid = ? ", uid).First(u)
	if u != nil && u.Id > 0 {
		return u, nil
	} else {
		return u, model.ModelError(db, global.MsgWarnModelNil)
	}
}
