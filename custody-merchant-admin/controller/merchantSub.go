package controller

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/module/dict"
	"custody-merchant-admin/router/web/handler"
)

// GetSubUserList
// 人员管理
// 获取人员列表
func GetSubUserList(c *handler.Context) error {
	u := new(domain.SelectUserInfo)
	u.Account = c.QueryParam("account")
	u.MerchantId = c.SwitchType("merchant_id", "int64").(int64)
	u.Offset, u.Limit = c.OffsetPage()
	res := handler.NewSuccess()

	if u.MerchantId == 0 && u.Account == "" {
		res.AddData("merchant", map[string]interface{}{})
		res.AddData("list", []domain.SelectUserList{})
		res.AddData("total", 0)
		return res.ResultOk(c)
	}
	// 先查询商家信息
	mainInfo, err := service.GetMainUserInfoService(u)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	// 商户信息是否存在
	if mainInfo.Id <= 0 {
		return handler.NewError(c, global.DataWarnNoDataErr)
	}
	u.MerchantId = mainInfo.Id
	list, total, err := service.GetUserInfoListService(u)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res.AddData("merchant", mainInfo)
	res.AddData("list", list)
	res.AddData("total", total)
	return res.ResultOk(c)
}

// GetSubUserInfo
// 商户账号管理
// 根据Id获取人员
func GetSubUserInfo(c *handler.Context) error {
	u := new(domain.UserAccount)
	u.Id = c.SwitchType("id", "int64").(int64)
	subInfo, err := service.GetUserInfoById(u.Id)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("result", subInfo)
	return res.ResultOk(c)
}

// DelSubUserInfo
// 删除人员配置
func DelSubUserInfo(c *handler.Context) error {
	u := new(domain.SelectUserList)
	err := c.DefaultBinder(u)
	u.State = 2
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	user := c.GetTokenUser()
	if user.MerchantId == u.Id {
		return handler.NewError(c, global.MsgWarnUpdateYourSelf)
	}
	err = service.UpdateSubUserById(u)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	key := global.GetCacheKey(global.MenuTree, u.Id)
	cache.GetRedisClientConn().Del(key)
	res := handler.NewResult(200, global.MsgSuccessDel)
	return res.ResultOk(c)
}

// ClearSubInfoErr
// 一键清除异常
func ClearSubInfoErr(c *handler.Context) error {
	u := new(domain.UserAccountErr)
	err := c.DefaultBinder(u)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	mp := map[string]interface{}{}
	sl := new(domain.SelectUserInfo)
	sl.MerchantId = u.Id
	list, _, err := service.GetUserInfoListService(sl)
	for _, clear := range u.ClearErr {
		user := c.GetTokenUser()
		if clear == 1 {
			mp["pwd_err"] = 0
			for _, userList := range list {
				service.AddUserOperateUId(userList.Id, user.Name, "清除账号密码异常")
			}
		}
		if clear == 2 {
			mp["phone_code_err"] = 0
			for _, userList := range list {
				service.AddUserOperateUId(userList.Id, user.Name, "清除手机验证码异常")
			}
		}
		if clear == 3 {
			mp["email_code_err"] = 0
			for _, userList := range list {
				service.AddUserOperateUId(userList.Id, user.Name, "清除邮箱验证码异常")
			}
		}
	}
	err = service.UpdateClearUserByPId(u.Id, mp)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewResult(200, "清除成功")
	return res.ResultOk(c)
}

