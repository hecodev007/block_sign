package admin

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service/admin/adminDeal"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/model/adminPermission/api"
	"time"
)

// GetMenuByRoleService
// 通过角色Id获取菜单
func GetMenuByRoleService(rid int) []*domain.TreeList {
	// 获取菜单
	return adminDeal.GetMenu(0, rid)
}

// GetNewMenuByRoleService
// 通过角色Id获取菜单
func GetNewMenuByRoleService(user *domain.JwtCustomClaims) []*domain.TreeList {
	mids, err := adminDeal.GetSysRolesByUid(user.Id)
	if err != nil {
		return nil
	}
	uInfo, err := adminDeal.GetUserInfoByUserId(user.Id)
	if err != nil {
		return nil
	}
	// 获取菜单
	return adminDeal.GetNewMenu(0, uInfo.Role, user.Admin, mids)
}

// GetRouterByTag
// 通过权限标识获取权限路径
func GetRouterByTag(tag string) (*api.Entity, error) {
	return adminDeal.GetRouterByTag(tag)
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

// FindMenuByUId
// 通过角色Id获取菜单
func FindMenuByUId(user *domain.JwtCustomClaims) ([]string, error) {
	return adminDeal.GetSysRolesTagByUid(user)
}
