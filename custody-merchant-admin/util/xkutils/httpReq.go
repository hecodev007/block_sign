package xkutils

import (
	"bytes"
	"custody-merchant-admin/module/log"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// Get
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

func Send() {

}

// PostJson
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

// PostJsonData
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

// GetByAuth
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

// PostJsonByAuth
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

// PostJsonByAuthAndTime
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

// PostJsonByAuthByClient
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

func PostForm(uri string, form url.Values) ([]byte, error) {
	log.Infof("PostForm  uri := %+v\n", uri)
	log.Infof("PostForm  form := %+v\n", form)
	resp, err := http.PostForm(uri, form)

	if err != nil {
		// handle error
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		return nil, err
	}
	return body, err
}
