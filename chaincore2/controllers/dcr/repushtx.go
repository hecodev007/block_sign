package dcr

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/daodcr"
	"github.com/group-coldwallet/chaincore2/models"
	service "github.com/group-coldwallet/chaincore2/service/dcr"
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

		blocktx := dao.NewBlockTX()
		res, err := blocktx.Select(txid)
		if !res || err != nil {
			set_resp(1, "data not found")
			break
		}

		blockInfo := dao.NewBlockInfo()
		err = blockInfo.GetBlockInfoByIndex(blocktx.Height)
		if err != nil {
			set_resp(1, "data not found")
			break
		}

		var tmpWatchList map[string]bool = make(map[string]bool)
		var blockvout_list []*dao.BlockTXVout
		var blockvin_list []*dao.BlockTXVin
		pushBlockTx := new(models.PushUtxoBlockInfo)
		pushBlockTx.Type = models.PushTypeTX
		pushBlockTx.Height = blocktx.Height
		pushBlockTx.Hash = blocktx.Hash
		pushBlockTx.CoinName = beego.AppConfig.String("coin")
		pushBlockTx.Confirmations = blockInfo.Confirmations
		pushBlockTx.Time = blockInfo.Timestamp
		{
			o := orm.NewOrm()
			var maps []orm.Params
			nums, err := o.Raw("select txid, height, hash, vin_txid, vin_voutindex from block_tx_vin where txid = ?", txid).Values(&maps)
			if err == nil && nums > 0 {
				for i := 0; i < len(maps); i++ {
					b := dao.NewBlockTXVin()
					b.Txid = maps[i]["txid"].(string)
					b.Height = common.StrToInt64(maps[i]["height"].(string))
					b.Hash = maps[i]["hash"].(string)
					b.Vintxid = maps[i]["vin_txid"].(string)
					b.VinVoutindex = common.StrToInt64(maps[i]["vin_voutindex"].(string))

					txvout := dao.NewBlockTXVout()
					result, err := txvout.Select(b.Vintxid, b.VinVoutindex)
					if result && err == nil {
						b.Address = txvout.Voutaddress
						b.Amount = txvout.Voutvalue
						tmpWatchList[b.Address] = true
					}

					blockvin_list = append(blockvin_list, b)
				}
			}
		}

		{
			o := orm.NewOrm()
			var maps []orm.Params
			nums, err := o.Raw("select txid, height, hash, vout_n, vout_value, vout_address, invaild, status from block_tx_vout where txid = ?", txid).Values(&maps)
			if err == nil && nums > 0 {
				for i := 0; i < len(maps); i++ {
					b := dao.NewBlockTXVout()
					b.Txid = maps[i]["txid"].(string)
					b.Height = common.StrToInt64(maps[i]["height"].(string))
					b.Hash = maps[i]["hash"].(string)
					b.Voutn = common.StrToInt(maps[i]["vout_n"].(string))
					b.Voutvalue = common.StrToInt64(maps[i]["vout_value"].(string))
					b.Voutaddress = maps[i]["vout_address"].(string)
					b.Invaild = common.StrToInt(maps[i]["invaild"].(string))
					b.Status = common.StrToInt(maps[i]["status"].(string))

					blockvout_list = append(blockvout_list, b)
					tmpWatchList[b.Voutaddress] = true
				}
			}
		}

		var pushtx models.PushUtxoTx
		pushtx.Txid = blocktx.Txid
		pushtx.Fee = blocktx.Sysfee
		pushtx.Coinbase = false
		if blocktx.Coinbase == 1 {
			pushtx.Coinbase = true
		}
		for i := 0; i < len(blockvin_list); i++ {
			// checkout address
			if blockvin_list[i].Address == "" && blockvin_list[i].Amount == 0 {

				for true {
					// 获取原始交易信息
					respdata, err := common.Request("getrawtransaction", []interface{}{blockvin_list[i].Vintxid, 1})
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
					if datas["result"] == nil {
						log.Debug("getrawtransaction not found", blockvin_list[i].Vintxid, "reindex = 1 and txindex = 1 ?")
						break
					}
					tx := datas["result"].(map[string]interface{})
					tmpvouts := tx["vout"].([]interface{})
					if tmpvouts != nil && tmpvouts[blockvin_list[i].VinVoutindex] != nil {
						vout := tmpvouts[blockvin_list[i].VinVoutindex].(map[string]interface{})
						blockvin_list[i].Amount = int64(vout["value"].(float64) * service.ValuePrecision)
						scriptPubKey := vout["scriptPubKey"].(map[string]interface{})
						if scriptPubKey["addresses"] != nil {
							addresses := scriptPubKey["addresses"].([]interface{})
							if len(addresses) == 1 {
								blockvin_list[i].Address = addresses[0].(string)
							}
						}
					}
					break
				}
			}
			value := service.GetValueStr(float64(blockvin_list[i].Amount))
			pushtx.Vin = append(pushtx.Vin, models.PushTxInput{Txid: blockvin_list[i].Vintxid, Vout: int(blockvin_list[i].VinVoutindex), Addresse: blockvin_list[i].Address, Value: value})
		}
		for i := 0; i < len(blockvout_list); i++ {
			value := service.GetValueStr(float64(blockvout_list[i].Voutvalue))
			pushtx.Vout = append(pushtx.Vout, models.PushTxOutput{Addresse: blockvout_list[i].Voutaddress, Value: value, N: blockvout_list[i].Voutn})
		}
		pushBlockTx.Txs = append(pushBlockTx.Txs, pushtx)

		pusdata, err := json.Marshal(&pushBlockTx)
		if err == nil {
			service.AddPushTask(blocktx.Height, blocktx.Txid, tmpWatchList, pusdata)
		}

		break
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
