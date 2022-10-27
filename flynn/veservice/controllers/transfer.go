package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"veservice/common"
	"veservice/models"
)

type TransferController struct {
	MainController
}

func (c *TransferController) Post() {

	// 返回数据
	resp := map[string]interface{}{
		"result": nil,
		"error":  nil,
	}

	//
	var hexs []*models.EthRawData
	for true {
		var jsonObj map[string]interface{}
		json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)

		if jsonObj["data"] == nil {
			break
		}
		switch v := jsonObj["data"].(type) {
		case []interface{}:
			list := jsonObj["data"].([]interface{})
			for i := 0; i < len(list); i++ {
				obj := list[i].(map[string]interface{})
				hex, errS := c.Sign(obj["data"].(map[string]interface{}))
				if errS != nil {
					resp["error"] = fmt.Errorf("transfer error,err=%v", errS)
					c.Data["json"] = resp
					c.ServeJSON()
					return
				}
				tx := new(models.EthRawData)
				tx.Hex = hex
				tx.Index = i
				hexs = append(hexs, tx)
			}
			break

		case map[string]interface{}:
			hex, errS := c.Sign(jsonObj["data"].(map[string]interface{}))
			if errS != nil {
				resp["error"] = fmt.Errorf("transfer error,err=%v", errS)
				c.Data["json"] = resp
				c.ServeJSON()
				return
			}
			tx := new(models.EthRawData)
			tx.Hex = hex
			tx.Index = 0
			hexs = append(hexs, tx)

		default:
			beego.Debug(v)
		}
		break
	}
	//广播交易
	if hexs == nil || len(hexs) == 0 {
		resp["error"] = "transfer error"
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}
	var reqParams = "0x"
	for _, h := range hexs {
		reqParams += removeHex0x(h.Hex)
	}
	if len(reqParams) == 2 {
		resp["error"] = "build tx params error"
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}
	url := fmt.Sprintf("%s/transactions", beego.AppConfig.String("url"))
	reqbody := map[string]interface{}{
		"raw": reqParams,
	}
	respData, err := common.PostJson(url, reqbody)
	if err != nil || len(respData) == 0 {
		resp["error"] = fmt.Sprintf("broadcast error,err=%v", err)
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}
	var res map[string]string
	json.Unmarshal(respData, &res)

	resp["result"] = res["id"]
	c.Data["json"] = resp
	c.ServeJSON()
}
