package ycash

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
)

// 获取节点区块高度
func getblock_count() (string, error) {
	// 操作neo节点
	req := httplib.Post(beego.AppConfig.String("nodeurl")) //.SetTimeout(time.Millisecond*60, time.Millisecond*120)
	if beego.AppConfig.String("rpcuser") != "" && beego.AppConfig.String("rpcpass") != "" {
		req.SetBasicAuth(beego.AppConfig.String("rpcuser"), beego.AppConfig.String("rpcpass"))
	}
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  "getblockcount",
		"params":  []interface{}{},
	}
	req.JSONBody(reqbody)
	respdata, err := req.String()
	return respdata, err
}

// 获取区块数据
func getblock_data(val interface{}) (string, error) {
	// 操作neo节点
	req := httplib.Post(beego.AppConfig.String("nodeurl")) //.SetTimeout(time.Millisecond*60, time.Millisecond*120)
	reqbody := map[string]interface{}{
		"id":      "1",
		"jsonrpc": "2.0",
		"method":  "getblock",
		"params":  []interface{}{val, 1},
	}
	req.JSONBody(reqbody)
	respdata, err := req.String()
	return respdata, err
}
