package addrManager

import (
	"encoding/json"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/orm"
	"time"

	//"encoding/hex"

	"github.com/astaxie/beego"
	"github.com/group-coldwallet/common/log"
)

type RemoveController struct {
	beego.Controller
}


func (c *RemoveController) Post() {
	// 返回数据
	resp := map[string]interface{}{
		"code":1,
		"message":"remove contract fail",
		"data":nil,
	}

	set_resp := func(code int, msg string) {
		resp["code"] = code
		resp["message"] = msg
	}

	coin := ""
	for true {
		var jsonObj map[string]interface{}
		json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)
		log.Debug(jsonObj)

		if jsonObj["contract_address"] == nil || jsonObj["coin_type"] == nil {
			set_resp(1, "param error")
			break
		}

		contract := jsonObj["contract_address"].(string)
		coin = jsonObj["coin_type"].(string)

		o := orm.NewOrm()
		var maps []orm.Params
		nums, err := o.Raw("select contract_address from contract_info where contract_address = ? and coin_type = ?", contract, coin).Values(&maps)
		if err == nil && nums > 0 {
			res, err := o.Raw("update contract_info set invaild = 1 where contract_address = ? and coin_type = ?", contract, coin).Exec()
			if err != nil {
				log.Debug(err)
				break
			}

			num, _ := res.RowsAffected()
			if num == 0 {
				break
			}
		} else {
			break
		}

		set_resp(0, "remove contract success")
		c.Data["json"] = resp
		c.ServeJSON()

		// post to data server
		url := beego.AppConfig.String(coin + "::removeurl")
		if url == "" {
			log.Debug(coin + " config not found !")
			return
		}
		req := httplib.Post(url).SetTimeout(time.Second*3, time.Second*10)
		if beego.AppConfig.String(coin + "::rpcuser") != "" && beego.AppConfig.String(coin + "::rpcpass") != "" {
			req.SetBasicAuth(beego.AppConfig.String(coin + "::rpcuser"), beego.AppConfig.String(coin + "::rpcpass"))
		}
		req.JSONBody(
			map[string]interface{}{
				"contract_addresses": []string {
					contract,
				},
			},
		)
		result, err := req.Bytes()
		if err != nil {
			log.Debug(err)
		} else {
			log.Debug(string(result))
		}

		return
	}

	c.Data["json"] = resp
	c.ServeJSON()
}



