package controller

import (
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/middleware/rabbitmq"
	"custody-merchant-admin/router/web/handler"
	"custody-merchant-admin/util/xkutils"
	"fmt"
	"strings"
	"time"
)

// SendLoginCode
// @Tags 验证码发送
// @Summary 发送登录短信验证码
// @Description 发送登录短信验证码
// @Accept  json
// @param X-Ca-Nonce header string false "随机数不可重复"
// @Param body body domain.AccountInfo true "传入参数"
// @Success 200 {object} ResultData "成功"
// @Failure 500	"失败"
// @Router /send/loginCode [post]
func SendLoginCode(c *handler.Context) error {
	u := new(domain.AccountInfo)
	send := false
	err := c.DefaultBinder(u)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	have := false
	cache.GetCacheStore().Get(global.GetCacheKey(global.SendLoginSec, u.Account), &have)
	if have {
		return handler.NewError(c, global.MsgWarnMoreSend)
	}
	if strings.Contains(u.Account, "@") {
		if !xkutils.VerifyEmailFormat(u.Account) {
			return handler.NewError(c, global.MsgWarnEmailFomat)
		}
		send, err = service.SendEmailCodeService(u.Account, global.EmailLogin)
	} else {
		send, err = service.SendPhoneSmsOrSnsService(u, global.PhoneSmsLogin, "登录")
	}
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	if !send {
		return handler.NewError(c, global.MsgWarnSendErr)
	}
	err = cache.GetCacheStore().Set(global.GetCacheKey(global.SendLoginSec, u.Account), true, time.Minute)
	if err != nil {
		return err
	}
	res := handler.NewResult(200, u.Account+"发送成功")
	return res.ResultOk(c)
}

// SendResetPwd
// @Tags 验证码发送
// @Summary 手机或者邮箱发送短信验证码
// @Description 手机或者邮箱发送短信验证码
// @Accept  json
// @param X-Ca-Nonce header string false "随机数不可重复"
// @Param body body domain.AccountInfo true "传入重置的账号"
// @Success 200 {object} ResultData "成功"
// @Failure 500	"失败"
// @Router /send/resetPwd [post]
func SendResetPwd(c *handler.Context) error {
	u := new(domain.AccountInfo)
	msg := ""
	send := false
	err := c.DefaultBinder(u)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	have := false
	cache.GetCacheStore().Get(global.GetCacheKey(global.ResetPwdSendSec, u.Account), &have)
	if have {
		return handler.NewError(c, "验证码发送频繁,请稍后重试")
	}
	if strings.Contains(u.Account, "@") {
		send, err = service.SendEmailCodeService(u.Account, global.EmailResetPwd)
		msg = u.Account + "邮箱：" + u.Account
	} else {
		send, err = service.SendPhoneSmsOrSnsService(u, global.PhoneSmsResetPwd, "重置密码")
		msg = "手机：" + u.Account
	}
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	if !send {
		return handler.NewError(c, msg+"，发送失败")
	}
	err = cache.GetCacheStore().Set(global.GetCacheKey(global.ResetPwdSendSec, u.Account), true, time.Minute*2)
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	res := handler.NewResult(200, msg+"，发送成功")
	return res.ResultOk(c)
}

// SendLimitCode
// @Tags 验证码发送
// @Summary 发送限制转帐的短信验证
// @Description 发送限制转帐的短信验证
// @Accept  json
// @param Authorization header string true "验证参数Bearer token"
// @param X-Ca-Nonce header string false "随机数不可重复"
// @Success 200 {object} ResultData "成功"
// @Failure 500	"失败"
// @Router /custody/limit/code [post]
func SendLimitCode(c *handler.Context) error {
	var err error
	send := false

	user := c.GetTokenUser()
	if strings.Contains(user.Account, "@") {
		send, err = service.SendEmailCodeService(user.Account, global.EmailLimitCode)
	} else {
		code, err := service.GetPhoneSmsCodeService(user.Account)
		if err != nil {
			return handler.NewError(c, err.Error())
		}
		send, err = service.SendPhoneSmsOrSnsService(&domain.AccountInfo{Account: user.Account, PhoneCode: code}, global.PhoneLimitCode, "同地址提币限制")
	}
	if err != nil {
		return handler.NewError(c, err.Error())
	}
	if !send {
		return handler.NewError(c, "发送失败")
	}
	str := "验证码已经发送至" + user.Account
	res := handler.NewResult(200, str)
	return res.ResultOk(c)
}

func MQTest(c *handler.Context) error {
	rmq := rabbitmq.NewMQ(rabbitmq.DefaultMQConfig)
	rmq.ConsumeSimple(func(body []byte) {
		fmt.Println(string(body))
	})
	rmq.PublishSimple("22222222")
	rmq.PublishSimple("you know")
	return nil
}
