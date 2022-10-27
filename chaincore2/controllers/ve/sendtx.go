package ve

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/chaincore2/common"
	"github.com/group-coldwallet/common/log"
)

type SendTXController struct {
	beego.Controller
}

func (c *SendTXController) Post() {
	// 返回数据
	newresp := map[string]interface{}{
		"code":    0,
		"message": "ok",
		"data":    nil,
	}

	// 返回数据
	resp := map[string]interface{}{
		"txid": nil,
	}

	for true {
		var jsonObj map[string]interface{}
		json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)
		log.Debug(jsonObj)

		result := jsonObj["result"].(string)
		if result == "" {
			newresp["code"] = 1
			newresp["message"] = "Request param error"
			break
		}

		respdata, err := common.Request("sendrawtransaction", []interface{}{result})
		if err != nil {
			newresp["code"] = 1
			newresp["message"] = err.Error
			beego.Error(err)
			break
		} else {
			log.Debug(string(respdata))
			var datas map[string]interface{}
			err = json.Unmarshal(respdata, &datas)
			if err != nil {
				newresp["code"] = 1
				newresp["message"] = err.Error
				beego.Error(err)
				break
			}

			if datas["result"] == nil {
				if datas["error"] != nil {
					resperr := datas["error"].(map[string]interface{})
					newresp["code"] = 1
					newresp["message"] = resperr["message"].(string)
					break
				}
				newresp["code"] = 1
				newresp["message"] = "fail"
				break
			}

			resp["txid"] = datas["result"].(string)
		}

		break
	}

	newresp["data"] = resp
	c.Data["json"] = newresp
	c.ServeJSON()
}
