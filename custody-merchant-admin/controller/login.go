package controller

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/internal/service/admin"
	"custody-merchant-admin/middleware/cache"
	user "custody-merchant-admin/model/adminPermission/user"
	"custody-merchant-admin/module/auth"
	"custody-merchant-admin/module/dict"
	"custody-merchant-admin/router/web/handler"
	"custody-merchant-admin/util/xkutils"
	"strings"
	"time"
)

// LoginByPassword
// 账号密码登录
func LoginByPassword(c *handler.Context) error {
	u := new(domain.LoginInfo)
	err := c.DefaultBinder(u)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	// 校验传参
	if u.Password == "" {
		return global.WarnMsgError(global.MsgWarnPasswordIsNil)
	}
	login, err := admin.LoginService(u)

	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	dos, err := admin.CheckNewAccountService(u.Account)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	if login == nil {
		res := handler.NewResult(416, "账号或密码错误")
		res.AddData("login", false)
		return res.ResultOk(c)
	}
	res := handler.NewResult(200, "success")
	res.AddData("login", true)
	res.AddData("isFirst", dos)
	return res.ResultOk(c)

}

// LoginByCode
// 验证码登录
func LoginByCode(c *handler.Context) error {
	u := new(domain.UserAccountLogin)
	login := new(user.Entity)
	err := c.DefaultBinder(u)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	if u.Account == "" {
		res := handler.NewResult(416, "登录的账号为空")
		return res.ResultOk(c)
	}
	if strings.Contains(u.Account, "@") {
		login, err = admin.LoginEmailService(u)
	} else {
		login, err = admin.LoginSmsService(u)
	}
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	// Throws unauthorized error
	if login == nil {
		return handler.NewCodeError(c, 417, "账号不存在")
	}

	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}

	isAdmin := xkutils.ThreeDo(dict.SysRoleTagList[login.Role-1] == "administrator", true, false).(bool)
	rk := xkutils.RandString(16)
	// Set custom claims
	claims := &domain.JwtCustomClaims{
		Id:         login.Id,
		MerchantId: login.Uid,
		Name:       login.Name,
		Account:    login.Phone,
		Admin:      isAdmin,
		Role:       login.Role,
		Nonce:      rk,
	}
	t, err := auth.GetJWT(claims)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	err = cache.GetRedisClientConn().Set(global.GetCacheKey(global.UserTokenVerify, login.Id), t, time.Hour*8)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	err = admin.InitMenuByRole(claims)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	dos, err := admin.CheckNewAccountService(claims.Account)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	if dos {
		err = cache.GetCacheStore().Set(global.GetCacheKey(global.CheckAccount, claims.Id), dos, time.Hour*8)
		if err != nil {
			return handler.NewError(c, err.Error())
		}
	}
	res := handler.NewSuccess()
	res.AddData("token", t)
	res.AddData("nonce", rk)
	res.AddData("account", u.Account)
	res.AddData("name", login.Name)
	res.AddData("isFirst", dos)
	return res.ResultOk(c)
}

// ResetPwdAndLogin
// 重设密码并且登录
func ResetPwdAndLogin(c *handler.Context) error {
	u := new(domain.RePwd)
	err := c.DefaultBinder(u)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	if u.Password == "" {
		return global.WarnMsgError(global.MsgWarnPasswordIsNil)
	}
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
	password, r := admin.ResetPasswordService(u.Account, key, u.Password, u.Code, isPhone)
	if r != nil {
		return handler.NewCodeError(c, 417, r.Error())
	}
	// 重置密码成功
	// 调用登录
	linfo := &domain.LoginInfo{
		Account:  u.Account,
		Password: u.RePassword,
	}
	login, err := admin.LoginService(linfo)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	// Throws unauthorized error
	if login == nil {
		return handler.NewCodeError(c, 417, "账号或者密码错误")
	}
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	isAdmin := xkutils.ThreeDo(dict.SysRoleTagList[login.Role-1] == "administrator", true, false).(bool)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	// Set custom claims
	claims := &domain.JwtCustomClaims{
		Id:         login.Id,
		Name:       login.Name,
		MerchantId: login.Uid,
		Account:    login.Phone,
		Admin:      isAdmin,
		Role:       login.Role,
	}
	t, err := auth.GetJWT(claims)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	err = cache.GetRedisClientConn().Set(global.GetCacheKey(global.UserTokenVerify, login.Id), t, time.Hour*8)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	err = admin.InitMenuByRole(claims)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewResult(200, password)
	res.AddData("token", t)
	res.AddData("account", u.Account)
	res.AddData("name", login.Name)
	res.AddData("isFirst", false)
	return res.ResultOk(c)
}

