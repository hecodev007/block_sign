package controllers

import (
	//"encoding/hex"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/group-coldwallet/chaincore2/common"
	"github.com/group-coldwallet/chaincore2/models"
)

type InfoController struct {
	beego.Controller
}

func (c *InfoController) Post() {
	// 返回数据
	resp := map[string]interface{}{
		"code":    1,
		"message": "fail",
		"data":    nil,
	}

	if beego.AppConfig.DefaultBool("enabledb", true) {
		var maps []orm.Params
		o := orm.NewOrm()
		heart := &models.HeartInfo{"", 0}
		heart.Coin = beego.AppConfig.DefaultString("coin", "unknow")
		num, err := o.Raw("select max(height) as maxindex from block_info").Values(&maps)
		if err == nil && num > 0 {
			if maps[0]["maxindex"] == nil {
				heart.Height = 0
			} else {
				heart.Height = common.StrToInt64(maps[0]["maxindex"].(string))
			}
		}

		resp["data"] = heart
		resp["code"] = 0
		resp["message"] = "ok"
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
