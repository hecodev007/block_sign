package stacks

import (
	"github.com/astaxie/beego"
)

type RpcController struct {
	beego.Controller
}


func (c *RpcController) Post() {
	// 返回数据
	resp := map[string]interface{}{
		"id":1,
		"jsonrpc":"2.0",
		"result":"",
	}

	c.Data["json"] = resp
	c.ServeJSON()
}


