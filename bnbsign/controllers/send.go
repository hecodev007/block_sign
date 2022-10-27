package controllers

import (
	"encoding/json"
	"github.com/astaxie/beego"
	_ "github.com/astaxie/beego/httplib"
	_ "github.com/binance-chain/go-sdk/client"
	"github.com/binance-chain/go-sdk/client/basic"
)

type SendController struct {
	beego.Controller
}

func (c *SendController) Post() {
	// 返回数据
	newresp := map[string]interface{}{
		"code": 0,
		"message":  "ok",
		"data": nil,
	}

	set_resp := func(code int, msg string) {
		newresp["code"] = code
		newresp["message"] = msg
	}

	for true {
		var jsonObj map[string]interface{}
		json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)
		beego.Debug(jsonObj)


		if jsonObj["txhex"] == nil || jsonObj["txhex"].(string) == "" {
			set_resp(1, "param error")
			break
		}

		txhex := jsonObj["txhex"].(string)
		param := map[string]string{
			"sync":"true",
		}
		client := basic.NewClient(beego.AppConfig.String("url"))
		result, err := client.PostTx([]byte(txhex), param)
		if err != nil {
			set_resp(1, err.Error())
		}

		if len(result) > 0 {
			if result[0].Code == 0 && result[0].Ok {
				newresp["data"] = result[0].Hash
			}
		}

		break
	}

	c.Data["json"] = newresp
	c.ServeJSON()
}

func (c *SendController) Get() {
	// 返回数据
	newresp := map[string]interface{}{
		"code": 0,
		"message":  "ok",
		"data": nil,
	}

	set_resp := func(code int, msg string) {
		newresp["code"] = code
		newresp["message"] = msg
	}

	for true {
		txhex := c.Input().Get("txhex")
		if txhex == "" {
			set_resp(1, "param error")
			break
		}

		param := map[string]string{
			"sync":"true",
		}
		client := basic.NewClient(beego.AppConfig.String("url"))
		result, err := client.PostTx([]byte(txhex), param)
		if err != nil {
			set_resp(1, err.Error())
		}

		if len(result) > 0 {
			if result[0].Code == 0 && result[0].Ok {
				newresp["data"] = result[0].Hash
			}
		}

		//req := httplib.Post("https://dex.binance.org/api/v1/broadcast")
		//req.Header("Content-Type", "text/plain")
		//req.Body(txhex)
		//result, err := req.Bytes()
		//if err != nil {
		//	set_resp(1, err.Error())
		//} else {
		//	resp, _ := req.Response()
		//	if resp.StatusCode != 200 {
		//		set_resp(1, resp.Status)
		//	} else {
		//		var tmp []interface{}
		//		json.Unmarshal(result, &tmp)
		//		c.Data["json"] = tmp[0]
		//		c.ServeJSON()
		//		return
		//	}
		//}

		break
	}

	c.Data["json"] = newresp
	c.ServeJSON()
}
