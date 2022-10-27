package global

import (
	"custody-merchant-admin/module/log"
	"fmt"
	"github.com/pkg/errors"
)

const (
	MsgWarnJwtErr         = "认证已经失效，请重新登录"
	MsgWarnMoreSend       = "验证码发送频繁,请稍后重试"
	MsgWarnAccountState   = "该账号已被冻结"
	MsgWarnNoEmailPhone   = "数据错误，手机号和邮箱为空"
	MsgWarnHaveEmail      = "数据错误，邮箱已被注册"
	MsgWarnHavePhone      = "数据错误，手机号被注册"
	MsgWarnUserTagIsNil   = "用户按钮权限为空"
	MsgWarnAccountIsNil   = "账号为空"
	MsgWarnCodeIsNil      = "验证码为空"
	MsgWarnPasswordNoTrue = "两次密码不一样"
	MsgWarnPasswordIsNil  = "密码为空"
	MsgWarnSendErr        = "发送失败"
	MsgWarnEmailFomat     = "邮箱格式错误"
	MsgWarnPidIsNil       = "父级信息为空"
	MsgWarnReSetPassword  = "密码重置失败"
	MsgWarnClearErr       = "清除数据失败"
)

func NewError(format string, v ...interface{}) error {
	var err error
	err = fmt.Errorf(format, v)
	log.Error(err.Error())
	return err
}

func DaoError(err error) error {
	log.Error(err.Error())
	return err
}

func WarnMsgError(msg string) error {
	log.Warn(msg)
	return errors.New(msg)
}
