package ar

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/chaincore2/common"
	service "github.com/group-coldwallet/chaincore2/service/ar"
)

type UpdateController struct {
	beego.Controller
}

func (c *UpdateController) Post() {
	// 返回数据
	resp := map[string]interface{}{
		"code":    0,
		"message": "ok",
		"data":    nil,
	}

	var jsonObj map[string]interface{}
	json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)

	if  jsonObj["uid"] == nil|| jsonObj["url"] == nil {
		resp["code"] = 1
		resp["message"] = "fail"
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}
	service.UpdateWatchAddress(int64(jsonObj["uid"].(float64)),jsonObj["url"].(string))

	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *UpdateController) Get() {
	// 返回数据
	resp := map[string]interface{}{
		"code":    0,
		"message": "ok",
		"data":    nil,
	}

	uid := c.Input().Get("uid")
	url := c.Input().Get("url")

	if uid != "" && url != "" {
		service.UpdateWatchAddress(common.StrToInt64(uid), url)
	} else {
		resp["code"] = 1
		resp["message"] = "Param address error"
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
