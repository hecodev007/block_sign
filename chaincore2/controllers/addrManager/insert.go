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

type InsertController struct {
	beego.Controller
}


func (c *InsertController) Post() {
	// 返回数据
	resp := map[string]interface{}{
		"code":1,
		"message":"contract register fail",
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

		if jsonObj["name"] == nil || jsonObj["contract_address"] == nil || jsonObj["decimal"] == nil || jsonObj["coin_type"] == nil {
			set_resp(1, "param error")
			break
		}

		name := jsonObj["name"].(string)
		contract := jsonObj["contract_address"].(string)
		decimal := int64(jsonObj["decimal"].(float64))
		coin = jsonObj["coin_type"].(string)

		o := orm.NewOrm()
		var maps []orm.Params
		nums, err := o.Raw("select contract_address from contract_info where contract_address = ? and coin_type = ?", contract, coin).Values(&maps)
		if err == nil && nums == 0 {
			res, err := o.Raw("insert into contract_info(name, contract_address, `decimal`, coin_type) values(?,?,?,?)", name, contract, decimal, coin).Exec()
			if err != nil {
				log.Debug(err)
				break
			}

			num, _ := res.RowsAffected()
			if num == 0 {
				break
			}
		}

		set_resp(0, "contract register success")
		c.Data["json"] = resp
		c.ServeJSON()

		// post to data server
		url := beego.AppConfig.String(coin + "::contracturl")
		if url == "" {
			log.Debug(coin + " config not found !")
			return
		}
		req := httplib.Post(url).SetTimeout(time.Second*3, time.Second*10)
		if beego.AppConfig.String(coin + "::rpcuser") != "" && beego.AppConfig.String(coin + "::rpcpass") != "" {
			req.SetBasicAuth(beego.AppConfig.String(coin + "::rpcuser"), beego.AppConfig.String(coin + "::rpcpass"))
		}
		req.JSONBody([]interface{}{
			map[string]interface{}{
				"name": name,
				"contract_address": contract,
				"decimal": decimal,
				"coin_type": coin,
			},
		})
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


