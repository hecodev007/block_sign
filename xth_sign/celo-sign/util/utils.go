package util

import (
	"bytes"
	"encoding/json"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 发送POST请求
// url：         请求地址
// data：        POST请求提交的数据
// contentType： 请求体格式，如：application/json
// content：     请求放回的内容
func PostJson(url string, data interface{}) ([]byte, error) {
	// 超时时间：30秒
	client := &http.Client{Timeout: 60 * time.Second}
	jsonStr, _ := json.Marshal(data)
	//log.Infof("发送内容：%s", jsonStr)
	resp, err := client.Post(url, "application/json;charset=UTF-8", bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

func HexToDec(hex string) int64 {
	if strings.HasPrefix(hex, "0x") {
		hex = hex[2:]
	}
	//tmp:=decimal.NewFromInt(0)
	//fmt.Println(len(hex))
	//for i,h:=range hex{
	//	pow:=int64(len(hex)-1-i)
	//	fmt.Println(pow)
	//	//16 **pow
	//	powDec:=hexDec.Pow(decimal.NewFromInt(pow))
	//
	//	d:=hexDecMap[string(h)]
	//
	//	tmp = tmp.Add(d.Mul(powDec))
	//}
	//return tmp.IntPart()
	n, err := strconv.ParseInt(hex, 16, 64)
	if err != nil {
		return -1
	}
	return n
}

func DecToHex(dec int64) string {
	return "0x" + strconv.FormatInt(dec, 16)
}
func Timestamp(seconds int64) string {
	var timelayout = "2006-01-02 T 15:04:05.000" //时间格式

	datatimestr := time.Unix(seconds, 0).Format(timelayout)

	return datatimestr

}
