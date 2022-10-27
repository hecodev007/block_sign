package api

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/model"
	"custody-merchant-admin/util/sql"
	"fmt"
)

type SysMenuTag struct {
	Tag string `json:"tag" gorm:"column:tag"`
}

// FindAdminMenuByTag
// 通过Tag获取菜单
func (e *Entity) FindAdminMenuByTag(tag string) ([]Entity, error) {
	var menus []Entity
	db := model.DB().Where("tag <> ?", tag).First(&menus)
	return menus, model.ModelError(db, global.MsgWarnModelNil)
}

// GetAdminMenuById
// 通过Id获取菜单
func (e *Entity) GetAdminMenuById(id int) (*Entity, error) {
	auth := Entity{}
	db := model.DB().Where("id =?", id).First(&auth)
	if auth.Id != 0 {
		return &auth, model.ModelError(db, global.MsgWarnModelNil)
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

// GetAdminMenuByTitle
// 通过菜单名获取菜单
func (e *Entity) GetAdminMenuByTitle(title string) (*Entity, error) {
	sm := Entity{}
	db := model.DB().Where("title =?", title).First(&sm)
	if sm.Id != 0 {
		return &sm, model.ModelError(db, global.MsgWarnModelNil)
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}
func (e *Entity) GroupMenuByTag(uId int64) (*[]SysMenuTag, error) {
	var rList []SysMenuTag
	sql := "select distinct sys_router.tag from user_auth left join sys_router on sys_router.id = user_auth.router_id where user_auth.user_id=? group by sys_router.tag"
	db := model.DB().Raw(sql, uId).Scan(&rList)
	return &rList, model.ModelError(db, global.MsgWarnModelNil)
}

// GetAdminMenuByPId
// 通过父级Id获取菜单
func (e *Entity) GetAdminMenuByPId(pid int64) ([]Entity, error) {
	var menu []Entity
	db := model.DB().Where("pid =? and menu_type = 0 ", pid).Order("sort").Find(&menu)
	return menu, model.ModelError(db, global.MsgWarnModelNil)
}

// GetAdminMenuList
// 获取菜单按钮访问列表
func (e *Entity) GetAdminMenuList(ids []int) ([]Entity, error) {
	var auths []Entity
	db := model.DB().Where("id in (?) and menu_type = 1", ids).Distinct("tag").Find(&auths)
	return auths, model.ModelError(db, global.MsgWarnModelNil)
}

// GetSysAllMenuTagList
// 获取全部菜单按钮访问列表
func (e *Entity) GetSysAllMenuTagList() ([]Entity, error) {
	var auths []Entity
	db := model.DB().Where("menu_type = 1").Distinct("tag").Find(&auths)
	return auths, model.ModelError(db, global.MsgWarnModelNil)
}

// SaveAdminMenu
// 新增菜单访问
func (e *Entity) SaveAdminMenu(menu *Entity) error {
	tx := model.DB().Begin()
	tx.Create(menu)
	err := model.ModelError(tx, global.MsgWarnModelAdd)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (e *Entity) GetMenuByRIdPId(pid int64, rid int) ([]Entity, error) {
	var menu []Entity
	db := model.DB()
	sql := "select admin_menu.* from admin_role_menu " +
		"left join admin_menu on admin_menu.id = admin_role_menu.mid " +
		"where admin_menu.hidden = 0 and admin_menu.pid = ? and admin_role_menu.rid = ? order by admin_menu.sort asc"
	db.Raw(sql, pid, rid).Scan(&menu)
	if len(menu) > 0 {
		return menu, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) GetMenuByHidden() ([]Entity, error) {
	var menu []Entity
	db := model.DB().Table("admin_menu").Where("admin_menu.hidden = 1").Find(&menu)
	if len(menu) > 0 {
		return menu, nil
	}
	return nil, model.ModelError(db, global.MsgWarnModelNil)
}

func (e *Entity) GetNewMenuByPid(pid int64, rId int, mids string) ([]Entity, error) {

	var (
		menu  []Entity
		build sql.SqlBuilder
	)
	db := model.DB()

	build.SqlAdd("select distinct admin_menu.* from admin_role_menu " +
		" left join admin_menu on admin_menu.id = admin_role_menu.mid " +
		" where admin_menu.pid = ? and admin_menu.menu_type = 0 ")
	if mids != "" {
		build.SqlAdd(fmt.Sprintf(" and admin_menu.id in (%s) ", mids))
	}
	build.SqlWhereVars(" and admin_role_menu.rid =? ", rId, rId != 0)
	db.Raw(build.ToSqlString(), pid).Order("admin_menu.sort asc").Find(&menu)
	return menu, model.ModelError(db, global.MsgWarnModelNil)
}

// UpDateAdminMenu
// 新增菜单访问
func (e *Entity) UpDateAdminMenu(menu *Entity) error {
	db := model.DB().Save(menu)
	err := model.ModelError(db, global.MsgWarnModelUpdate)
	if err != nil {
		return err
	}
	return nil
}
