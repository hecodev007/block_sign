package common

import (
	"encoding/base64"
	"github.com/astaxie/beego/logs"
	"strings"
)

func EncodeBasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func DecodeBasicAuth(authInfo string) string {
	authInfoSplit := strings.Split(authInfo, " ")
	authByte, err := base64.StdEncoding.DecodeString(authInfoSplit[1])
	if err != nil {
		logs.Debug("解析auth info失败")
		return ""
	}
	return string(authByte)
}