// GetSalt
// 获取用户的盐值
func GetSalt(c *handler.Context) error {
	u := new(domain.AccountInfo)
	u.Account = c.QueryParam("account")
	u.PhoneCode = c.QueryParam("phone_code")

	str, err := admin.GetSaltStrService(u)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	res := handler.NewSuccess()
	res.AddData("salt", str)
	return res.ResultOk(c)
}

// NewPassword
// 第一次密码设置
func NewPassword(c *handler.Context) error {
	u := new(domain.NewPwd)
	err := c.Bind(u)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	us := c.GetTokenUser()
	if u.RePassword != u.Password {
		return handler.NewError(c, "两次密码不一样")
	} else {
		_, err := admin.NewPasswordService(us.Id, u.Password)
		if err != nil {
			return handler.NewError(c, err.Error())
		}
		key := global.GetCacheKey(global.CheckAccount, us.Id)
		cache.GetCacheStore().Delete(key)
		res := handler.NewResult(200, "密码设置成功")
		return res.ResultOk(c)
	}
}

// VerifyDataAccount
// 校验账号是否有效
func VerifyDataAccount(c *handler.Context) error {

	acc := new(domain.AccountInfo)
	err := c.DefaultBinder(acc)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	if acc.Account == "" {
		return handler.NewError(c, global.MsgWarnAccountErr)
	}
	dos, err := admin.ValiDataAccountService(acc.Account)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	check, err := admin.CheckNewAccountService(acc.Account)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	ac := acc.Account + global.MsgWarnAccountErr
	if dos {
		ac = acc.Account + "账号存在"
	}
	res := handler.NewResult(200, ac)
	res.AddData("account", dos)
	res.AddData("isFirst", check)
	return res.ResultOk(c)
}

// ResetPassword
// 重设密码
func ResetPassword(c *handler.Context) error {
	u := new(domain.RePwd)
	err := c.DefaultBinder(u)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
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
	password, r := admin.ResetPasswordService(u.Account, key, u.Password, u.Code, isPhone)
	if r != nil {
		return handler.NewCodeError(c, 417, r.Error())
	}
	cache.GetCacheStore().Delete(key)
	// 返回结果
	res := handler.NewResult(200, password)

	return res.ResultOk(c)
}

// CheckNewAccount
// 检查是否为新账号
func CheckNewAccount(c *handler.Context) error {

	user := c.GetTokenUser()
	dos, err := admin.CheckNewAccountService(user.Account)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	res := handler.NewSuccess()
	res.Data["account"] = dos
	return res.ResultOk(c)
}

// GetUserPersonal
// 获取个人信息
func GetUserPersonal(c *handler.Context) error {
	user := c.GetTokenUser()
	info, err := admin.GetAccountPersonaInfo(user.Id)
	if err != nil {
		return handler.NewCodeError(c, 417, err.Error())
	}
	if info == nil {
		return handler.NewCodeError(c, 416, global.MsgWarnSysNotUserInfo)
	}

	info.Id = user.Id
	dos, err := admin.CheckNewAccountService(user.Account)
	m := new(domain.MerchantService)
	m.Id = user.Id
	audit, err := service.FindSuperAudit(m)
	res := handler.NewSuccess()
	info.Account = user.Account
	res.Data["userInfo"] = info
	res.Data["merchantList"] = audit.HaveService
	res.Data["isFirst"] = dos
	return res.ResultOk(c)
}

// Logout
// 退出登录
func Logout(c *handler.Context) error {
	user := c.TokenUserOut()
	key := global.GetCacheKey(global.AdminMenuTree, user.Id)
	bkey := global.GetCacheKey(global.AdminMenuBtn, user.Id)
	cache.GetRedisClientConn().Del(key)
	cache.GetRedisClientConn().Del(bkey)
	res := handler.NewResult(200, user.Name+",退出登录成功")
	return res.ResultOk(c)
}
