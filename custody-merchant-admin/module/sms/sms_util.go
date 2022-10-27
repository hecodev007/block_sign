package sms

import (
	"bytes"
	"crypto/md5"
	"custody-merchant-admin/module/log"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type SmsInfo struct {
	AppKey  string
	Secret  string
	AppCode string
	url     string
}

// NewSms
// 传入配置
func NewSms(appKey, secret, code, url string) *SmsInfo {
	return &SmsInfo{
		AppKey:  appKey,
		Secret:  secret,
		AppCode: code,
		url:     url,
	}
}

// SendSms
// 发送短信
func (sms *SmsInfo) SendSms(phone, sendMsg string) (bool, error) {
	v := make(map[string]interface{})
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	w := md5.New()
	str := sms.AppKey + sms.Secret + timestamp
	_, err := io.WriteString(w, str)
	if err != nil {
		return false, err
	}
	md5str := fmt.Sprintf("%x", w.Sum(nil))

	v["sign"] = md5str
	v["timestamp"] = timestamp
	v["phone"] = phone
	v["extend"] = ""
	v["appcode"] = sms.AppCode
	v["appkey"] = sms.AppKey
	v["msg"] = sendMsg

	bytesData, err := json.Marshal(v)
	if err != nil {
		fmt.Println(err.Error())
		return false, err
	}
	//利用指定的method,url以及可选的body返回一个新的请求.如果body参数实现了io.Closer接口，Request返回值的Body 字段会被设置为body，并会被Client类型的Do、Post和PostFOrm方法以及Transport.RoundTrip方法关闭。
	body := bytes.NewReader(bytesData)
	//把form数据编下码
	//客户端,被Get,Head以及Post使用
	client := &http.Client{}
	reqest, err := http.NewRequest("POST", sms.url, body)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
	}
	//必须设定该参数,POST参数才能正常提交
	reqest.Header.Set("Content-Type", "application/json;charset=utf-8")
	resp, err := client.Do(reqest) //发送请求
	if err != nil {
		return false, err
	}
	if resp != nil {
		defer resp.Body.Close() //一定要关闭resp.Body
	}
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		if err != nil {
			return false, err
		}
	}
	return true, nil
}
