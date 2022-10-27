package ruby

import (
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/chaincore2/common"
	"github.com/group-coldwallet/chaincore2/service/ruby"
)

type RebackController struct {
	beego.Controller
}

func (c *RebackController) Get() {
	// 返回数据
	resp := map[string]interface{}{
		"code":    0,
		"message": "ok",
		"data":    nil,
	}

	height := c.Input().Get("height")
	if height != "" {
		ruby.SyncBlockData(common.StrToInt64(height))
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
