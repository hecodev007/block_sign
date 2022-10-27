package util

import (
	"encoding/base64"
	"github.com/bsc-sign/util/log"
	"strings"
)

func EncodeBasicAuth(user, password string) string {
	auth := user + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func DecodeBasicAuth(authInfo string) string {
	authInfoSplit := strings.Split(authInfo, " ")
	authByte, err := base64.StdEncoding.DecodeString(authInfoSplit[1])
	if err != nil {
		log.Info("解析auth info失败")
		return ""
	}
	return string(authByte)
}
