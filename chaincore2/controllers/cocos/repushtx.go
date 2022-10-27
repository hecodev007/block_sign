package cocos

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/chaincore2/common"
	"github.com/group-coldwallet/chaincore2/models"
	service "github.com/group-coldwallet/chaincore2/service/cocos"
	"github.com/group-coldwallet/common/log"
	"github.com/shopspring/decimal"
	"time"
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
	log.Infof("重推交易： %v", jsonObj)

	for true {
		if jsonObj["txid"] == nil || jsonObj["uid"] == nil {
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
		var pushtxs []models.PushAccountTx
		var tmpWatchList map[string]bool = make(map[string]bool)
		{
			//o := orm.NewOrm()
			//var maps []orm.Params
			//nums, err := o.Raw("select txid, height, hash, sys_fee, fromaccount, toaccount,amount,memo from block_tx where txid = ?", txid).Values(&maps)
			//if err == nil && nums > 0 {
			//	for i := 0; i < len(maps); i++ {
			//		height = common.StrToInt64(maps[i]["height"].(string))
			//		hash = maps[i]["hash"].(string)
			//
			//		var b models.PushAccountTx
			//		b.Txid = maps[i]["txid"].(string)
			//		b.Fee = common.StrToFloat64(maps[i]["sys_fee"].(string))
			//		b.From = maps[i]["fromaccount"].(string)
			//		b.To = maps[i]["toaccount"].(string)
			//		b.Memo = maps[i]["memo"].(string)
			//		b.Amount = maps[i]["amount"].(string)
			//
			//		pushtxs = append(pushtxs, b)
			//		_from := service.GetAccountIdByAccount(b.From)
			//		if _from != "" {
			//			tmpWatchList[_from] = true
			//		}
			//		_to := service.GetAccountIdByAccount(b.To)
			//		if _to != "" {
			//			tmpWatchList[_to] = true
			//		}
			//	}
			//}
			// write by flynn 使用接口去查询链上交易
			txData, err := getTxOnChain(txid)
			if err != nil {
				set_resp(1, fmt.Sprintf("get tx data error,%v", err))
				break
			}
			var b models.PushAccountTx
			b.From = txData["from"].(string)
			b.To = txData["to"].(string)
			b.Memo = txData["memo"].(string)
			b.Amount = txData["amount"].(string)
			b.Fee = txData["fee"].(float64)
			b.Txid = txid
			if txData["asset_id"].(string) != "1.3.0" {
				b.Contract = txData["asset_id"].(string)
			}
			if b.From != "" && service.WatchAddressList[txData["from_id"].(string)] != nil {
				tmpWatchList[txData["from_id"].(string)] = true
			}
			if b.To != "" && service.WatchAddressList[txData["to_id"].(string)] != nil {
				tmpWatchList[txData["to_id"].(string)] = true
			}
			pushtxs = append(pushtxs, b)
		}

		//blockInfo := dao.NewBlockInfo()
		//err := blockInfo.GetBlockInfoByIndex(height)
		//if err != nil {
		//	set_resp(1, "data not found")
		//	break
		//}

		blockInfo, err := getChainInfo(height)
		if err == nil {
			pushBlockTx := new(models.PushAccountBlockInfo)
			pushBlockTx.Type = models.PushTypeAccountTX
			pushBlockTx.Height = height
			pushBlockTx.Hash = blockInfo["hash"].(string)
			pushBlockTx.CoinName = beego.AppConfig.String("coin")
			pushBlockTx.Confirmations = beego.AppConfig.DefaultInt64("confirmations", 6) + 1
			timestamp := blockInfo["timestamp"].(string)
			t, _ := time.Parse("2006-01-02T15:04:05", timestamp)
			pushBlockTx.Time = t.Unix()
			pushBlockTx.Txs = pushtxs
			pusdata, err := json.Marshal(&pushBlockTx)
			if err == nil {
				log.Infof("添加重推数据： %s", string(pusdata))
				service.AddPushTask(height, txid, tmpWatchList, pusdata)
			}
			break
		} else {
			set_resp(1, fmt.Sprintf("get chain info error,%v", err))
			break
		}
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

func getTxOnChain(txid string) (map[string]interface{}, error) {
	respdata, err := common.Request("get_transaction_by_id", []interface{}{txid})
	if err != nil {
		return nil, fmt.Errorf("rpc get txid error,%v", err)
	}
	var datas map[string]interface{}
	err = json.Unmarshal(respdata, &datas)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal tx error,%v", err)
	}
	if datas["result"] == nil {
		return nil, fmt.Errorf("get_transaction_by_id not found %s reindex = 1 and txindex = 1 ?", txid)
	}
	tx := datas["result"].(map[string]interface{})
	if tx == nil {
		return nil, fmt.Errorf("tx is null")
	}
	operations := tx["operations"].([]interface{})
	_operations := operations[0].([]interface{})
	if int(_operations[0].(float64)) != 0 {
		return nil, fmt.Errorf("operations is not equal 0,operations=%f", _operations[0].(float64))
	}
	txinfo := _operations[1].(map[string]interface{})
	from_id := txinfo["from"].(string)
	to_id := txinfo["to"].(string)
	amountobj := txinfo["amount"].(map[string]interface{})
	asset_id := amountobj["asset_id"].(string)
	if service.WatchContractList[asset_id] == nil {
		return nil, fmt.Errorf("do not contain this asset_id = %s", asset_id)
	}
	var decimalAmount decimal.Decimal
	switch amountobj["amount"].(type) {
	case string:
		decimalAmount, _ = decimal.NewFromString(amountobj["amount"].(string))
	case float64:
		decimalAmount = decimal.NewFromFloat(amountobj["amount"].(float64))
	}

	// 手续费
	var feeAmount decimal.Decimal
	if tx["operation_results"] != nil {
		operation_results := tx["operation_results"].([]interface{})
		for _, v := range operation_results {
			vv := v.([]interface{})
			if len(vv) >= 2 {
				tmp := vv[1].(map[string]interface{})
				if tmp["fees"] != nil {
					tmp2 := tmp["fees"].([]interface{})
					if len(tmp2) > 0 {
						fees := tmp2[0].(map[string]interface{})
						if fees["asset_id"] == service.CocosAssetId {
							feeAmount = decimal.NewFromFloat(fees["amount"].(float64))
							feeAmount = feeAmount.Div(service.PrecisionDecimal)
						}
					}
				}
			}
		}
	}
	coin_set := service.WatchContractList[asset_id]
	_amount := decimalAmount.Shift(-int32(coin_set.Decimal)).String()
	var (
		from, to, memo string
		innerAccount   bool
	)
	if service.AccountMap[from_id] != "" {
		from = service.AccountMap[from_id]
	} else {
		from = service.GetAccountById(from_id)
	}
	if service.AccountMap[to_id] != "" {
		to = service.AccountMap[to_id]
	} else {
		to = service.GetAccountById(to_id)
	}
	if from != "" || to != "" {
		innerAccount = true
	}
	if txinfo["memo"] != nil {
		memoinfo := txinfo["memo"].([]interface{})
		if int(memoinfo[0].(float64)) == 0 {
			memo = memoinfo[1].(string)
		} else if innerAccount {
			memo = service.GetRawMemo(memoinfo[1].(map[string]interface{}))
		}
	}
	respData := make(map[string]interface{})
	respData["amount"] = _amount
	respData["fee"], _ = feeAmount.Float64()
	respData["from"] = from
	respData["to"] = to
	respData["from_id"] = from_id
	respData["to_id"] = to_id
	respData["memo"] = memo
	respData["asset_id"] = asset_id
	return respData, nil
}

func getChainInfo(height int64) (map[string]interface{}, error) {
	respdata, err := common.RequestStr("get_block", []interface{}{height})
	if err != nil {
		return nil, fmt.Errorf("rpc get chain info error,height=%d,err=%v", height, err)
	}
	var datas map[string]interface{}
	err = json.Unmarshal([]byte(respdata), &datas)
	if err != nil {

		return nil, fmt.Errorf("json unmarshal block data error,%v", err)
	}

	if datas["result"] == nil {
		return nil, fmt.Errorf("block data is null ,height=%d", height)
	}

	// 区块详情
	result := datas["result"].(map[string]interface{})
	if result == nil {
		return nil, fmt.Errorf("block result is null ,height=%d", height)
	}
	hash := result["block_id"].(string)
	respData := make(map[string]interface{})
	respData["hash"] = hash
	respData["timestamp"] = result["timestamp"].(string)
	return respData, nil
}
