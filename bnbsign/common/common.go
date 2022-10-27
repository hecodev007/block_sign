package common

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"strconv"
	"time"
)

// string to int64
func StrToInt64(str string) int64 {
	val, _ := strconv.ParseInt(str, 10, 64)
	return val
}

// string to int
func StrToInt(str string) int {
	val, _ := strconv.Atoi(str)
	return val
}

// string to float
func StrToFloat32(str string) float32 {
	tmp, _ := strconv.ParseFloat(str, 32)
	return float32(tmp)
}

// string to float64
func StrToFloat64(str string) float64 {
	tmp, _ := strconv.ParseFloat(str, 64)
	return tmp
}

// float64 to string
func Float64ToString(val float64) string {
	return strconv.FormatFloat(val, 'E', -1, 64)
}

// int to string
func IntToString(val int) string {
	return strconv.Itoa(val)
}

// int64 to string
func Int64ToString(val int64) string {
	return strconv.FormatInt(val, 64)
}

// int64 to string
func UInt64ToString(val uint64) string {
	return strconv.FormatUint(val, 64)
}

// utc时间转换成zh字符串时间
func TimeToStr(val int64) string {
	tm := time.Unix(val, 0)
	return tm.Format("2006-01-02 15:04:05")
}

// utc时间转换成zh字符串时间
func StrToTime(val string) int64 {
	p, _ := time.Parse("2006-01-02 15:04:05", val)
	return p.Unix()
}

// 获取毫秒UTC时间
func GetMillTime() int64 {
	timestamp := time.Now().UnixNano() / 1000000
	return timestamp
}

//加密字符串
func AesEncrypt(strMesg string, key []byte) (string, error) {
	var iv = []byte(key)[:aes.BlockSize]
	encrypted := make([]byte, len(strMesg))
	aesBlockEncrypter, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesEncrypter := cipher.NewCFBEncrypter(aesBlockEncrypter, iv)
	aesEncrypter.XORKeyStream(encrypted, []byte(strMesg))

	str := base64.StdEncoding.EncodeToString(encrypted)
	return str, nil
}

//解密字符串
func AesDecrypt(srcStr string, key []byte) (strDesc string, err error) {
	defer func() {
		//错误处理
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	// 解码字符串
	src, err := base64.StdEncoding.DecodeString(srcStr)

	var iv = []byte(key)[:aes.BlockSize]
	decrypted := make([]byte, len(src))
	var aesBlockDecrypter cipher.Block
	aesBlockDecrypter, err = aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	aesDecrypter := cipher.NewCFBDecrypter(aesBlockDecrypter, iv)
	aesDecrypter.XORKeyStream(decrypted, src)

	return string(decrypted), err
}
