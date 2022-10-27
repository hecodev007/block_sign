package global

import (
	"github.com/jinzhu/gorm"
)

var DBEngine *gorm.DB

const (
	HOUR               = "HOUR"
	DAY                = "DAY"
	WEEK               = "WEEK"
	MONTH              = "MONTH"
	YEAR               = "YEAR"
	YyyyMmDd           = "2006-01-02"
	YyyyMmDdHhMmSs     = "2006-01-02 15:04:05"
	YyyyMmDdHhMmSsNnnn = "2006-01-02 15:04:05.9999"
	PhoneCodeErr       = "PHONE_CODE_ERR"
	EmailCodeErr       = "EMAIL_CODE_ERR"
	PwdCodeErr         = "PWD_CODE_ERR"
	XCaNonce           = "X-Ca-Nonce"
	XCaTime            = "X-Ca-Time"
	XCaSignStr         = "X-Ca-SignStr"
)
