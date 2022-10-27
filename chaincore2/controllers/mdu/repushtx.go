package mdu

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daomdu"
	"github.com/group-coldwallet/chaincore2/models"
	service "github.com/group-coldwallet/chaincore2/service/mdu"
	"github.com/group-coldwallet/common/log"
)

type RepushTxController struct {
	beego.Controller
}

func (c *RepushTxController) Post() {
	// 返回数据
	resp := map[string]interface{}{
		"code":    0,
		"message": "ok",
		"data":    nil,
	}

	set_resp := func(code int, message string) {
		resp["code"] = code
		resp["message"] = message
	}

	var jsonObj map[string]interface{}
	json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)
	log.Debug(jsonObj)

	for true {
		if jsonObj["txid"] == nil || jsonObj["uid"] == nil {
			set_resp(1, "param error")
			break
		}

		// 读取 txid
		txid := jsonObj["txid"].(string)
		uid := int64(jsonObj["uid"].(float64))
		if service.UserWatchList[uid] == nil {
			set_resp(1, "user not found")
			break
		}

		var height int64 = 0
		var hash string = ""
		var pushtxs []models.PushAccountTx
		var tmpWatchList map[string]bool = make(map[string]bool)
		{
			o := orm.NewOrm()
			var maps []orm.Params
			nums, err := o.Raw("select txid, height, hash, sys_fee, fromaccount, toaccount,amount,memo,contractaddress from block_tx where txid = ?", txid).Values(&maps)
			if err == nil && nums > 0 {
				for i := 0; i < len(maps); i++ {
					height = common.StrToInt64(maps[i]["height"].(string))
					hash = maps[i]["hash"].(string)

					var b models.PushAccountTx
					b.Txid = maps[i]["txid"].(string)
					b.Fee = common.StrToFloat64(maps[i]["sys_fee"].(string))
					b.From = maps[i]["fromaccount"].(string)
					b.To = maps[i]["toaccount"].(string)
					b.Memo = maps[i]["memo"].(string)
					b.Amount = maps[i]["amount"].(string)
					b.Contract = maps[i]["contractaddress"].(string)

					pushtxs = append(pushtxs, b)
					tmpWatchList[b.From] = true
					tmpWatchList[b.To] = true
				}
			}
		}

		blockInfo := dao.NewBlockInfo()
		err := blockInfo.GetBlockInfoByIndex(height)
		if err != nil {
			set_resp(1, "data not found")
			break
		}

		pushBlockTx := new(models.PushAccountBlockInfo)
		pushBlockTx.Type = models.PushTypeAccountTX
		pushBlockTx.Height = height
		pushBlockTx.Hash = hash
		pushBlockTx.CoinName = beego.AppConfig.String("coin")
		pushBlockTx.Confirmations = blockInfo.Confirmations
		pushBlockTx.Time = blockInfo.Timestamp
		pushBlockTx.Txs = pushtxs

		pusdata, err := json.Marshal(&pushBlockTx)
		if err == nil {
			service.AddPushTask(height, hash, tmpWatchList, pusdata)
		}

		break
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