// ClearSubInfoErrById
// 清除异常
func ClearSubInfoErrById(c *handler.Context) error {

	u := new(domain.UserAccountErr)
	err := c.DefaultBinder(u)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	mp := map[string]interface{}{}

	if len(u.ClearErr) == 0 {
		return handler.NewResult(200, global.MsgWarnClearErr).ResultOk(c)
	}
	subInfo, err := service.GetDaoUserInfoById(u.MerchantId)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	if subInfo == nil || subInfo.Id <= 0 {
		return handler.NewError(c, global.MsgWarnAccountErr)
	}
	errNums := 0
	if subInfo.EmailCodeErr >= 5 {
		errNums += 1
	}
	if subInfo.PhoneCodeErr >= 5 {
		errNums += 1
	}
	if subInfo.PwdErr >= 5 {
		errNums += 1
	}
	for _, clear := range u.ClearErr {
		user := c.GetTokenUser()
		if clear == 1 {
			mp["pwd_err"] = 0
			errNums -= 1
			service.AddUserOperateUId(u.Id, user.Name, "清除账号密码异常")
		}
		if clear == 2 {
			mp["phone_code_err"] = 0
			errNums -= 1
			service.AddUserOperateUId(u.Id, user.Name, "清除手机验证码异常")
		}
		if clear == 3 {
			mp["email_code_err"] = 0
			errNums -= 1
			service.AddUserOperateUId(u.Id, user.Name, "清除邮箱验证码异常")

		}
	}
	if errNums <= 0 {
		mp["state"] = 0
		mp["reason"] = ""
	}
	err = service.UpdateClearUserById(u.MerchantId, mp)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewResult(200, "清除成功")
	return res.ResultOk(c)
}

// FreezeOrThawSubUserInfo
// 冻结/解冻人员配置
func FreezeOrThawSubUserInfo(c *handler.Context) error {
	u := new(domain.SelectUserList)
	err := c.DefaultBinder(u)
	user := c.GetTokenUser()
	if user.MerchantId == u.Id {
		return handler.NewError(c, global.MsgWarnUpdateYourSelf)
	}
	subInfo, err := service.GetUserInfoById(u.Id)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	if u.Reason == "" {
		u.Reason = "冻结账户"
	}
	if subInfo.Id > 0 {
		if subInfo.State == 1 {
			u.State = 0
			if u.Reason == "" {
				u.Reason = "解冻账户"
			}
			service.AddUserOperateUId(u.Id, user.Name, "解冻："+u.Reason)
		}
		if subInfo.State == 0 {
			u.State = 1
			service.AddUserOperateUId(u.Id, user.Name, "冻结："+u.Reason)
		}
	}

	err = service.UpdateSubUserById(u)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	key := global.GetCacheKey(global.MenuTree, u.Id)
	cache.GetRedisClientConn().Del(key)
	res := handler.NewResult(200, global.MsgSuccessDel)
	return res.ResultOk(c)
}

func UpdateSubUserInfo(ctx *handler.Context) error {
	u := new(domain.SaveUserInfo)
	err := ctx.DefaultBinder(u)
	if err != nil {
		return handler.NewError(ctx, err.Error())
	}
	user := ctx.GetTokenUser()
	if user.MerchantId == u.Id {
		return handler.NewError(ctx, global.MsgWarnUpdateYourSelf)
	}
	if dict.SysMerchantRoleTagList[u.Role-1] == "administrator" {
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
	err = service.UpdateSubUserInfo(u)
	if err != nil {
		return handler.NewError(ctx, err.Error())
	}
	service.AddUserOperateUId(u.Id, user.Name, "更新个人信息")
	key := global.GetCacheKey(global.AdminMenuTree, u.Id)
	bkey := global.GetCacheKey(global.AdminMenuBtn, u.Id)
	cache.GetRedisClientConn().Del(global.GetCacheKey(global.UserTokenVerify, u.Id))
	cache.GetRedisClientConn().Del(key)
	cache.GetRedisClientConn().Del(bkey)
	res := handler.NewResult(200, global.MsgSuccessUpdate)
	return res.ResultOk(ctx)
}

func FindUserOperateUId(c *handler.Context) error {
	u := new(domain.UserAccount)
	u.Id = c.SwitchType("id", "int64").(int64)
	lst, err := service.FindUserOperateUId(u.Id)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("list", lst)
	return res.ResultOk(c)
}
