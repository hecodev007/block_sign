package rsk

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daorsk"
	"github.com/group-coldwallet/chaincore2/models"
	service "github.com/group-coldwallet/chaincore2/service/rsk"
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
		// zh 改下，不要从数据库拿，从节点拿
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
		pushBlockTx.Time = common.StrToTime(blockInfo.Timestamp)
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

/*

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

		url := beego.AppConfig.String("nodeurl")
		client, err := api.Dial(url)
		if err != nil {
			break
		}

		var height int64 = 0
		var hash string = ""
		var pushtxs []models.PushAccountTx
		// zh 改下，不要从数据库拿，从节点拿
		var tmpWatchList map[string]bool = make(map[string]bool)
		{

			transaction,err2 := client.GetTransaction(context.TODO(),txid)
			if err2 != nil {
				fmt.Println("158:",err2.Error())
				break
			}
			if transaction != nil {
				var b models.PushAccountTx
				b.Txid = txid

				data,err:=base64.RawURLEncoding.DecodeString(transaction.Owner()) //获取From字段
				h:=sha256.New()
				h.Write(data)
				b.From = utils.EncodeToBase64(h.Sum(nil))
				b.To = transaction.Target() //转到哪而去

				fee,err := decimal.NewFromString(transaction.Reward()) //获取手续费
				if err != nil {
					log.Error(err)
					break
				}
				b.Fee,_ = fee.Div(decimal.New(1, 12)).Float64()

				//获取交易金额
				quantity,err := decimal.NewFromString(transaction.Quantity())
				if err != nil {
					log.Error(err)
					break
				}
				b.Amount = quantity.Div(decimal.New(1, 12)).String()


				hash = transaction.Hash()
				fmt.Printf("\n 188 hash:%s\n",hash)
				fmt.Printf("\n 188 ID:%s\n",utils.EncodeToBase64(transaction.ID()))
				fmt.Printf("\n 188 Owner:%s\n",transaction.Owner())

				b.Txid = utils.EncodeToBase64(transaction.ID())

				pushtxs = append(pushtxs, b)
				tmpWatchList[b.From] = true
				tmpWatchList[b.To] = true
			}
		}

		blockInfo,err := client.GetBlockByID(context.TODO(),hash) //Ldgnu2uz_wv1oCHEV5Y_8JlgCEst9AotPn3oub3X79Q
		if err != nil {
			log.Error(err)
			break
		}

		pushBlockTx := new(models.PushAccountBlockInfo)
		pushBlockTx.Type = models.PushTypeAccountTX
		pushBlockTx.Height = height
		pushBlockTx.Hash = hash
		pushBlockTx.CoinName = beego.AppConfig.String("coin")
		pushBlockTx.Confirmations = beego.AppConfig.DefaultInt64("confirmations", 6)
		pushBlockTx.Time = int64(blockInfo.Timestamp)
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
*/
