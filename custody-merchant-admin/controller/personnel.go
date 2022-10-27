package controller

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/module/dict"
	"custody-merchant-admin/router/web/handler"
	"custody-merchant-admin/util/xkutils"
)

// GetUserList
// 人员管理
// 获取人员列表
func GetUserList(c *handler.Context) error {
	u := new(domain.SelectUserInfo)
	u.Name = c.QueryParam("name")
	u.Account = c.QueryParam("account")
	u.Offset, u.Limit = c.OffsetPage()
	list, total, err := service.GetAdminUserInfoListService(u, c.GetTokenUser())
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("list", list)
	res.AddData("total", total)
	return res.ResultOk(c)
}

func AddAdminUserInfo(ctx *handler.Context) error {
	u := new(domain.SaveUserInfo)
	err := ctx.DefaultBinder(u)
	if err != nil {
		return handler.NewError(ctx, err.Error())
	}
	user := ctx.GetTokenUser()
	if dict.SysRoleTagList[u.Role-1] == "administrator" {
		return handler.NewCodeError(ctx, 417, global.MsgWarnUpdateSuper)
	}
	if !user.Admin {
		if user.Id != u.Pid {
			return handler.NewCodeError(ctx, 417, global.OperationWarn)
		}
		if u.Role <= 2 {
			return handler.NewCodeError(ctx, 417, global.OperationWarn)
		}
		if user.Id == u.Id {
			return handler.NewCodeError(ctx, 417, global.MsgWarnUpdateYourSelf)
		}
	}
	err = service.AddAdminUserInfo(u)
	if err != nil {
		return handler.NewError(ctx, err.Error())
	}
	res := handler.NewResult(200, global.MsgSuccessAdd)
	return res.ResultOk(ctx)
}

func UpdateAdminUserInfo(ctx *handler.Context) error {
	u := new(domain.SaveUserInfo)
	err := ctx.DefaultBinder(u)
	if err != nil {
		return handler.NewError(ctx, err.Error())
	}
	user := ctx.GetTokenUser()
	if dict.SysRoleTagList[u.Role-1] == "administrator" {
		return handler.NewCodeError(ctx, 417, global.MsgWarnUpdateSuper)
	}
	if !user.Admin {
		if user.Id != u.Pid {
			return handler.NewCodeError(ctx, 417, global.OperationWarn)
		}
		if u.Role <= 2 {
			return handler.NewCodeError(ctx, 417, global.OperationWarn)
		}
		if user.Id == u.Id {
			return handler.NewCodeError(ctx, 417, global.MsgWarnUpdateYourSelf)
		}
	}
	err = service.UpdateAdminUserInfo(u)
	if err != nil {
		return handler.NewError(ctx, err.Error())
	}
	key := global.GetCacheKey(global.AdminMenuTree, u.Id)
	bkey := global.GetCacheKey(global.AdminMenuBtn, u.Id)
	cache.GetRedisClientConn().Del(global.GetCacheKey(global.UserTokenVerify, u.Id))
	cache.GetRedisClientConn().Del(key)
	cache.GetRedisClientConn().Del(bkey)
	res := handler.NewResult(200, global.MsgSuccessUpdate)
	return res.ResultOk(ctx)
}

// DelAdminUserInfo
// 删除人员配置
func DelAdminUserInfo(c *handler.Context) error {
	u := new(domain.SelectUserList)
	err := c.DefaultBinder(u)
	if err != nil {
		return handler.NewError(c, err.Error())
	}

	user := c.GetTokenUser()
	if u.Id == user.Id {
		return handler.NewError(c, global.MsgWarnUpdateYourSelf)
	}
	if user.Admin && u.Id == user.Id {
		return handler.NewError(c, global.MsgWarnDelSuper)
	}
	if !user.Admin && u.Id == 1 {
		return handler.NewError(c, global.MsgWarnSqlUpdate)
	}
	if !user.Admin {
		err = service.HaveAdminUserByPIdAndUId(u.Id, user.Id)
		if err != nil {
			return handler.NewCodeError(c, 417, err.Error())
		}
	}
	u.State = 2
	total, err := service.DelAdminUserInfo(u)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	if total == 0 {
		return handler.NewError(c, err.Error())
	}
	key := global.GetCacheKey(global.AdminMenuTree, user.Id)
	cache.GetRedisClientConn().Del(key)
	res := handler.NewResult(200, global.MsgSuccessDel)
	return res.ResultOk(c)
}

// UpdateAdminUserState
// 更新账号状态
func UpdateAdminUserState(c *handler.Context) error {
	u := new(domain.SelectUserList)
	err := c.DefaultBinder(u)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	user := c.GetTokenUser()
	if u.Id == user.Id {
		return handler.NewError(c, global.MsgWarnUpdateYourSelf)
	}
	if user.Admin && u.Id == user.Id {
		return handler.NewError(c, global.MsgWarnUpdateSuper)
	}
	if !user.Admin && u.Id == 1 {
		return handler.NewError(c, global.MsgWarnUpdateSuper)
	}
	if u.State < 0 && u.State > 2 {
		return handler.NewError(c, global.OperationWarn)
	}

	if !user.Admin {
		err = service.HaveAdminUserByPIdAndUId(u.Id, user.Id)
		if err != nil {
			return handler.NewCodeError(c, 417, err.Error())
		}
	}
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}

	total, err := service.UpdateUserState(u.Id, u)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	msg := "success"
	if total == 0 {
		msg = "更新0条数据"
	}
	key := global.GetCacheKey(global.AdminMenuTree, user.Id)
	cache.GetRedisClientConn().Del(key)

	if u.State == 2 {
		msg = global.MsgSuccessDel
	}
	if u.State == 1 {
		msg = global.MsgSuccessFreeze

	}
	if u.State == 0 {
		msg = global.MsgSuccessThaw
	}
	res := handler.NewResult(200, msg)
	return res.ResultOk(c)
}

// SaveSuperAudit
// 修改添加超级审核员
func SaveSuperAudit(c *handler.Context) error {
	m := new(domain.MerchantService)
	err := c.DefaultBinder(m)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	err = service.SaveSuperAudit(m)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	res := handler.NewResult(200, global.MsgSuccessUpdate)
	return res.ResultOk(c)
}

// GetSuperAudit
// 获取超级审核员
func GetSuperAudit(c *handler.Context) error {
	m := new(domain.MerchantService)
	user := c.GetTokenUser()
	m.Id = c.SwitchType("id", "int64").(int64)
	audit, err := service.FindSuperAudit(m)
	if err != nil {
		return err
	}
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("result", audit)
	res.AddData("show", xkutils.ThreeDo(user.Id == m.Id, 0, 1).(int))
	return res.ResultOk(c)
}

// HaveUserId
// 查看用户Id是否存在
func HaveUserId(c *handler.Context) error {
	id := xkutils.StrToInt64(c.QueryParam("id"))
	err := service.HaveUserId(id)
	msg := "success"
	code := 200
	if err != nil {
		msg = err.Error()
		code = 416
	}
	if err == nil {
		msg = global.MsgIdRight
		code = 200
	}
	res := handler.NewResult(code, msg)
	return res.ResultOk(c)
}

// GetUserById
// 查看用户Id是否存在
func GetUserById(c *handler.Context) error {
	id := xkutils.StrToInt64(c.QueryParam("id"))
	user := c.GetTokenUser()
	info, err := service.GetAdminUserInfoById(id)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}

	res := handler.NewSuccess()
	res.AddData("result", info)
	res.AddData("show", xkutils.ThreeDo(user.Id == info.Id, 0, 1).(int))
	return res.ResultOk(c)
}
