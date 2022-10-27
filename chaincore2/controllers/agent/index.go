package agent

import (
	"encoding/json"
	"github.com/astaxie/beego/httplib"
	"time"

	//"encoding/hex"

	"github.com/astaxie/beego"
	"log"
)

type IndexController struct {
	beego.Controller
}

func (c *IndexController) Post() {
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
		log.Println(jsonObj)

		if jsonObj["uid"] == nil || jsonObj["url"] == nil || jsonObj["data"] == nil {
			set_resp(1, "param error")
			break
		}
		if riskData,err := IsRisk([]byte(jsonObj["data"].(string)));err != nil {
			log.Printf("报警错误:%s",err.Error())
		} else if riskData != nil {
			log.Printf("原始数据:%s",jsonObj["data"].(string))
			log.Printf("报警数据:%s",string(riskData))
			//jsonObj["data"] = string(riskData)
		}
		url := jsonObj["url"].(string)
		req := httplib.Post(url).SetTimeout(time.Second*3, time.Second*10)
		req.SetBasicAuth("rylink", "fdhj&%@#13*74")
		req.Body([]byte(jsonObj["data"].(string)))
		result, err := req.Bytes()
		if err != nil {
			set_resp(1, err.Error())
		} else {
			resp, _ := req.Response()
			if resp.StatusCode != 200 {
				set_resp(1, resp.Status)
			} else {
				//测试备用
				//testurl := beego.AppConfig.DefaultString("testurl", "")
				if TestUrl != "" {
					req2 := httplib.Post(TestUrl).SetTimeout(time.Second*3, time.Second*10)
					req2.SetBasicAuth("rylink", "fdhj&%@#13*74")
					req2.Body([]byte(jsonObj["data"].(string)))
					result2, _ := req2.Bytes()
					if result2 != nil {
						log.Printf("success: testurl:%s,post return==>%s", TestUrl, string(result2))
					} else {
						log.Printf("fail:testurl:%s,post return==>%s", TestUrl, string(result2))
					}
				}

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
