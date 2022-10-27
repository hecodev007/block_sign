package hc

import (
	"encoding/json"
	"github.com/group-coldwallet/chaincore2/common"
	"github.com/group-coldwallet/common/log"
	//"encoding/hex"

	"github.com/astaxie/beego"
)

type RpcController struct {
	beego.Controller
}

func (c *RpcController) Post() {
	// 返回数据
	resp := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"result":  "",
	}

	for true {
		var jsonObj map[string]interface{}
		json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)
		log.Debug(jsonObj)

		params := jsonObj["params"].([]interface{})
		respdata, err := common.Request(jsonObj["method"].(string), params)
		if respdata == nil || err != nil {
			beego.Error(err)
			break
		}

		log.Debug(string(respdata))
		var rawdata map[string]interface{}
		err = json.Unmarshal(respdata, &rawdata)
		if err != nil {
			beego.Error(err)
			break
		}

		c.Data["json"] = rawdata
		c.ServeJSON()
		return
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
