package neo

//
//import (
//	"encoding/json"
//	"github.com/astaxie/beego"
//	"github.com/astaxie/beego/orm"
//	"github.com/group-coldwallet/chaincore2/common"
//	dao "github.com/group-coldwallet/chaincore2/dao/daoneo"
//	"github.com/group-coldwallet/chaincore2/models"
//	"github.com/group-coldwallet/chaincore2/service/neo"
//	"github.com/group-coldwallet/common/log"
//	"github.com/shopspring/decimal"
//	"strings"
//)
//
//type RepushTxController struct {
//	beego.Controller
//}
//
//func (c *RepushTxController) Post() {
//	// 返回数据
//	resp := map[string]interface{}{
//		"code":    0,
//		"message": "ok",
//		"data":    nil,
//	}
//
//	set_resp := func(code int, message string) {
//		resp["code"] = code
//		resp["message"] = message
//	}
//
//	var jsonObj map[string]interface{}
//	json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)
//	log.Debug(jsonObj)
//
//	for true {
//		if jsonObj["txid"] == nil || jsonObj["uid"] == nil {
//			set_resp(1, "param error")
//			break
//		}
//
//		// 读取 txid
//		txid := jsonObj["txid"].(string)
//		uid := int64(jsonObj["uid"].(float64))
//		if neo.UserWatchList[uid] == nil {
//			set_resp(1, "user not found")
//			break
//		}
//
//		blocktx := dao.NewBlockTX()
//		res, err := blocktx.Select(txid)
//		if !res || err != nil {
//			set_resp(1, "data not found")
//			break
//		}
//
//		blockInfo := dao.NewBlockInfo()
//		err = blockInfo.GetBlockInfoByIndex(blocktx.Height)
//		if err != nil {
//			set_resp(1, "data not found")
//			break
//		}
//
//		var tmpWatchList map[string]bool = make(map[string]bool)
//		var blockvout_list []*dao.BlockTXVout
//		var blockvin_list []*dao.BlockTXVin
//		pushBlockTx := new(models.PushUtxoBlockInfo)
//		pushBlockTx.Type = models.PushTypeTX
//		pushBlockTx.Height = blocktx.Height
//		pushBlockTx.Hash = blocktx.Hash
//		pushBlockTx.CoinName = beego.AppConfig.String("coin")
//
//		confirmations := beego.AppConfig.DefaultInt64("confirmations", 12)
//		//pushBlockTx.Confirmations = blockInfo.Confirmations
//		pushBlockTx.Confirmations = confirmations
//		pushBlockTx.Time = blockInfo.Timestamp
//		{
//			o := orm.NewOrm()
//			var maps []orm.Params
//			nums, err := o.Raw("select txid, height, hash, vin_txid, vin_voutindex from block_tx_vin where txid = ?", txid).Values(&maps)
//			if err == nil && nums > 0 {
//				for i := 0; i < len(maps); i++ {
//					b := dao.NewBlockTXVin()
//					b.Txid = maps[i]["txid"].(string)
//					b.Height = common.StrToInt64(maps[i]["height"].(string))
//					b.Hash = maps[i]["hash"].(string)
//					b.Vintxid = maps[i]["vin_txid"].(string)
//					b.VinVoutindex = common.StrToInt(maps[i]["vin_voutindex"].(string))
//
//					txvout := dao.NewBlockTXVout()
//					result, err := txvout.Select(b.Vintxid, b.VinVoutindex)
//					if result && err == nil {
//						tmpWatchList[txvout.Voutaddress] = true
//						b.Address = txvout.Voutaddress
//						b.Amount = txvout.Voutvalue
//					}
//
//					blockvin_list = append(blockvin_list, b)
//				}
//			}
//		}
//
//		{
//			o := orm.NewOrm()
//			var maps []orm.Params
//			nums, err := o.Raw("select txid, height, hash, vout_n, vout_value, vout_address, invaild, status, asset_name, asset_selltxid, asset_id, asset_value from block_tx_vout where txid = ?", txid).Values(&maps)
//			if err == nil && nums > 0 {
//				for i := 0; i < len(maps); i++ {
//					b := dao.NewBlockTXVout()
//					b.Txid = maps[i]["txid"].(string)
//					b.Height = common.StrToInt64(maps[i]["height"].(string))
//					b.Hash = maps[i]["hash"].(string)
//					b.Voutn = common.StrToInt(maps[i]["vout_n"].(string))
//					b.Voutvalue = maps[i]["vout_value"].(string)
//					b.Voutaddress = maps[i]["vout_address"].(string)
//					b.Invaild = common.StrToInt(maps[i]["invaild"].(string))
//					b.Status = common.StrToInt(maps[i]["status"].(string))
//					b.AssetName = maps[i]["asset_name"].(string)
//					b.AssetId = maps[i]["asset_id"].(string)
//
//					blockvout_list = append(blockvout_list, b)
//					tmpWatchList[b.Voutaddress] = true
//				}
//			}
//		}
//
//		var pushtx models.PushUtxoTx
//		pushtx.Txid = blocktx.Txid
//		pushtx.Fee = blocktx.Sysfee
//		for i := 0; i < len(blockvin_list); i++ {
//			// checkout address
//			if blockvin_list[i].Address != "" {
//				for true {
//					// 解析原始交易信息
//					respdata, err := neo.Request("getrawtransaction", []interface{}{blockvin_list[i].Vintxid, 1})
//					if err != nil {
//						beego.Error(err)
//						break
//					} else {
//						log.Infof("repush getrawtransaction:%s", string(respdata))
//					}
//
//					var datas map[string]interface{}
//					err = json.Unmarshal(respdata, &datas)
//					if err != nil {
//						log.Debug(err)
//						break
//					}
//					if datas["result"] == nil {
//						log.Debug("getrawtransaction not found", blockvin_list[i].Vintxid, "reindex = 1 and txindex = 1 ?")
//						break
//					}
//					tx := datas["result"].(map[string]interface{})
//					tmpvouts := tx["vout"].([]interface{})
//					if tmpvouts != nil && tmpvouts[blockvin_list[i].VinVoutindex] != nil {
//						vout := tmpvouts[blockvin_list[i].VinVoutindex].(map[string]interface{})
//						blockvin_list[i].Address = vout["address"].(string)
//						if neo.WatchContractList[vout["asset"].(string)] != nil {
//							blockvin_list[i].AssetName = neo.WatchContractList[vout["asset"].(string)].Name
//						}
//						blockvin_list[i].Amount = vout["value"].(string)
//						blockvin_list[i].AssetId = vout["asset"].(string)
//						tmpWatchList[blockvin_list[i].Address] = true
//					}
//					break
//				}
//			}
//
//			if neo.WatchContractList[blockvin_list[i].AssetId] == nil {
//				continue
//			}
//			var assets *models.AssetsInfo = nil
//			if blockvin_list[i].AssetId != "" && blockvin_list[i].AssetName != "" {
//				// token
//				assets = new(models.AssetsInfo)
//				assets.AssetId = blockvin_list[i].AssetId
//				assets.Name = blockvin_list[i].AssetName
//			}
//			amount, _ := decimal.NewFromString(blockvin_list[i].Amount)
//			_amount := amount.Div(decimal.New(1, int32(neo.WatchContractList[blockvin_list[i].AssetId].Decimal))).String()
//			log.Info("vin 添加")
//			pushtx.Vin = append(pushtx.Vin, models.PushTxInput{Txid: blockvin_list[i].Vintxid, Vout: blockvin_list[i].VinVoutindex, Addresse: blockvin_list[i].Address, Value: _amount, AssetId: &assets.AssetId, AssetName: &assets.Name})
//		}
//		for i := 0; i < len(blockvout_list); i++ {
//			if neo.WatchContractList[blockvout_list[i].AssetId] == nil {
//				continue
//			}
//			var assets *models.AssetsInfo = nil
//			if blockvout_list[i].AssetId != "" && blockvout_list[i].AssetName != "" {
//				// token
//				assets = new(models.AssetsInfo)
//				assets.AssetId = blockvout_list[i].AssetId
//				assets.Name = blockvout_list[i].AssetName
//			}
//			amount := decimal.Zero
//			if strings.ToLower(assets.Name) == "neo" {
//				amount, _ = decimal.NewFromString(blockvout_list[i].Voutvalue)
//			} else {
//				amount = decimal.NewFromFloat(float64(blockvout_list[i].AssetValue))
//			}
//			_amount := amount.Div(decimal.New(1, int32(neo.WatchContractList[blockvout_list[i].AssetId].Decimal))).String()
//			pushtx.Vout = append(pushtx.Vout, models.PushTxOutput{Addresse: blockvout_list[i].Voutaddress, Value: _amount, N: blockvout_list[i].Voutn, AssetId: &assets.AssetId, AssetName: &assets.Name})
//		}
//		pushBlockTx.Txs = append(pushBlockTx.Txs, pushtx)
//		pusdata, err := json.Marshal(&pushBlockTx)
//		if err == nil {
//			neo.AddPushTask(blocktx.Height, blocktx.Txid, tmpWatchList, pusdata)
//		}
//		break
//	}
//
//	c.Data["json"] = resp
//	c.ServeJSON()
//}
