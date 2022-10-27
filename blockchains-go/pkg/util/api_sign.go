package util

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/log"
	"math/big"
	"net/url"
	"sort"
	"strings"
	"time"
)

//1.所有接口都需要进行鉴权，参数为clientId, ts, nonce, sign。clientId是api key, client_key为密钥，请妥善保管。

//2.client_id为api key，ts为当前时间戳，与服务器时间差正负5秒会被拒绝，nonce为随机字符串，不能与上次请求所使用相同。

//3.签名方法, 将client_id, ts, nonce进行排序连接，使用hmac-sha256方法进行签名，例如待签名字符串为: client_id=abc&nonce=xyz&ts=1571293029

//4.签名: sign = hmac.New(client_key, sign_str, sha256)

//5.Content-Type: application/x-www-form-urlencoded

//6.post接口，请将参数放在请求体里面。

//参数为client_id, ts, nonce, sign
//只验证了那几个参数，其他多余参数没有加入验签，需要注意
type ApiSignParams struct {
	ClientId string `json:"client_id" form:"client_id"` //api key
	Ts       int64  `json:"ts" form:"ts"`               //ts为当前时间戳 与服务器时间差正负5秒会被拒绝
	Nonce    string `json:"nonce" form:"nonce"`         //nonce为随机字符串 不能与上次请求所使用相同
	Sign     string `json:"sign" form:"sign"`           //sign = GetSign hmac.New(client_key, sign_str, sha256)
}

type CustodyApiSignParams map[string]interface{}


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

	str := EncodeQueryString(params)
	sign = ComputeHmac256(str, s.ApiSecret)
	return
}


//获取签名sign
func (s *CustodyApiSignParams) GetSign() (sign string, err error) {
	param := *s
	apiKey := param["client_id"].(string)
	apiSecret := param["api_secret"].(string)
	nonce := interface2String(param["nonce"])
	//ts := fmt.Sprintf("%v", param["api_key"])


	if apiKey == "" || apiSecret == "" || nonce == "" {
		return "", errors.New("params error")
	}
	nonceStr := SignNonceMap[apiKey]
	if nonce == nonceStr {
		return "", errors.New("same nonce as last time")
	} else {
		SignNonceMap[apiKey] = nonce
	}


	str := EncodeQueryInterface(*s)

	log.Infof("enstr1 := %v\n", str)
	log.Infof("enstr2 := %v\n", apiSecret)
	sign = ComputeHmac256(str, apiSecret)
	log.Infof("enstr3 sign:= %v\n", sign)
	return
}


//获取签名sign
func GetSignParamsForCallBack(apikey, apiSecret string) (mapData map[string]interface{}, err error) {
	//ts = strconv.FormatInt(time.Now().Unix(), 10)
	//nonce = createRandomString(6)
	if apikey == "" || apiSecret == "" {
		return nil, errors.New("params error")
	}
	nowTime := time.Now().Unix()
	nonce := createRandomString(6)
	params := make(map[string]string)
	params["client_id"] = apikey
	params["ts"] = fmt.Sprintf("%v", nowTime)
	params["nonce"] = nonce
	str := EncodeQueryString(params)
	sign := ComputeHmac256(str, apiSecret)

	mapData = make(map[string]interface{}, 0)
	mapData["client_id"] = apikey
	mapData["apits"] = fmt.Sprintf("%v", nowTime)
	mapData["apinonce"] = nonce
	mapData["apisign"] = sign
	return
}

//产生随机字符串，主要是测试使用
func createRandomString(len int) string {
	var container string
	var str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := bytes.NewBufferString(str)
	length := b.Len()
	bigInt := big.NewInt(int64(length))
	for i := 0; i < len; i++ {
		randomInt, _ := rand.Int(rand.Reader, bigInt)
		container += string(str[randomInt.Int64()])
	}
	return container
}

//签名
func ComputeHmac256(data string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	_, err := h.Write([]byte(data))
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

/// 拼接query字符串
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
		lines[i] = url.QueryEscape(k) + "=" + url.QueryEscape(query[k])
	}
	return strings.Join(lines, "&")
}

// 拼接除sign以外的所有query字符串
func EncodeQueryInterface(query map[string]interface{}) string {
	keys := make([]string, 0)
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var lines = make([]string, 0)
	for _,item := range keys {
		if item == "sign" || item == "api_secret" {
			continue
		}
		s := interface2String(query[item])
		lines = append(lines ,url.QueryEscape(item) + "=" + s)
	}
	return strings.Join(lines, "&")
}

func interface2String(inter interface{}) string{
	switch inter.(type) {
	case string:
		s := inter.(string)
		return s
	case int:
		i :=  inter.(int)
		s := fmt.Sprintf("%v",i)
		return s
	case int64:
		i :=  inter.(int64)
		s := fmt.Sprintf("%v",i)
		return s
	case float64:
		i :=  inter.(float64)
		//s := strconv.FormatFloat(i, 'E', -1, 64)
		s := fmt.Sprintf("%.f", i)
		return s
	}
	return ""

}