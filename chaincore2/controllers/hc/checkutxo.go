package hc

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/chaincore2/common"
	"github.com/group-coldwallet/chaincore2/dao/daohc"
	"github.com/group-coldwallet/common/log"
)

type CheckTXController struct {
	beego.Controller
}

// request json
//{
// 	"result":[],
//}
func (c *CheckTXController) Get() {
	// 返回数据
	newresp := map[string]interface{}{
		"code":    0,
		"message": "ok",
		"data":    0, // 有效 0，无效1
	}

	txid := c.Input().Get("txid")
	vout := c.Input().Get("vout")
	if txid == "" || vout == "" {
		beego.Error("param error")
		c.Data["json"] = newresp
		c.ServeJSON()
		return
	}

	txvout := daohc.NewBlockTXVout()
	ret, err := txvout.Select(txid, common.StrToInt64(vout))
	if ret == false || err != nil {
		beego.Error(err)

		// 查询交易是否在交易池
		for true {
			respdata, err := common.Request("getrawtransaction", []interface{}{txid, 0})
			if err != nil {
				beego.Error(err)
				break
			} else {
				//log.Debug(string(respdata))
			}
			var datas map[string]interface{}
			err = json.Unmarshal(respdata, &datas)
			if err != nil || datas["error"] != nil {
				log.Debug(err, datas["error"])
				break
			}

			if datas["result"] != nil {
				newresp["data"] = 1
			}
			break
		}

	} else {
		if txvout.Status == 2 {
			newresp["data"] = 1
		}
	}

	c.Data["json"] = newresp
	c.ServeJSON()
}
