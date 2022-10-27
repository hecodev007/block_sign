package controller

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service/admin"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/router/web/handler"
	"custody-merchant-admin/util/xkutils"
	"time"
)

// GetMenuByRole
// @Tags 登录操作
// @Summary 获取菜单
// @Description 获取菜单
// @Accept  json
// @param Authorization header string true "验证参数Bearer token"
// @param X-Ca-Nonce header string false "随机数不可重复"
// @Success 200 {object} ResultData "成功"
// @Failure 500	"失败"
// @Router /custody/getMenu [get]
func GetMenuByRole(c *handler.Context) error {
	var (
		menuTree []*domain.TreeList
		btns     []string
		err      error
	)
	user := c.GetTokenUser()

	key := global.GetCacheKey(global.AdminMenuTree, user.Id)
	cache.GetRedisClientConn().Get(key, &menuTree)
	if menuTree == nil {
		menuTree = admin.GetNewMenuByRoleService(user)
		err := cache.GetRedisClientConn().Set(key, menuTree, 8*time.Hour)
		if err != nil {
			return handler.NewCodeError(c, 417, err.Error())
		}
	}
	bkey := global.GetCacheKey(global.AdminMenuBtn, user.Id)
	cache.GetRedisClientConn().Get(bkey, &btns)
	if btns == nil {
		btns, err = admin.FindMenuByUId(user)
		if err != nil {
			return handler.NewCodeError(c, 417, err.Error())
		}
		err := cache.GetRedisClientConn().Set(bkey, btns, 8*time.Hour)
		if err != nil {
			return handler.NewCodeError(c, 417, err.Error())
		}
	}
	res := handler.NewSuccess()
	res.AddData("list", menuTree)
	res.AddData("btnList", btns)
	return res.ResultOk(c)
}

// GetBaseMenuByRid
// 根据Rid获取基本的菜单
func GetBaseMenuByRid(c *handler.Context) error {

	// 获取前端传回的角色
	rId := xkutils.StrToInt(c.QueryParam("rid"))
	// 根据角色进行菜单查询
	key := global.GetCacheKey(global.MenuTreeRole, rId)
	var menuTree []*domain.TreeList
	cache.GetRedisClientConn().Get(key, &menuTree)
	if menuTree == nil {
		menuTree = admin.GetMenuByRoleService(rId)
		err := cache.GetRedisClientConn().Set(key, menuTree, 8*time.Hour)
		if err != nil {
			return handler.NewCodeError(c, 417, err.Error())
		}
	}
	res := handler.NewSuccess()
	res.AddData("list", menuTree)
	return res.ResultOk(c)
}

// GetUserAllBtnTag
// 获取用户的全部权限标识别
func GetUserAllBtnTag(c *handler.Context) error {
	user := c.GetTokenUser()
	allMenu, err := admin.FindMenuByUId(user)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("list", allMenu)
	return res.ResultOk(c)
}
