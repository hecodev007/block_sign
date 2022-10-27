package admin

import (
	conf "custody-merchant-admin/config"
	"custody-merchant-admin/global"
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service/admin/adminDeal"
	"custody-merchant-admin/internal/service/deals"
	"custody-merchant-admin/middleware/cache"
	modelUser "custody-merchant-admin/model/adminPermission/user"
	user "custody-merchant-admin/model/adminPermission/user"
	"custody-merchant-admin/module/log"
	"custody-merchant-admin/util/library"
	"custody-merchant-admin/util/xkutils"
	"errors"
	"strings"
	"time"
)

func LoginService(u *domain.LoginInfo) (*user.Entity, error) {

	if u.Password == "" {
		return nil, global.WarnMsgError(global.MsgWarnPasswordNil)
	}
	isPhone := true
	if strings.Contains(u.Account, "@") {
		if !xkutils.VerifyEmailFormat(u.Account) {
			return nil, errors.New(global.MsgWarnEmailFormatErr)
		}
		isPhone = false
	}
	account := u.Account
	pwd, err := library.RSADecryptPassword(u.Password)
	if err != nil {
		log.Errorf(err.Error())
		return nil, errors.New(global.MsgWarnDecryptErr)
	}
	if isPhone {
		us, err := deals.ValiDataPhone(account)
		if err == nil && us.Id > 0 {
			acc, err := adminDeal.LoginPhone(account, library.EncryptSha256Password(pwd, us.Salt))
			if err != nil {
				return nil, err
			}
			if acc != nil {
				us.PwdErr = 0
				us.PhoneCodeErr = 0
				us.EmailCodeErr = 0
			} else {
				if us.PwdErr <= 5 {
					us.PwdErr += 1
				} else {
					us.State = 1
				}
			}
			if _, err = adminDeal.UpdateUserByUid(us.Id, map[string]interface{}{"pwd_err": us.PwdErr, "email_code_err": us.EmailCodeErr, "phone_code_err": us.PhoneCodeErr, "state": us.State}); err != nil {
				return nil, err
			}
			err = CheckAccountErr(us)
			if err != nil {
				return nil, err
			}
			return acc, err
		}
	}
	if !isPhone {
		us, _ := deals.ValiDataEmail(account)
		if us != nil && us.Id > 0 {
			acc, err := adminDeal.LoginEmail(account, library.EncryptSha256Password(pwd, us.Salt))
			if err != nil {
				return nil, err
			}
			if acc != nil {
				us.PwdErr = 0
				us.State = 0
			} else {
				if us.PwdErr <= 5 {
					us.PwdErr += 1
				} else {
					us.State = 1
				}
			}
			if _, err = adminDeal.UpdateUserByUid(us.Id, map[string]interface{}{"pwd_err": us.PwdErr, "state": us.State}); err != nil {
				return nil, err
			}
			err = CheckAccountErr(us)
			if err != nil {
				return nil, err
			}
			return acc, err
		}
	}
	return nil, global.WarnMsgError(global.MsgWarnAccountErr)
}

func LoginSmsService(u *domain.UserAccountLogin) (*modelUser.Entity, error) {
	code := ""
	if u.Code == "" {
		return nil, global.WarnMsgError(global.MsgWarnCodeErr)
	}
	if u.Account == "" {
		return nil, global.WarnMsgError(global.MsgWarnPhoneFormatErr)
	}
	key := global.GetCacheKey(global.PhoneSmsLogin, u.Account)
	store := cache.GetCacheStore()
	store.Get(key, &code)
	dataPhone, err := deals.ValiDataPhone(u.Account)
	if err != nil {
		return nil, err
	}
	if dataPhone != nil && dataPhone.Id > 0 {
		if (code == u.Code || conf.EnvPro) && dataPhone.PhoneCodeErr <= 5 {
			AsUpdateUserByUid(dataPhone.Id, map[string]interface{}{"email_code_err": 0, "phone_code_err": 0, "pwd_err": 0, "state": 0, "login_time": time.Now().Local()})
			return dataPhone, nil
		} else {
			if dataPhone.PhoneCodeErr <= 5 {
				dataPhone.PhoneCodeErr = dataPhone.PhoneCodeErr + 1
			} else {
				dataPhone.State = 1
			}
			AsUpdateUserByUid(dataPhone.Id, map[string]interface{}{"phone_code_err": dataPhone.PhoneCodeErr, "state": dataPhone.State, "login_time": time.Now().Local()})
			err = CheckAccountErr(dataPhone)
			if err != nil {
				return nil, err
			}
			return nil, global.WarnMsgError(global.MsgWarnCodeErr)
		}
	} else {
		return nil, global.WarnMsgError(global.MsgWarnPhoneFormatErr)
	}
}

