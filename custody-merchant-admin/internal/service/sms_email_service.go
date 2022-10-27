package service

import (
	. "custody-merchant-admin/config"
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service/deals"
	"custody-merchant-admin/middleware/cache"
	"custody-merchant-admin/module/emails"
	"custody-merchant-admin/module/sms"
	"errors"
	"time"
)

func SendPhoneSmsOrSnsService(u *domain.AccountInfo, key, msg string) (bool, error) {
	var (
		send bool
		err  error
	)
	dataPhone, err := deals.ValiDataPhone(u.Account)
	if err != nil {
		return false, err
	}
	if dataPhone == nil || dataPhone.Id == 0 {
		return false, global.WarnMsgError("手机号不存在")
	}
	if u.PhoneCode != "" {
		dataPhone.PhoneCode = u.PhoneCode
		deals.UpdateAdminUserByUid(dataPhone.Id, map[string]interface{}{
			"phone_code": u.PhoneCode,
		})
	}
	if dataPhone.PhoneCode == "" {
		dataPhone.PhoneCode = "+86"
	}
	message, code := sms.GetRands(msg)
	//m := Conf.Sms["inland"]
	//phone := u.Account
	pCode, err := deals.FindPhoneCode(dataPhone.PhoneCode)
	if err != nil {
		return false, err
	}
	if pCode == nil {
		return false, global.WarnMsgError("手机区域不支持")
	}
	if pCode.Tag != "China" {
		//m = Conf.Sms["iso"]
		//phone = dataPhone.PhoneCode + " " + u.Account
		message = sms.GetEnRands(code)
		//send, err = sms.NewSms(m.AppKey, m.AppSecret, m.AppCode, m.Batch).SendSms(phone, message)
		//if err != nil {
		//	return send, err
		//}
	}
	sm := Conf.Wlwx
	sendSms := sms.WlwxEmailConfig{
		CustomName:   sm.CustomName,
		CustomPwd:    sm.CustomPwd,
		SmsClientUrl: sm.SmsClientUrl,
		Uid:          sm.Uid,
		Content:      message,
		DestMobiles:  u.Account,
		NeedReport:   true,
		SpCode:       sm.SpCode,
	}
	send, err = sendSms.SendWlwxSms()
	if err != nil {
		return send, err
	}
	k := global.GetCacheKey(key, u.Account)
	store := cache.GetCacheStore()
	err = store.Set(
		k,
		code,
		10*time.Minute,
	)
	if err != nil {
		return true, err
	}
	if dataPhone.PhoneCode == "" {
		_, err := deals.UpdateAdminUserByUid(dataPhone.Id, map[string]interface{}{"phone_code": u.PhoneCode})
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func SendEmailCodeService(email, key string) (bool, error) {

	message, code := sms.GetRandsAndEn()
	m := Conf.Email
	em := emails.EmailConfig{
		IamUserName:  m.IamUserName,
		Recipient:    email,
		SmtpUsername: m.SmtpUsername,
		SmtpPassword: m.SmtpPassword,
		Host:         m.Host,
		Port:         m.Port,
		Title:        m.Title,
	}
	dataEmail, err := deals.ValiDataEmail(email)
	if err != nil {
		return false, err
	}
	if dataEmail == nil || dataEmail.Id == 0 {
		return false, errors.New("账号无效")
	}
	sendEmail, err := em.SendEmail(message)
	if err != nil {
		return false, err
	}
	err = cache.GetCacheStore().Set(
		global.GetCacheKey(key, email),
		code,
		time.Duration(5)*time.Minute,
	)
	if err != nil {
		return sendEmail, err
	}
	return true, nil
}

func GetPhoneSmsCodeService(phone string) (string, error) {

	phones, err := deals.ValiDataPhone(phone)
	if err != nil {
		return "", err
	}
	return phones.PhoneCode, nil
}
