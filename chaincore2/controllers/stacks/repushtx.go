package stacks

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daostacks"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/chaincore2/service/stacks"
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
		if stacks.UserWatchList[uid] == nil {
			set_resp(1, "user not found")
			break
		}

		blocktx := dao.NewBlockTX()
		if blocktx.SelectCount(txid) == 0 {
			// 获取原始交易信息
			log.Debug(txid)
			respdata, err := common.Request("getrawtransaction", []interface{}{txid, 1})
			if err != nil {
				beego.Error(err)
				continue
			} else {
				//log.Debug(string(respdata))
			}

			var datas map[string]interface{}
			err = json.Unmarshal(respdata, &datas)
			if err != nil || datas["error"] != nil {
				log.Debug(err, datas["error"])
				continue
			}

			if datas["result"] == nil {
				continue
			}

			tx := datas["result"].(map[string]interface{})
			blockhash := tx["blockhash"].(string)

			_blockInfo := dao.NewBlockInfo()
			if _blockInfo.GetBlockCountByHash(blockhash) > 0 {
				// 解析指定txid
				blockdata, err := common.RequestStr("getblock", []interface{}{blockhash, 1})
				if err != nil {
					beego.Error(err)
					return
				} else {
					//log.Debug(respdata)
				}

				var datas map[string]interface{}
				err = json.Unmarshal([]byte(blockdata), &datas)
				if err != nil {
					log.Debug(err)
					continue
				}

				if datas["result"] == nil {
					continue
				}

				// 区块详情
				result := datas["result"].(map[string]interface{})

				highindex, hash := int64(result["height"].(float64)), result["hash"].(string)
				blockInfo := stacks.Parse_block(result, false)
				if blockInfo == nil {
					continue
				}

				err = stacks.Parse_block_tx_todb(0, hash, highindex, tx, blockInfo)
				if err != nil {
					continue
				}
			} else {
				// 解析整个高度
				stacks.SyncBlockDataHash(blockhash)
			}

			c.Data["json"] = resp
			c.ServeJSON()
			return
		}

		var pushtxs []models.PushAccountTx
		var tmpWatchList map[string]bool = make(map[string]bool)
		{
			o := orm.NewOrm()
			var maps []orm.Params
			nums, err := o.Raw("select txid, height, hash, sys_fee, fromaccount, toaccount,amount,memo from block_tx where txid = ?", txid).Values(&maps)
			if err == nil && nums > 0 {
				for i := 0; i < len(maps); i++ {
					var b models.PushAccountTx
					b.Txid = maps[i]["txid"].(string)
					b.Fee = stacks.GetValue(common.StrToFloat64(maps[i]["sys_fee"].(string)))
					b.From = stacks.ConvertAdress("btc", maps[i]["fromaccount"].(string))
					b.To = stacks.ConvertAdress("btc", maps[i]["toaccount"].(string))
					b.Amount = stacks.GetValueStr(common.StrToFloat64(maps[i]["amount"].(string)))
					b.Memo = maps[i]["memo"].(string)

					pushtxs = append(pushtxs, b)
					tmpWatchList[maps[i]["fromaccount"].(string)] = true
					tmpWatchList[maps[i]["toaccount"].(string)] = true
				}
			}
		}

		blockInfo := dao.NewBlockInfo()
		err := blockInfo.GetBlockInfoByIndex(blocktx.Height)
		if err != nil {
			set_resp(1, "data not found")
			break
		}

		pushBlockTx := new(models.PushAccountBlockInfo)
		pushBlockTx.Type = models.PushTypeAccountTX
		pushBlockTx.Height = blockInfo.Height
		pushBlockTx.Hash = blockInfo.Hash
		pushBlockTx.CoinName = beego.AppConfig.String("coin")
		pushBlockTx.Confirmations = blockInfo.Confirmations + 1
		pushBlockTx.Time = blockInfo.Timestamp
		pushBlockTx.Txs = pushtxs

		pusdata, err := json.Marshal(&pushBlockTx)
		if err == nil {
			stacks.AddPushTask(blocktx.Height, blocktx.Txid, tmpWatchList, pusdata)
		}
		break
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