func LoginEmailService(u *domain.UserAccountLogin) (*modelUser.Entity, error) {
	code := ""
	if u.Code == "" {
		return nil, global.WarnMsgError(global.MsgWarnCodeErr)
	}
	if !xkutils.VerifyEmailFormat(u.Account) {
		return nil, global.WarnMsgError(global.MsgWarnEmailFormatErr)
	}
	cache.GetCacheStore().Get(global.GetCacheKey(global.EmailLogin, u.Account), &code)
	dataEmail, err := deals.ValiDataEmail(u.Account)
	if err != nil {
		return nil, err
	}
	if dataEmail != nil && dataEmail.Id > 0 {
		if (code == u.Code || conf.EnvPro) && dataEmail.EmailCodeErr <= 5 {
			AsUpdateUserByUid(dataEmail.Id, map[string]interface{}{"email_code_err": 0, "phone_code_err": 0, "pwd_err": 0, "state": 0, "login_time": time.Now().Local()})
			return dataEmail, nil
		} else {
			if dataEmail.EmailCodeErr <= 5 {
				dataEmail.EmailCodeErr = dataEmail.EmailCodeErr + 1
			} else {
				dataEmail.State = 1
			}
			AsUpdateUserByUid(dataEmail.Id, map[string]interface{}{"email_code_err": dataEmail.EmailCodeErr, "state": dataEmail.State, "login_time": time.Now().Local()})
			err = CheckAccountErr(dataEmail)
			if err != nil {
				return nil, err
			}
			return nil, global.WarnMsgError(global.MsgWarnCodeErr)
		}
	} else {
		return nil, global.WarnMsgError(global.MsgWarnAccountErr)
	}
}

// ResetPasswordService
// 重设密码
func ResetPasswordService(account, key, newPass, code string, isPhone bool) (string, error) {
	var c = ""
	err := cache.GetCacheStore().Get(key, &c)
	if err != nil {
		return "", err
	}
	if c == code {
		salt := xkutils.RandString(10)
		password, err := library.RSADecryptPassword(newPass)
		if err != nil {
			return "", err
		}

		pwd := library.EncryptSha256Password(password, salt)
		mp := map[string]interface{}{
			"password": pwd,
			"salt":     salt,
		}
		if isPhone {
			err = adminDeal.UpdatePwdByPhone(account, mp)
		} else {
			err = adminDeal.UpdatePwdByEmail(account, mp)
		}
		if err != nil {
			return "", err
		}
		return global.MsgSuccessReSetPassword, nil
	}
	return "error", global.WarnMsgError(global.MsgWarnCodeErr)
}

// NewPasswordService
// 设置密码
func NewPasswordService(id int64, newPass string) (string, error) {
	salt := xkutils.RandString(10)
	password, err := library.RSADecryptPassword(newPass)

	if err != nil {
		return "", err
	}
	if "HOO@2022" == password {
		return "", errors.New("新密码不能和原始密码相同")
	}
	pwd := library.EncryptSha256Password(password, salt)
	err = adminDeal.UpdatePwdById(id, pwd, salt)
	if err != nil {
		return "", err
	}
	return "更新成功", nil
}

// ValiDataAccountService
// 设置密码
func ValiDataAccountService(acc string) (bool, error) {

	if find := strings.Contains(acc, "@"); find {
		email, err := deals.ValiDataEmail(acc)
		if err != nil {
			return false, err
		}
		if email != nil && email.Id > 0 {
			return true, err
		}
	} else {
		phone, err := deals.ValiDataPhone(acc)
		if err != nil {
			return false, err
		}
		if phone != nil && phone.Id > 0 {
			return true, err
		}
	}
	return false, nil
}

// CheckNewAccountService
// 检查是否为新账号
func CheckNewAccountService(acc string) (bool, error) {
	var (
		err error
	)
	user := new(modelUser.Entity)
	if find := strings.Contains(acc, "@"); find {
		user, err = deals.ValiDataEmail(acc)
		if err != nil {
			return false, err
		}
	} else {
		user, err = deals.ValiDataPhone(acc)
		if err != nil {
			return false, err
		}
	}
	if user == nil || user.Id == 0 {
		return false, nil
	}
	if user.Salt == "noSalt" {
		return true, nil
	} else {
		return false, nil
	}

}

// GetAccountPersonaInfo
// 获取账号信息
func GetAccountPersonaInfo(id int64) (*domain.UserPersonal, error) {
	return adminDeal.GetUserInfoByUserId(id)
}

// AsUpdateUserByUid
// 异步更新
func AsUpdateUserByUid(id int64, mp map[string]interface{}) {
	xkutils.AsyncBackCall(func() {
		adminDeal.UpdateUserByUid(id, mp)
	})
}

func GetSaltStrService(u *domain.AccountInfo) (string, error) {
	var (
		phone, email string
		info         = new(modelUser.Entity)
		err          error
	)
	if strings.Contains(u.Account, "@") {
		email = u.Account
		info, err = deals.ValiDataEmail(u.Account)
	} else {
		phone = u.Account
		info, err = deals.ValiDataPhone(u.Account)
	}
	if err != nil {
		return "", err
	}
	if info == nil || info.Id == 0 {
		return "", errors.New(global.MsgWarnAccountErr)
	}
	andEmail, err := adminDeal.GetSaltByPhoneAndEmail(phone, email)
	if err != nil {
		return "", err
	}
	if andEmail != nil {
		return andEmail.Salt, err
	}
	return "", err

}

func CheckAccountErr(user *modelUser.Entity) error {
	if user.State == 1 {
		return errors.New(global.MsgWarnAccountState)
	}
	if user.PwdErr >= 5 {
		return errors.New(global.MsgWarnPwdCodeErr)
	}
	if user.EmailCodeErr >= 5 {
		return errors.New(global.MsgWarnEmailCodeErr)
	}
	if user.PhoneCodeErr >= 5 {
		return errors.New(global.MsgWarnPhoneCodeErr)
	}
	return nil
}
