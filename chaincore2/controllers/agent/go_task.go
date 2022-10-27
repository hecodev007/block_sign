package agent

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
)

type GoTaskController struct {
	beego.Controller
}

type GoTask struct {
	OpenPush bool `json:"push"`
}

func (c *GoTaskController) Post() {
	resp := map[string]interface{}{
		"code":    0,
		"message": "ok",
		"data":    nil,
	}

	set_resp := func(result interface{}, err interface{}) {
		resp["result"] = result
		if err != nil {
			resp["error"] = map[string]interface{}{"code": -1, "message": err.(string)}
		}
	}
	var params GoTask
	var err error
	if err = json.Unmarshal(c.Ctx.Input.RequestBody, &params); err == nil {
		if params.OpenPush {
			TestUrl = beego.AppConfig.DefaultString("testurl", "")
			resp["data"] = "open"
		} else {
			TestUrl = ""
			resp["data"] = "close"
		}
		c.Data["json"] = resp
	} else {
		set_resp(nil, fmt.Errorf("Invalid params,err:%s", err.Error()))
	}
	c.ServeJSON()
}
