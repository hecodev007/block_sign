package neo

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"time"
)

func Request(method string, params []interface{}) ([]byte, error) {
	if params == nil {
		params = []interface{}{}
	}
	req := httplib.Post(beego.AppConfig.String("nodeurl")).SetTimeout(time.Second*3, time.Second*10)
	if beego.AppConfig.String("rpcuser") != "" && beego.AppConfig.String("rpcpass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("rpcuser"), beego.AppConfig.String("rpcpass"))
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
	req := httplib.Post(beego.AppConfig.String("nodeurl")).SetTimeout(time.Second*3, time.Second*10)
	if beego.AppConfig.String("rpcuser") != "" && beego.AppConfig.String("rpcpass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("rpcuser"), beego.AppConfig.String("rpcpass"))
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
	req := httplib.Post(beego.AppConfig.String("nodeurl")).SetTimeout(time.Second*3, time.Second*10)
	if beego.AppConfig.String("rpcuser") != "" && beego.AppConfig.String("rpcpass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("rpcuser"), beego.AppConfig.String("rpcpass"))
	}
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	req.JSONBody(reqbody)

	//bb,_ := json.Marshal(reqbody)
	//log.Infof("reqbody:%s",string(bb))
	bytes, err := req.Bytes()
	//log.Infof("reqbody bytes:%s",string(bytes))
	if bytes == nil || err != nil {
		return err
	}
	err = json.Unmarshal(bytes, v)
	return err
}
