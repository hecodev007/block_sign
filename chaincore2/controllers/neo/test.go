package neo

import (
	"encoding/json"
	//"encoding/hex"

	"github.com/astaxie/beego"
	"github.com/group-coldwallet/common/log"
)

type TestController struct {
	beego.Controller
}


func (c *TestController) Post() {
	// 返回数据
	resp := map[string]interface{}{
		"code":0,
		"message":"",
	}

	for true {
		var jsonObj map[string]interface{}
		json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)
		log.Debug(jsonObj)

		break
	}

	c.Data["json"] = resp
	c.ServeJSON()
}


