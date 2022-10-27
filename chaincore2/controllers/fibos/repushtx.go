package fibos

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daofibos"
	"github.com/group-coldwallet/chaincore2/models"
	service "github.com/group-coldwallet/chaincore2/service/fibos"
	"github.com/group-coldwallet/common/log"
	"strings"
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
		var symbol string = ""
		var actions []models.PushEosAction
		var tmpWatchList map[string]bool = make(map[string]bool)
		{
			o := orm.NewOrm()
			var maps []orm.Params
			nums, err := o.Raw("select txid, height, hash, contract, fromaccount, toaccount,amount,memo from block_tx where txid = ?", txid).Values(&maps)
			if err == nil && nums > 0 {
				for i := 0; i < len(maps); i++ {
					if i == 0 {
						height = common.StrToInt64(maps[i]["height"].(string))
						hash = maps[i]["hash"].(string)
						txid = maps[i]["txid"].(string)
					}

					var b models.PushEosAction
					b.Contract = maps[i]["contract"].(string)
					b.From = maps[i]["fromaccount"].(string)
					b.To = maps[i]["toaccount"].(string)
					b.Memo = maps[i]["memo"].(string)
					amount := maps[i]["amount"].(string)

					tmpWatchList[b.From] = true
					tmpWatchList[b.To] = true
					_tmp := strings.Split(amount, " ")
					if len(_tmp) > 1 {
						symbol = _tmp[1]
					}
					b.Amount = _tmp[0]
					b.Token = symbol

					actions = append(actions, b)
				}
			}
		}

		blockInfo := dao.NewBlockInfo()
		err := blockInfo.GetBlockInfoByIndex(height)
		if err != nil {
			set_resp(1, "data not found")
			break
		}

		pushBlockTx := new(models.PushEosBlockInfo)
		pushBlockTx.Type = models.PushTypeEosTX
		pushBlockTx.Height = height
		pushBlockTx.Hash = hash
		pushBlockTx.CoinName = beego.AppConfig.String("coin")
		pushBlockTx.Confirmations = blockInfo.Confirmations
		pushBlockTx.Time = blockInfo.Timestamp
		var pushtx models.PushEosTx
		pushtx.Txid = txid
		pushtx.Status = "executed"
		pushtx.Fee = 0
		pushtx.Actions = actions
		pushBlockTx.Txs = append(pushBlockTx.Txs, pushtx)

		pusdata, err := json.Marshal(&pushBlockTx)
		if err == nil {
			if len(tmpWatchList) > 0 {
				service.AddPushTask(height, txid, tmpWatchList, pusdata)
			}
		}

		break
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
