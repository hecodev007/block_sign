package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
)

func Request(method string, params []interface{}) ([]byte, error) {
	if params == nil {
		params = []interface{}{}
	}
	req := httplib.Post(beego.AppConfig.String("url"))
	if beego.AppConfig.String("usr") != "" && beego.AppConfig.String("pass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("usr"), beego.AppConfig.String("pass"))
	}

	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	req.JSONBody(reqbody)
	return req.Bytes()
}

func RequestStr(method string, params []interface{}) (string, error) {
	if params == nil {
		params = []interface{}{}
	}
	req := httplib.Post(beego.AppConfig.String("url"))
	if beego.AppConfig.String("usr") != "" && beego.AppConfig.String("pass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("usr"), beego.AppConfig.String("pass"))
	}
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	fmt.Println(params)
	req.JSONBody(reqbody)
	return req.String()
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
	//log.Infof("发送内容：%s", jsonStr)
	resp, err := client.Post(url, "application/json;charset=UTF-8", bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, _ := ioutil.ReadAll(resp.Body)
	return result, nil
}

func GetJson(url string, data map[string]interface{}) ([]byte, error) {

	if data != nil {
		param := ""
		for k, v := range data {
			param += fmt.Sprintf("%s&%v", k, v)
		}
		url = fmt.Sprintf("%s?%s", url, param)
	}
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, _ := ioutil.ReadAll(resp.Body)
	return result, nil
}
