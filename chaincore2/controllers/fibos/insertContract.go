package fibos

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/common/log"

	service "github.com/group-coldwallet/chaincore2/service/fibos"
)

type InsertContractController struct {
	beego.Controller
}

func (c *InsertContractController) Post() {
	// 返回数据
	resp := map[string]interface{}{
		"code":    0,
		"message": "contract register success",
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

		if jsonObj["name"] == nil || jsonObj["contract_address"] == nil || jsonObj["decimal"] == nil || jsonObj["coin_type"] == nil {
			set_resp(1, "param error")
			break
		}

		name := jsonObj["name"].(string)
		contract := jsonObj["contract_address"].(string)
		decimal := int(jsonObj["decimal"].(float64))
		coin := jsonObj["coin_type"].(string)
		if coin != beego.AppConfig.String("coin") {
			set_resp(1, "coin error")
			break
		}

		service.InsertContract(name, decimal, contract)
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
