package common

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"io/ioutil"
	"net/http"
	"time"
)

func Request(method string, params []interface{}) ([]byte, error) {
	if params == nil {
		params = []interface{}{}
	}
	req := httplib.Post(beego.AppConfig.String("nodeurl")).SetTimeout(time.Second*10, time.Second*10)
	if beego.AppConfig.String("rpcuser") != "" && beego.AppConfig.String("rpcpass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("rpcuser"), beego.AppConfig.String("rpcpass"))
	}
	if beego.AppConfig.DefaultBool("enabletls", false) {
		req.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
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
	req := httplib.Post(beego.AppConfig.String("nodeurl"))
	if beego.AppConfig.String("rpcuser") != "" && beego.AppConfig.String("rpcpass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("rpcuser"), beego.AppConfig.String("rpcpass"))
	}
	if beego.AppConfig.DefaultBool("enabletls", false) {
		req.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	req.JSONBody(reqbody)
	return req.String()
}

func RequestObject(method string, params []interface{}, v interface{}) error {
	if params == nil {
		params = []interface{}{}
	}
	req := httplib.Post(beego.AppConfig.String("nodeurl"))
	if beego.AppConfig.String("rpcuser") != "" && beego.AppConfig.String("rpcpass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("rpcuser"), beego.AppConfig.String("rpcpass"))
	}
	if beego.AppConfig.DefaultBool("enabletls", false) {
		req.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	req.JSONBody(reqbody)
	bytes, err := req.Bytes()
	if bytes == nil || err != nil {
		return err
	}
	err = json.Unmarshal(bytes, v)
	return err
}

func RequestUrl(method string, params []interface{}, url string) ([]byte, error) {
	if params == nil {
		params = []interface{}{}
	}
	req := httplib.Post(url)
	if beego.AppConfig.String("rpcuser") != "" && beego.AppConfig.String("rpcpass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("rpcuser"), beego.AppConfig.String("rpcpass"))
	}
	if beego.AppConfig.DefaultBool("enabletls", false) {
		req.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
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

func Post(method string, params []interface{}) ([]byte, error) {
	if params == nil {
		params = []interface{}{}
	}
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	reqdata, err := json.Marshal(reqbody)
	if reqdata == nil || err != nil {
		return nil, err
	}
	c := &http.Client{
		Timeout: 15 * time.Second,
	}
	req, _ := http.NewRequest("POST", beego.AppConfig.String("nodeurl"), bytes.NewBuffer(reqdata))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "Keep-Alive")
	if beego.AppConfig.String("rpcuser") != "" && beego.AppConfig.String("rpcpass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("rpcuser"), beego.AppConfig.String("rpcpass"))
	}

	var res *http.Response
	res, err = c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return content, nil
}
