package ksm

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daoksm"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/chaincore2/service/ksm"
	"github.com/group-coldwallet/common/log"
	"github.com/shopspring/decimal"
	"reflect"
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
		if ksm.UserWatchList[uid] == nil {
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
					d, _ := decimal.NewFromString(maps[i]["sys_fee"].(string))
					b.Fee, _ = d.Float64() // ksm.GetValueFromStr(maps[i]["sys_fee"].(string))
					b.From = maps[i]["fromaccount"].(string)
					b.To = maps[i]["toaccount"].(string)
					b.Memo = maps[i]["memo"].(string)

					d2, _ := decimal.NewFromString(maps[i]["amount"].(string))
					b.Amount = d2.String()
					b.Contract = maps[i]["contractaddress"].(string)

					pushtxs = append(pushtxs, b)
					tmpWatchList[b.From] = true
					tmpWatchList[b.To] = true
				}
			} else if nums <= 0 {
				set_resp(1, "txid not found")
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
			ksm.AddPushTask(height, hash, tmpWatchList, pusdata)
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
	var (
		time     = int64(0)
		hasTrans = true
	)

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
		if jsonObj["txid"] == nil || jsonObj["uid"] == nil || jsonObj["height"] == nil {
			set_resp(1, "param error")
			break
		}

		// 读取 txid
		txid := jsonObj["txid"].(string)
		uid := int64(jsonObj["uid"].(float64))
		height := int64(jsonObj["height"].(float64))
		if ksm.UserWatchList[uid] == nil {
			set_resp(1, "user not found")
			break
		}

		// 获取详细交易数据
		blocktx := dao.NewBlockTX()

		var hash string = ""
		ksmClient := ksm.NewKsmBlock(beego.AppConfig.String("nodeurl"))
		hash, err := ksmClient.GethashkByHeight(height)
		if err != nil {
			log.Error(err)
			set_resp(1, err.Error())
			break
		}

		var pushtxs []models.PushAccountTx
		// 从节点拿
		var tmpWatchList map[string]bool = make(map[string]bool)
		{
			block1, err := ksm.GetBlockTransMethodsByHeight(height)
			if err != nil {
				log.Error(err)
				set_resp(1, err.Error())
				break
			}
			for _, v := range block1.Extrinsics { //遍历交易数据，获取时间
				//if v.Method == "balances.transfer" || v.Method == "balances.transferKeepAlive" {
				if v.Method.Pallet == "balances" && v.Method.Method == "transfer" {
					hasTrans = true
				}
			}

			if !hasTrans {
				set_resp(1, "has no transfer events in this block!")
				break
			}

			//当有交易信息时候捕捉交易信息
			block, err := ksm.GetBlockTransByHeight(height)
			if err != nil {
				log.Error(err)
				set_resp(1, err.Error())
				break
			}
			flag := true
			for _, v := range block.Extrinsics { //遍历交易数据，获取时间
				//if v.Method == "timestamp.set" {
				if v.Method.Pallet == "timestamp" || v.Method.Method == "set" {
					//args := v.Args.(map[string]interface{})
					args := v.Args.Now
					if len(args) > 0 {
						timeStr := args
						time = common.StrToInt64(timeStr) / 1000
					}
				}

			}

			for _, v := range block.Extrinsics { //遍历交易数据,获取真正的转账数据
				log.Debug("repush txid", txid)
				dd, _ := json.Marshal(v)
				log.Debug("data extrinsics:", string(dd))

				if v.Hash == txid { //只检查有效的交易
					//if v.Method == "balances.transfer" || v.Method == "balances.transferKeepAlive" {
					if v.Method.Pallet == "balances" && (v.Method.Method == "transfer" || v.Method.Method == "transferKeepAlive") {
						flag = false
						if v.Success {
							var (
								b      models.PushAccountTx
								amount decimal.Decimal
							)
							s := v.Args.Dest
							dd, _ := json.Marshal(s)
							if reflect.TypeOf(s) != nil && reflect.TypeOf(s).String() != "map[string]interface {}" {
								log.Infof("不支持的数据类型 %s，暂时不解析,内容：%s", reflect.TypeOf(s), string(dd))
								continue
							}
							ksmDest := new(ksm.Dest)
							err = json.Unmarshal(dd, ksmDest)
							if err != nil || ksmDest.ID == "" {
								log.Infof("错误解析内容:%s", string(dd))
								continue
							}

							b.Txid = txid
							b.From = v.Signature.Signer.Id
							b.To = ksmDest.ID
							amount, _ = decimal.NewFromString(v.Args.Value)
							b.Amount = amount.Shift(-1 * models.KSM_DECIMAL).String()
							fee, _ := decimal.NewFromString(v.Info.PartialFee)
							b.Fee, _ = fee.Shift(-1 * models.KSM_DECIMAL).Float64() //手续费
							pushtxs = append(pushtxs, b)
							tmpWatchList[b.From] = true
							tmpWatchList[b.To] = true

							blocktx.Hash = hash
							blocktx.From = b.From
							blocktx.To = b.To
							blocktx.Amount = b.Amount
							blocktx.Height = height
							blocktx.Sysfee = fee.Shift(-1 * models.KSM_DECIMAL).String()
							blocktx.Txid = txid
						} else {
							set_resp(1, "失败类型交易，不推送")
						}

					}
				}
			}
			if flag { //出现这种情况表明区块下没有这个交易ID
				set_resp(1, "the block has not the txid")
				break
			}
		}

		if len(tmpWatchList) > 0 {
			pushBlockTx := new(models.PushAccountBlockInfo)
			pushBlockTx.Type = models.PushTypeAccountTX
			pushBlockTx.Height = height
			pushBlockTx.Hash = hash
			pushBlockTx.CoinName = beego.AppConfig.String("coin")
			//pushBlockTx.Confirmations = beego.AppConfig.DefaultInt64("confirmations", 6)
			pushBlockTx.Confirmations = 13
			pushBlockTx.Time = time
			pushBlockTx.Txs = pushtxs

			//fmt.Println("269:",pushBlockTx)
			//fmt.Println("270 pushBlockTx.Confirmations:",pushBlockTx.Confirmations)
			pusdata, err := json.Marshal(&pushBlockTx)
			if err == nil {
				ksm.AddPushTask(height, hash, tmpWatchList, pusdata)
			}

			has, err := blocktx.Exist(txid)
			if err != nil {
				log.Error(err)
				break
			}
			if !has { //有则更新，无则插入
				num, err := blocktx.Insert()
				if num <= 0 || err != nil {

					log.Error(err)
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
