package ar

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Dev43/arweave-go/api"
	"github.com/Dev43/arweave-go/utils"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daoar"
	"github.com/group-coldwallet/chaincore2/models"
	service "github.com/group-coldwallet/chaincore2/service/ar"
	"github.com/group-coldwallet/common/log"
	"github.com/shopspring/decimal"
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

type RepushTxWithHeightController struct {
	beego.Controller
}


func (c *RepushTxWithHeightController) Post() {
	var time = int64(0)
	// 返回数据
	resp := map[string]interface{}{
		"code":    1,
		"message": "fail",
		"data":    nil,
	}

	set_resp := func(code int, message string) {
		resp["code"] = code
		resp["message"] = message
	}

	var jsonObj map[string]interface{}
	json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)

	for true {
		if jsonObj["txid"] == nil || jsonObj["uid"] == nil|| jsonObj["height"] == nil {
			set_resp(1, "param error")
			break
		}

		// 读取 txid
		txid := jsonObj["txid"].(string)
		uid := int64(jsonObj["uid"].(float64))
		height := int64(jsonObj["height"].(float64))
		if service.UserWatchList[uid] == nil {
			set_resp(1, "user not found")
			break
		}

		url := beego.AppConfig.String("nodeurl")
		client, err := api.Dial(url)
		if err != nil {
			break
		}

		// 获取详细交易数据
		blocktx := dao.NewBlockTX()

		var hash string = ""
		var pushtxs []models.PushAccountTx
		// zh 改下，不要从数据库拿，从节点拿
		var tmpWatchList map[string]bool = make(map[string]bool)
		{

			transaction,err2 := client.GetTransaction(context.TODO(),txid)
			if err2 != nil {
				log.Error(err)
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
				//hash = transaction.Hash()

				url := beego.AppConfig.String("nodeurl")
				client2, err := api.Dial(url)
				if err != nil {
					log.Error(err)
					break
				}
				blockInfo,err := client2.GetBlockByHeight(context.TODO(),height)
				if err != nil {
					log.Error(err)
					break
				}
				hash = blockInfo.IndepHash
				var hasTxt = false
				for _,v := range blockInfo.Txs { //这段代码是遍历区块的交易ID查询传入的交易ID是否在区块交易ID列表中
					txid2 := fmt.Sprintf("%s", v)
					if txid2 == txid {
						hasTxt = true
					}
				}
				if !hasTxt {
					log.Error(errors.New("the block has not the txid@"))
					resp["message"] = "the block has not the txid"
					break
				}
				time = int64(blockInfo.Timestamp)
				b.Txid = utils.EncodeToBase64(transaction.ID())

				pushtxs = append(pushtxs, b)
				tmpWatchList[b.From] = true
				tmpWatchList[b.To] = true

				blocktx.Hash = hash
				blocktx.From = b.From
				blocktx.To = b.To
				blocktx.Amount = b.Amount
				blocktx.Height = height
				blocktx.Sysfee = common.Float64ToString(b.Fee)
				blocktx.Txid = txid
			}
		}

		if len(tmpWatchList) > 0 {
			pushBlockTx := new(models.PushAccountBlockInfo)
			pushBlockTx.Type = models.PushTypeAccountTX
			pushBlockTx.Height = height
			pushBlockTx.Hash = hash
			pushBlockTx.CoinName = beego.AppConfig.String("coin")
			pushBlockTx.Confirmations = beego.AppConfig.DefaultInt64("confirmations", 6)
			pushBlockTx.Time = time
			pushBlockTx.Txs = pushtxs

			pusdata, err := json.Marshal(&pushBlockTx)
			if err == nil {
				service.AddPushTask(height, hash, tmpWatchList, pusdata)
			}

			has,err := blocktx.Exist(txid)
			if err != nil {
				beego.Error(err)
				break
			}
			if has { //有则更新，无则插入
				_, err := blocktx.Update()
				if err != nil {
					//fmt.Println("256:",err.Error())
					beego.Error(err)
				}
			} else {
				num, err := blocktx.Insert()
				if num <= 0 || err != nil {
					beego.Error(err)
				}
			}
			resp["code"] = 0
			resp["message"] = "ok"
		}

		break
	}

	c.Data["json"] = resp
	c.ServeJSON()
}