package adminDeal

import (
	"custody-merchant-admin/db"
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service/base"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/model/adminPermission/api"
	modelMenu "custody-merchant-admin/model/adminPermission/menu"
	"custody-merchant-admin/model/adminPermission/permission"
	"custody-merchant-admin/util/xkutils"
	"fmt"
	"strings"
	"time"
)

// GetMenu
// 递归获取树形菜单
func GetMenu(pid int64, roleId int) []*domain.TreeList {
	var (
		menus    = modelMenu.NewEntity()
		treeList []*domain.TreeList
	)
	m, err := menus.GetMenuByRIdPId(pid, roleId)
	if err != nil {
		return nil
	}
	for _, v := range m {
		child := GetMenu(v.Id, roleId)
		node := &domain.TreeList{
			Id:    v.Id,
			Label: v.Label,
			Sort:  v.Sort,
			Path:  v.Path,
			Pid:   v.Pid,
			Icon:  v.Icon,
		}
		node.Children = child
		treeList = append(treeList, node)
	}
	return treeList
}

// GetNewMenu
// 递归获取树形菜单
func GetNewMenu(pid int64, rid int, admin bool, ms string) []*domain.TreeList {
	var (
		treeList []*domain.TreeList
		menu     []modelMenu.Entity
		menus    = modelMenu.NewEntity()
	)

	if admin {
		m, err := menus.GetAdminMenuByPId(pid)
		if err != nil {
			return nil
		}
		menu = m
	} else {
		m, err := menus.GetNewMenuByPid(pid, rid, ms)
		if err != nil {
			return nil
		}
		menu = m
	}

	for _, v := range menu {
		if v.Id == 5 {
			fmt.Printf("1")
		}
		child := GetNewMenu(v.Id, rid, admin, ms)
		node := &domain.TreeList{
			Id:         v.Id,
			Label:      v.Label,
			Sort:       v.Sort,
			Path:       v.Path,
			Pid:        v.Pid,
			Icon:       v.Icon,
			ActiveMenu: v.ActiveMenu,
			Hidden:     v.Hidden,
		}
		node.Children = child
		treeList = append(treeList, node)
	}
	return treeList
}

func SaveAdminUserMenu(uid int64, menus []int) error {
	m := modelMenu.NewEntity()
	uPerm := permission.NewEntity()
	// 添加访问权限
	err := casbinMenus(uid, menus)
	if err != nil {
		return err
	}
	hidden, err := m.GetMenuByHidden()
	if err != nil {
		return err
	}
	for _, entity := range hidden {
		menus = append(menus, int(entity.Id))
	}
	mlist := xkutils.IntListToStrList(menus)
	mids := strings.Join(mlist, ",")
	up := &permission.Entity{
		Uid: uid,
		Mid: mids,
	}
	permission, err := uPerm.GetUserPermission(uid)
	if err != nil {
		return err
	}
	if permission != nil {
		up.Id = permission.Id
	}
	// 添加菜单
	return uPerm.SaveUserPermission(up)
}

func GetSysRolesByUid(uid int64) (string, error) {
	dao := permission.NewEntity()
	byUid, err := dao.GetPermissionByUid(uid)
	if err != nil {
		return "", err
	}
	if byUid != nil {
		return byUid.Mid, nil
	}
	return "", nil
}

func GetRouterByTag(tag string) (*api.Entity, error) {
	dao := api.NewEntity()
	return dao.GetSysAPIByTag(tag)
}

// 菜单的访问权限处理
func casbinMenus(uid int64, menus []int) error {
	var (
		dao  = modelMenu.NewEntity()
		aDao = api.NewEntity()
		tags = []api.Entity{}
	)
	// 先删除
	err := db.DeleteRuleByV0(uid)
	if err != nil {
		return err
	}
	// 查询用户菜单的所有按钮权限
	mlist, err := dao.GetAdminMenuList(menus)
	if err != nil {
		return err
	}
	// 判断获取访问
	for i, _ := range mlist {
		rtag, err := aDao.FindSysAPIByTag(mlist[i].Tag)
		if err != nil {
			return err
		}
		tags = append(tags, rtag...)
	}
	b := &base.CasbinService{}
	err = b.CasbinCreateBatch(fmt.Sprintf("%d", uid), tags)
	if err != nil {
		return err
	}
	return nil
}

func GetSysRolesTagByUid(user *domain.JwtCustomClaims) ([]string, error) {
	var (
		tagMap []string
		mlist  []modelMenu.Entity
		dao    = modelMenu.NewEntity()
		pDao   = permission.NewEntity()
		err    error
	)
	uid := user.Id
	if !user.Admin {
		rolesByUid, err := pDao.GetPermissionByUid(uid)
		if err != nil {
			return nil, err
		}
		byString, err := xkutils.IntSplitByString(rolesByUid.Mid, ",")
		if err != nil {
			return nil, err
		}
		if len(byString) == 0 {
			return nil, global.WarnMsgError(global.MsgWarnUserTagIsNil)
		}
		mlist, err = dao.GetAdminMenuList(byString)
		if err != nil {
			return nil, err
		}
	} else {
		mlist, err = dao.GetSysAllMenuTagList()
		if err != nil {
			return nil, err
		}
	}
	for i, _ := range mlist {
		tagMap = append(tagMap, mlist[i].Tag)
	}
	return tagMap, nil
}

// InitMenuByRole
// 通过角色Id获取菜单
func InitMenuByRole(user *domain.JwtCustomClaims) error {
	key := global.GetCacheKey(global.AdminMenuTree, user.Id)
	menuTree := GetNewMenuByRoleService(user)
	if len(menuTree) > 0 {
		cache.GetRedisClientConn().Set(key, menuTree, 8*time.Hour)
	}
	return nil
}

// GetNewMenuByRoleService
// 通过角色Id获取菜单
func GetNewMenuByRoleService(user *domain.JwtCustomClaims) []*domain.TreeList {
	mids, err := GetSysRolesByUid(user.Id)
	if err != nil {
		return nil
	}
	uInfo, err := GetUserInfoByUserId(user.Id)
	if err != nil {
		return nil
	}
	// 获取菜单
	return GetNewMenu(0, uInfo.Role, user.Admin, mids)
}
