package controller

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/internal/service/admin"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/router/web/handler"
	"strings"
)

// UpdateOurInfo
// 更新账号信息
func UpdateOurInfo(c *handler.Context) error {
	u := new(domain.SaveUserInfo)
	err := c.DefaultBinder(u)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	user := c.GetTokenUser()
	u.Id = user.Id
	err = service.UpdateOurInfo(u)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	res := handler.NewSuccess()
	return res.ResultOk(c)
}

// UpdateOurPassword
// 更新账号密码
func UpdateOurPassword(c *handler.Context) error {

	u := new(domain.RePwd)
	err := c.DefaultBinder(u)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	user := c.GetTokenUser()
	u.Account = user.Account
	// 校验传参
	if u.Account == "" {
		return global.WarnMsgError(global.MsgWarnAccountIsNil)
	}
	if u.Code == "" {
		return global.WarnMsgError(global.MsgWarnCodeIsNil)
	}
	if u.RePassword != u.Password {
		return global.WarnMsgError(global.MsgWarnPasswordNoTrue)
	}
	// 判断邮箱或者手机号
	key := global.GetCacheKey(global.PhoneSmsResetPwd, u.Account)
	isPhone := true
	if strings.Contains(u.Account, "@") {
		isPhone = false
		key = global.GetCacheKey(global.EmailResetPwd, u.Account)
	}
	dos, err := admin.ValiDataAccountService(u.Account)
	if !dos {
		return handler.NewError(c, global.MsgWarnAccountErr)
	}
	// 密码重设
	_, r := admin.ResetPasswordService(u.Account, key, u.Password, u.Code, isPhone)
	if r != nil {
		return handler.NewCodeError(c, 417, r.Error())
	}

	// 返回结果
	res := handler.NewResult(200, global.MsgSuccessReSetPassword)
	return res.ResultOk(c)
}

func ClearRedis(c *handler.Context) error {
	ri := new(domain.RedisKeyInfo)
	c.DefaultBinder(ri)
	cache.GetRedisClientConn().Del(ri.Key)
	res := handler.NewResult(200, "已经清理")
	return res.ResultOk(c)
}

func GetRedisByKey(c *handler.Context) error {
	ri := new(domain.RedisKeyInfo)
	c.DefaultBinder(ri)
	cents := ""
	cache.GetRedisClientConn().Get(ri.Key, &cents)
	res := handler.NewSuccess()
	res.AddData("redis", cents)
	return res.ResultOk(c)
}
