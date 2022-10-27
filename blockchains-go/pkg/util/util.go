package util

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type CollectingStatus string

const (
	signKey        = "sign"
	key            = "key"
	expireDuration = 120

	CollectingProcessing   CollectingStatus = "1"
	CollectingAmountEnough CollectingStatus = "2"
	CollectingFailure      CollectingStatus = "3"
	CollectingIgnore       CollectingStatus = "4"

	CollectingExpire = 5 * time.Minute
)

func IsInArrayStr(target string, str_array []string) bool {
	for _, element := range str_array {
		if target == element {
			return true

		}
	}
	return false
}

func NoticeSignV(data map[string]interface{}, mchKey string, mchId string) string {
	sign := SignHmac256V(data, mchKey)
	expire := time.Now().Add(time.Second * expireDuration).Unix()
	return mchId + ":" + strconv.FormatInt(expire, 10) + ":" + sign
}

func NoticeSign(data map[string]string, mchKey string, mchId string) string {
	sign := SignHmac256(data, mchKey)
	expire := time.Now().Add(time.Second * expireDuration).Unix()
	return mchId + ":" + strconv.FormatInt(expire, 10) + ":" + sign
}

//md5签名
func Sign(data map[string]string, mchKey string) string {
	var (
		vdata map[string]string
		keys  []string
		query []string
	)
	vdata = make(map[string]string)
	for k, v := range data {
		if k != signKey {
			vdata[k] = v
		}
	}
	for k, _ := range vdata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		query = append(query, k+"="+vdata[k])
	}
	return hex.EncodeToString(Md5Hash([]byte(hex.EncodeToString(Md5Hash([]byte(mchKey+strings.Join(query, "&")))) + mchKey)))
}

//hmac签名
func SignHmac256V(data map[string]interface{}, mchKey string) string {
	var (
		vdata map[string]string
		keys  []string
		query []string
	)
	vdata = make(map[string]string)
	for k, v := range data {
		if k != signKey && k != key {
			vdata[k] = fmt.Sprintf("%v", v)
		}
	}
	//这样分开处理是因为考虑有中间值插入
	//所以应分开处理
	for k, _ := range vdata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		query = append(query, k+"="+vdata[k])
	}
	fmt.Println(strings.Join(query, "&"))
	return base64.StdEncoding.EncodeToString([]byte(hex.EncodeToString(HmacSha256([]byte(strings.Join(query, "&")+"&key="+mchKey), []byte(mchKey)))))
}

//hmac256签名
func SignHmac256(data map[string]string, mchKey string) string {
	var (
		vdata map[string]string
		keys  []string
		query []string
	)
	vdata = make(map[string]string)
	for k, v := range data {
		if k != signKey && k != key {
			vdata[k] = v
		}
	}
	//这样分开处理是因为考虑有中间值插入
	//所以应分开处理
	for k, _ := range vdata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		query = append(query, k+"="+vdata[k])
	}
	return base64.StdEncoding.EncodeToString([]byte(hex.EncodeToString(HmacSha256([]byte(strings.Join(query, "&")+"&key="+mchKey), []byte(mchKey)))))
}

func HmacSha256(data []byte, key []byte) []byte {
	k := []byte(key)
	h := hmac.New(sha256.New, k)
	h.Write(data)
	return h.Sum(nil)
}

func Md5Hash(data []byte) []byte {
	signByte := []byte(data)
	hash := md5.New()
	hash.Write(signByte)
	return hash.Sum(nil)
}

//===新增interface签名，建议使用这个签名====
func SignInterface(data map[string]interface{}, mchKey string) string {
	var (
		vdata map[string]interface{}
		keys  []string
		query []string
	)
	vdata = make(map[string]interface{})
	for k, v := range data {
		if k != signKey {
			vdata[k] = v
		}
	}
	for k, _ := range vdata {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		val := fmt.Sprintf("%v", vdata[k])
		query = append(query, k+"="+val)
	}
	return hex.EncodeToString(Md5Hash([]byte(hex.EncodeToString(Md5Hash([]byte(mchKey+strings.Join(query, "&")))) + mchKey)))
}
