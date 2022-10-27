package common

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
)

func Request(method string, params []interface{}) ([]byte, error) {
	if params == nil {
		params = []interface{}{}
	}
	req := httplib.Post(beego.AppConfig.String("url"))
	req.SetBasicAuth(beego.AppConfig.String("usr"), beego.AppConfig.String("pass"))
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
	req.SetBasicAuth(beego.AppConfig.String("usr"), beego.AppConfig.String("pass"))
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	req.JSONBody(reqbody)
	return req.String()
}
