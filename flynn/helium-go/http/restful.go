package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"net/http"
	"strings"
)

/*
封装Get/Post方法
*/

func _httpRequest(reqType string, reqUrl string, postData string, requestHeaders map[string]string) ([]byte, error) {
	req, _ := http.NewRequest(reqType, reqUrl, strings.NewReader(postData))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/31.0.1650.63 Safari/537.36")
	if requestHeaders != nil {
		for k, v := range requestHeaders {
			req.Header.Add(k, v)
		}
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("HttpStatusCode:%d ,Desc:%s", resp.StatusCode, resp.Status))
	}
	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bodyData, nil
}

func HttpGet(reqUrl string, reply interface{}) error {
	respData, err := _httpRequest(http.MethodGet, reqUrl, "", nil)
	if err != nil {
		return err
	}
	//fmt.Println(string(respData))
	err = json.Unmarshal(respData, &reply)
	if err != nil {
		return err
	}
	return nil
}

func HttpPost(reqUrl, postData string, reply interface{}) error {
	headers := map[string]string{
		"Content-Type": "application/json;charset=UTF-8"}
	respData, err := _httpRequest(http.MethodPost, reqUrl, postData, headers)
	if err != nil {
		return err
	}
	err = json.Unmarshal(respData, &reply)
	if err != nil {
		return err
	}
	return nil
}

//func HttpPostForm(client *http.Client,reqUrl string,postData Value)([]byte,error){
//	headers := map[string]string{
//		"Content-Type": "application/json;charset=UTF-8"}
//	params,err:=json.Marshal(postData)
//	if err != nil {
//		return nil,err
//	}
//	log.Printf("reqUrl:%s",reqUrl)
//	return _httpRequest(client,http.MethodPost,reqUrl,string(params),headers)
//}
