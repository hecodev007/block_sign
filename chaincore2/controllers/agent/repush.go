package agent

import (
	"encoding/json"
	"github.com/astaxie/beego/httplib"
	"log"
	"time"

	//"encoding/hex"

	"github.com/astaxie/beego"
)

type RepushController struct {
	beego.Controller
}

func (c *RepushController) Post() {
	// 返回数据
	resp := map[string]interface{}{
		"code":    0,
		"message": "",
		"data":    nil,
	}

	set_resp := func(code int, msg string) {
		resp["code"] = code
		resp["message"] = msg
	}

	for true {
		var jsonObj map[string]interface{}
		json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)
		//log.Debug(jsonObj)

		if jsonObj["uid"] == nil || jsonObj["coin"] == nil || jsonObj["txid"] == nil {
			set_resp(1, "param error")
			break
		}

		coin := jsonObj["coin"].(string)
		url := beego.AppConfig.String(coin + "::repushurl")
		req := httplib.Post(url).SetTimeout(time.Second*3, time.Second*10)
		req.SetBasicAuth(beego.AppConfig.String(coin+"::rpcuser"), beego.AppConfig.String(coin+"::rpcpass"))

		datas := map[string]interface{}{
			"uid":        jsonObj["uid"],
			"txid":       jsonObj["txid"],
			"height":     jsonObj["height"],
			"isInternal": jsonObj["isInternal"],
		}
		req.JSONBody(datas)
		dd, _ := json.Marshal(datas)
		log.Printf("repush  send datas:%s \n", string(dd))
		result, err := req.Bytes()
		if err != nil {
			set_resp(1, err.Error())
		} else {
			resp, _ := req.Response()
			if resp.StatusCode != 200 {
				set_resp(1, resp.Status)
			} else {
				var tmp map[string]interface{}
				json.Unmarshal(result, &tmp)
				c.Data["json"] = tmp
				c.ServeJSON()
				return
			}
		}

		break
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
