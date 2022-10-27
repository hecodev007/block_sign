package crypt

import (
	"crypto/hmac"
	"crypto/sha256"
	"custody-merchant-admin/module/log"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

type ApiSign struct {
	ApiKey    string //
	ApiSecret string //
	Ts        string //ts为当前时间戳 与服务器时间差正负5秒会被拒绝
	Nonce     string //nonce为随机字符串 不能与上次请求所使用相同
}

//由于nonce为随机字符串 不能与上次请求所使用相同，使用map临时存储
//key为client_id value为nonce
var SignNonceMap map[string]string

func init() {
	SignNonceMap = make(map[string]string, 0)
}

//获取签名sign
func (s *ApiSign) GetSign() (sign string, err error) {
	//ts = strconv.FormatInt(time.Now().Unix(), 10)
	//nonce = createRandomString(6)
	if s.ApiKey == "" || s.ApiSecret == "" || s.Nonce == "" || s.Ts == "" {
		return "", errors.New("params error")
	}
	nonceStr := SignNonceMap[s.ApiKey]
	if s.Nonce == nonceStr {
		return "", errors.New("same nonce as last time")
	} else {
		SignNonceMap[s.ApiKey] = s.Nonce
	}
	params := make(map[string]string)
	params["client_id"] = s.ApiKey
	params["ts"] = s.Ts
	params["nonce"] = s.Nonce
	log.Infof("加密前数据结构 %+v", params)
	str := EncodeQueryString(params)
	log.Infof("加密拼接 %+v", str)
	sign = computeHmac256(str, s.ApiSecret)
	log.Infof("加密后sign %+v", sign)
	return
}

// EncodeQueryString
// 拼接query字符串
func EncodeQueryString(query map[string]string) string {
	keys := make([]string, 0)
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var len = len(keys)
	var lines = make([]string, len)
	for i := 0; i < len; i++ {
		var k = keys[i]
		s := interface2String(query[k])
		lines[i] = url.QueryEscape(k) + "=" + s
	}
	return strings.Join(lines, "&")
}

// computeHmac256
// 签名
func computeHmac256(data string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	_, err := h.Write([]byte(data))
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func interface2String(inter interface{}) string {
	switch inter.(type) {
	case string:
		s := inter.(string)
		return s
	case int:
		i := inter.(int)
		s := fmt.Sprintf("%v", i)
		return s
	case int64:
		i := inter.(int64)
		s := fmt.Sprintf("%v", i)
		return s
	case float64:
		i := inter.(float64)
		//s := strconv.FormatFloat(i, 'E', -1, 64)
		s := fmt.Sprintf("%.f", i)
		return s
	}
	return ""

}
