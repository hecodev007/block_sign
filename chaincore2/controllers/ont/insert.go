package ont

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/chaincore2/common"
	"github.com/group-coldwallet/chaincore2/service/ont"
	"github.com/group-coldwallet/common/log"
)

type InsertController struct {
	beego.Controller
}

func (c *InsertController) Post() {
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
		ont.InsertWatchAddress(int64(obj["uid"].(float64)), obj["address"].(string), obj["url"].(string))
	}

	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *InsertController) Get() {
	// 返回数据
	resp := map[string]interface{}{
		"code":    0,
		"message": "ok",
		"data":    nil,
	}

	uid := c.Input().Get("uid")
	address := c.Input().Get("address")
	url := c.Input().Get("url")

	if address != "" && url != "" {
		ont.InsertWatchAddress(common.StrToInt64(uid), address, url)
	} else {
		resp["code"] = 1
		resp["message"] = "Param address error"
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
