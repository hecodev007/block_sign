package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bsc-sign/util/log"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func Post(apiUrl, path string, data url.Values) ([]byte, error) {
	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = path
	u.RawQuery = data.Encode()
	urlStr := fmt.Sprintf("%v", u)

	client := &http.Client{}
	r, _ := http.NewRequest("POST", urlStr, bytes.NewBufferString(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, _ := client.Do(r)

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// 发送POST请求
// url：         请求地址
// data：        POST请求提交的数据
// contentType： 请求体格式，如：application/json
// content：     请求放回的内容
func PostJson(url string, data interface{}) ([]byte, error) {
	// 超时时间：30秒
	client := &http.Client{Timeout: 60 * time.Second}
	jsonStr, _ := json.Marshal(data)
	log.Infof("PostJson发送内容：%s", jsonStr)
	resp, err := client.Post(url, "application/json;charset=UTF-8", bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	result, _ := ioutil.ReadAll(resp.Body)
	return result, nil
}

func RemoveHex0x(hexStr string) string {
	if strings.HasPrefix(hexStr, "0x") {
		return hexStr[2:]
	}
	return hexStr
}

var hexDecMap = map[string]decimal.Decimal{
	"0": decimal.NewFromInt(0),
	"1": decimal.NewFromInt(1),
	"2": decimal.NewFromInt(2),
	"3": decimal.NewFromInt(3),
	"4": decimal.NewFromInt(4),
	"5": decimal.NewFromInt(5),
	"6": decimal.NewFromInt(6),
	"7": decimal.NewFromInt(7),
	"8": decimal.NewFromInt(8),
	"9": decimal.NewFromInt(9),
	"a": decimal.NewFromInt(10),
	"b": decimal.NewFromInt(11),
	"c": decimal.NewFromInt(12),
	"d": decimal.NewFromInt(13),
	"e": decimal.NewFromInt(14),
	"f": decimal.NewFromInt(15),
}

var hexDec = decimal.NewFromInt(16)

func HexToDec(hex string) *big.Int {
	if strings.HasPrefix(hex, "0x") {
		hex = hex[2:]
	}

	bigIntValue, ok := new(big.Int).SetString(hex, 16)
	if !ok {
		return big.NewInt(-1)
	}
	return bigIntValue
}

func DecToHex(dec int64) string {
	return "0x" + strconv.FormatInt(dec, 16)
}
func Timestamp(seconds int64) string {
	var timelayout = "2006-01-02 T 15:04:05.000" // 时间格式

	datatimestr := time.Unix(seconds, 0).Format(timelayout)

	return datatimestr

}

func Del0xToLower(address string) string {
	if strings.HasPrefix(address, "0x") {
		return strings.ToLower(strings.TrimPrefix(address, "0x"))
	}
	return strings.ToLower(address)
}

// ParseBigInt parse hex string value to big.Int
func ParseBigInt(value string) (*big.Int, error) {
	i := &big.Int{}
	_, err := fmt.Sscan(value, i)

	return i, err
}
