package global

import (
	"fmt"
	"strings"
)

const (
	CoinsList              = "Admin:COINS:LIST"
	PhoneCode              = "Admin:PHONE:CODE"
	ChainCoin              = "Admin:CHAIN:COIN"
	CdnNonce               = "Admin:CDN:NONCE:%s"
	CheckAccount           = "Admin:USER:CHECK:ACCOUNT:FIRST:%d"
	MenuTree               = "USER:MENU:TREE:%d"
	MenuBtn                = "USER:MENU:BTN:%d"
	AdminMenuTree          = "Admin:MENU:TREE:%d"
	AdminMenuBtn           = "Admin:MENU:btn:%d"
	MenuTreeRole           = "Admin:USER:MENU:TREE:ROLE:%d"
	PhoneSmsLogin          = "Admin:PHONE:SMS:LOGIN:%s"
	SendLoginSec           = "Admin:Send:LOGIN:Sec:%s"
	ResetPwdSendSec        = "Admin:Send:ResetPwd:Sec:%s"
	EmailLogin             = "Admin:EMAIL:LOGIN:%s"
	PhoneSmsResetPwd       = "Admin:PHONE:SMS:RESET:PWD:%s"
	PhoneLimitCode         = "Admin:PHONE:Limit:CODE:%s"
	EmailResetPwd          = "Admin:EMAIL:RESET:PWD:%s"
	EmailLimitCode         = "Admin:EMAIL:Limit:CODE:%s"
	UserTokenVerify        = "Admin:USER:TOKEN:VERIFY:%d"
	ServiceCoin            = "SERVICE:COIN:%d:%d"
	TestBillListSuccess    = "Admin:Test:Bill:List:Success"
	TestBillListErr        = "Admin:Test:Bill:List:Err"
	CustodyPriceHoo        = "Custody:Price:Hoo"
	CustodyHooGeekPriceHoo = "Custody:Price:HooGeek"
	CustodyHooFee          = "Custody:Price:Fee"
	AESRandomStr           = "Admin:AES:Random:%d"
)

const (
	AllSysRouter = "USER:SYSROUTER"
)

func GetCacheKey(format string, v ...interface{}) string {
	key := strings.Replace(fmt.Sprintf(format, v), "[", "", 1)
	key = strings.Replace(key, "]", "", 1)
	return key
}
