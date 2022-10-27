package agent

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/group-coldwallet/common/log"
	"strings"
	"time"
)

type RpcController struct {
	beego.Controller
}

var SupportMethod map[string]map[string]bool = make(map[string]map[string]bool)

func (c *RpcController) Post() {
	// 返回数据
	resp := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"result":  "",
		"error":   nil,
	}

	set_resp := func(result interface{}, err interface{}) {
		resp["result"] = result
		if err != nil {
			resp["error"] = map[string]interface{}{"code": -1, "message": err.(string)}
		}
	}

	for true {
		var jsonObj map[string]interface{}
		json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)
		log.Debug(jsonObj)

		if jsonObj["method"] == nil || jsonObj["params"] == nil {
			set_resp(nil, "Invalid params")
			break
		}

		coin := c.Ctx.Input.Param(":coin")
		log.Debug(coin)
		if SupportMethod[coin] == nil {
			// 解析支持方法
			SupportMethod[coin] = make(map[string]bool)
			support := beego.AppConfig.String(coin + "::support")
			methods := strings.Split(support, ",")
			for i := 0; i < len(methods); i++ {
				SupportMethod[coin][methods[i]] = true
			}
		}

		// 验证方法是不是支持
		if SupportMethod[coin] == nil || !SupportMethod[coin][jsonObj["method"].(string)] {
			set_resp(nil, "Invalid method")
			break
		}

		url := beego.AppConfig.String(coin + "::rpcurl")
		req := httplib.Post(url).SetTimeout(time.Second*10, time.Second*10)
		if beego.AppConfig.String(coin+"::rpcuser") != "" && beego.AppConfig.String(coin+"::rpcpass") != "" {
			req.SetBasicAuth(beego.AppConfig.String(coin+"::rpcuser"), beego.AppConfig.String(coin+"::rpcpass"))
		}
		req.JSONBody(jsonObj)
		result, err := req.Bytes()
		if err != nil {
			set_resp(nil, err.Error())
		} else {
			resp, _ := req.Response()
			if resp.StatusCode != 200 {
				set_resp(nil, resp.Status)
			} else {
				var tmp map[string]interface{}
				json.Unmarshal(result, &tmp)
				c.Data["json"] = tmp
				c.ServeJSON()
				return
			}
		}

		break
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
