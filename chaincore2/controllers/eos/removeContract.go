package eos

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/common/log"

	service "github.com/group-coldwallet/chaincore2/service/eos"
)

type RemoveContractController struct {
	beego.Controller
}

func (c *RemoveContractController) Post() {
	// 返回数据
	resp := map[string]interface{}{
		"code":    0,
		"message": "contract remove success",
		"data":    nil,
	}

	set_resp := func(code int, msg string) {
		resp["code"] = code
		resp["message"] = msg
	}

	for true {
		var jsonObj map[string]interface{}
		json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)
		log.Debug(jsonObj)

		if jsonObj["contract_address"] == nil || jsonObj["coin_type"] == nil {
			set_resp(1, "param error")
			break
		}

		contract := jsonObj["contract_address"].(string)
		coin := jsonObj["coin_type"].(string)
		if coin != beego.AppConfig.String("coin") {
			set_resp(1, "coin error")
			break
		}

		service.RemoveContract(contract)
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
