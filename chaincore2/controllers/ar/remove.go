package ar

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/chaincore2/common"
	service "github.com/group-coldwallet/chaincore2/service/ar"
	"github.com/group-coldwallet/common/log"
)

type RemoveController struct {
	beego.Controller
}

func (c *RemoveController) Post() {
	// 返回数据
	resp := map[string]interface{}{
		"code":    0,
		"message": "ok",
		"data":    nil,
	}

	var jsonObj []interface{}
	json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)
	log.Debug(jsonObj)

	for i := 0; i < len(jsonObj); i++ {
		obj := jsonObj[i].(map[string]interface{})
		service.RemoveWatchAddressByUserId(int64(obj["uid"].(float64)), obj["address"].(string))
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *RemoveController) Get() {
	// 返回数据
	resp := map[string]interface{}{
		"code":    0,
		"message": "ok",
		"data":    nil,
	}

	uid := c.Input().Get("uid")
	address := c.Input().Get("address")

	if address != "" {
		service.RemoveWatchAddressByUserId(common.StrToInt64(uid), address)
	} else {
		resp["code"] = 1
		resp["message"] = "Param address error"
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
