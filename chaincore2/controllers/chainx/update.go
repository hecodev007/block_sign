package chainx

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/chaincore2/common"
	"github.com/group-coldwallet/chaincore2/service/chainx"
	"github.com/group-coldwallet/common/log"
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

	var jsonObj []interface{}
	json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)
	log.Debug(jsonObj)

	for i := 0; i < len(jsonObj); i++ {
		obj := jsonObj[i].(map[string]interface{})
		chainx.UpdateWatchAddress(int64(obj["uid"].(float64)), obj["url"].(string))
	}

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
		chainx.UpdateWatchAddress(common.StrToInt64(uid), url)
	} else {
		resp["code"] = 1
		resp["message"] = "Param address error"
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
