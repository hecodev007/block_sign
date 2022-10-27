package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/blockchains-go/log"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 发送GET请求
// url：         请求地址
// response：    请求返回的内容
func Get(url string) ([]byte, error) {

	// 超时时间：60秒
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
	}

	return result.Bytes(), nil
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
	log.Infof("发送URL： %s", url)
	log.Infof("发送内容： %s", jsonStr)
	resp, err := client.Post(url, "application/json;charset=UTF-8", bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("resp status code is not equal 200 ,Code=[%d]", resp.StatusCode)
	}
	result, _ := ioutil.ReadAll(resp.Body)
	return result, nil
}

// 发送POST请求
// url：         请求地址
// data：        POST请求提交的数据
// contentType： 请求体格式，如：application/json
// content：     请求放回的内容
func PostJsonData(url string, data []byte) ([]byte, error) {
	// 超时时间：30秒
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Post(url, "application/json;charset=UTF-8", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	result, _ := ioutil.ReadAll(resp.Body)
	return result, nil
}

// 发送GET请求
// url：         请求地址
// data：        GET请求提交的数据
// contentType： 请求体格式，如：application/json
// content：     请求放回的内容
func GetByAuth(url, user, pwd string) ([]byte, error) {

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	if user != "" && pwd != "" {
		req.SetBasicAuth(user, pwd)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// PostJsonWithRetry 支持重试的POST请求
// 重试次数+1 = 实际请求的次数
// 每次重试间隔+2秒（2、4、6、8、10...）
func PostJsonWithRetry(url, user, pwd string, data interface{}, retryCount uint32) ([]byte, error) {
	var (
		response []byte
		err      error
	)
	for i := 0; i < int(retryCount)+1; i++ {
		if i > 0 {
			sleepDuration := time.Duration((i + 1) * 2)
			log.Infof("%s  retry in %d seconds", url, i+1, sleepDuration)
			time.Sleep(sleepDuration * time.Second)
		}

		response, err = PostJsonByAuth(url, user, pwd, data)
		if err == nil {
			return response, nil
		}
		if !strings.HasPrefix(err.Error(), "http resp status error") {
			// 非HTTP错误，直接返回，不再重试
			return response, nil
		}

		log.Infof("POST 请求发生错误 %v", err)
	}
	return response, err
}

// 发送POST请求
// url：         请求地址
// data：        POST请求提交的数据
// contentType： 请求体格式，如：application/json
// content：     请求放回的内容
func PostJsonByAuth(url, user, pwd string, data interface{}) ([]byte, error) {
	client := &http.Client{Timeout: 360 * time.Second}
	jsonStr, _ := json.Marshal(data)
	log.Infof("发送url：%s", url)
	log.Infof("发送内容：%s", string(jsonStr))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	if user != "" && pwd != "" {
		req.SetBasicAuth(user, pwd)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http resp status error,Status=%d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// 发送POST请求
// url：         请求地址
// data：        POST请求提交的数据
// contentType： 请求体格式，如：application/json
// content：     请求放回的内容
func PostJsonByAuthAndTime(url, user, pwd string, data interface{}, timeOutSec int) ([]byte, error) {
	client := &http.Client{Timeout: time.Duration(timeOutSec) * time.Second}
	jsonStr, _ := json.Marshal(data)
	// log.Infof("发送内容：%s", string(jsonStr))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	if user != "" && pwd != "" {
		req.SetBasicAuth(user, pwd)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http resp status error,Status=%d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Infof("PostJsonByAuthAndTime response %s", string(body))
	return body, nil
}

// 发送POST请求
// url：         请求地址
// data：        POST请求提交的数据
// contentType： 请求体格式，如：application/json
// content：     请求放回的内容
func PostJsonByAuthByClient(url, user, pwd string, data interface{}, client *http.Client) ([]byte, error) {
	jsonStr, _ := json.Marshal(data)
	// log.Infof("发送内容：%s", string(jsonStr))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	if user != "" && pwd != "" {
		req.SetBasicAuth(user, pwd)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http resp status error,Status=%d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// 给商户回调，自封装签名
// data 发送内容，
func PostMapForCallBack(urlStr string, mapParams map[string]interface{}, apiKey, apiSecret string) ([]byte, error) {
	log.Infof("回调1", urlStr)
	log.Infof("回调的url: ", urlStr)
	mapData, err := GetSignParamsForCallBack(apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("PostForCallBack error :%s", err.Error())
	}
	log.Infof("回调2")
	for k, v := range mapData {
		mapParams[k] = v

	}
	log.Infof("回调3")
	urldata := make(url.Values)
	for k, v := range mapParams {
		urldata.Set(k, fmt.Sprintf("%v", v))
		if k == "is_in" {
			isInType := fmt.Sprintf("%v", v)
			if isInType == "2" {
				urldata.Set("msg", "success")
			}
		}
	}
	log.Infof("回调4")
	log.Infof("回调url：%s", urlStr)
	log.Infof("回调内容：%s", urldata.Encode())

	unescape, err := url.QueryUnescape(urldata.Encode())
	if err != nil {
		return nil, err
	}
	log.Infof("unescape 回调内容：%s", unescape)
	var resp *http.Response
	if urldata.Get("coin_name") == "xdag" {
		log.Info("xdag推送")
		resp, err = http.Post(urlStr, "application/x-www-form-urlencoded", strings.NewReader(unescape))
	} else {
		// resp, err := client.Post(url, "application/x-www-form-urlencoded", bytes.NewBuffer(postData))
		resp, err = http.PostForm(urlStr, urldata)
	}
	if err != nil {
		log.Infof("回调 http.Post err: %s", err.Error())
		return nil, err
	}
	log.Infof("回调5")
	defer resp.Body.Close()
	result, _ := ioutil.ReadAll(resp.Body)
	log.Infof("调用交易所返回： %s", result)
	return result, nil
}

// 给商户回调，自封装签名
// data 发送内容，
func PostByteForCallBack(urlStr string, dataPamrams []byte, apiKey, apiSecret string) ([]byte, error) {
	// 先转换为map
	mapParams := make(map[string]interface{})
	d := json.NewDecoder(bytes.NewReader(dataPamrams))
	d.UseNumber()
	d.Decode(&mapParams)
	mapData, err := GetSignParamsForCallBack(apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("PostForCallBack error :%s", err.Error())
	}
	for k, v := range mapData {
		mapParams[k] = v
	}
	urldata := make(url.Values)
	for k, v := range mapParams {
		urldata.Set(k, fmt.Sprintf("%v", v))
	}
	log.Infof("出txId回调url：%s", urlStr)
	log.Infof("出txId回调内容：%v", urldata)
	// resp, err := client.Post(url, "application/x-www-form-urlencoded", bytes.NewBuffer(postData))
	resp, err := http.PostForm(urlStr, urldata)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	result, _ := ioutil.ReadAll(resp.Body)
	log.Infof("出txId回调返回来的内容：%s", string(result))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("StatusCode ！= 200,resp:%s", string(result))
	}
	return result, nil
}
