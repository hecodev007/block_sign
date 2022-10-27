package neo

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/chaincore2/models/neo_model"
	"github.com/group-coldwallet/chaincore2/service/neo"
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
		if neo.UserWatchList[uid] == nil {
			set_resp(1, "user not found")
			break
		}

		//blocktx := dao.NewBlockTX()
		//res, err := blocktx.Select(txid)
		//if !res || err != nil {
		//	set_resp(1, "data not found")
		//	break
		//}
		//
		//blockInfo := dao.NewBlockInfo()
		//err = blockInfo.GetBlockInfoByIndex(blocktx.Height)
		//if err != nil {
		//	set_resp(1, "data not found")
		//	break
		//}

		//查询详情
		txRawData, err := neo.Request("getrawtransaction", []interface{}{txid, 1})
		if err != nil {
			set_resp(1, fmt.Sprintf("txRawData error:%s", err.Error()))
			break
		}
		txData := new(neo_model.GetrawtransactionResp)
		json.Unmarshal(txRawData, txData)
		if txData == nil || txData.Result == nil {
			set_resp(1, fmt.Sprintf("txData error:%s", err.Error()))
			break
		}

		//查询高度
		blockRawData, err := neo.Request("getblock", []interface{}{txData.Result.Blockhash, 1})
		if err != nil {
			set_resp(1, fmt.Sprintf("blockRawData error:%s", err.Error()))
			break
		}
		blockData := new(neo_model.GetblockResp)
		json.Unmarshal(blockRawData, blockData)
		if blockData == nil || blockData.Result.Index == 0 {
			set_resp(1, fmt.Sprintf("blockData error:%s", err.Error()))
			break
		}

		var tmpWatchList map[string]bool = make(map[string]bool)
		//var blockvout_list []*dao.BlockTXVout
		//var blockvin_list []*dao.BlockTXVin
		pushBlockTx := new(models.PushUtxoBlockInfo)
		pushBlockTx.Type = models.PushTypeTX
		pushBlockTx.Height = blockData.Result.Index
		pushBlockTx.Hash = txData.Result.Blockhash
		//pushBlockTx.CoinName = beego.AppConfig.String("coin")
		pushBlockTx.CoinName = "NEO"
		//confirmations := beego.AppConfig.DefaultInt64("confirmations", 12)
		pushBlockTx.Confirmations = txData.Result.Confirmations
		pushBlockTx.Time = txData.Result.Blocktime

		var pushtx models.PushUtxoTx
		pushtx.Txid = txData.Result.Txid
		pushtx.Fee = 0 //主链币暂时不处理gas手续费

		for _, vin := range txData.Result.Vin {
			//查询vin信息，只保留neo信息，gas暂时忽略
			vinRawdata, _ := neo.Request("getrawtransaction", []interface{}{vin.Txid, 1})
			vinData := new(neo_model.GetrawtransactionResp)
			json.Unmarshal(vinRawdata, vinData)
			if vinData == nil || vinData.Result == nil || len(vinData.Result.Vout) < (vin.Vout+1) {
				log.Infof("vin data error")
				continue
			}
			if vinData.Result.Vout[vin.Vout].Asset != neo.NeoAssert {
				log.Infof("暂时忽略非neo资产")
				continue
			}
			tmpWatchList[vinData.Result.Vout[vin.Vout].Address] = true
			pushtx.Vin = append(pushtx.Vin,
				models.PushTxInput{
					Txid:      txData.Result.Txid,
					Vout:      vin.Vout,
					Addresse:  vinData.Result.Vout[vin.Vout].Address,
					Value:     vinData.Result.Vout[vin.Vout].Value.String(),
					AssetId:   &vinData.Result.Vout[vin.Vout].Asset,
					AssetName: &neo.WatchContractList[vinData.Result.Vout[vin.Vout].Asset].Name})
		}

		for _, vout := range txData.Result.Vout {
			//查询vout信息，只保留neo信息，gas暂时忽略
			if vout.Asset != neo.NeoAssert {
				log.Infof("暂时忽略非neo资产")
				continue
			}
			pushtx.Vout = append(
				pushtx.Vout,
				models.PushTxOutput{
					Addresse:  vout.Address,
					Value:     vout.Value.String(),
					N:         vout.N,
					AssetId:   &vout.Asset,
					AssetName: &neo.WatchContractList[vout.Asset].Name},
			)
		}

		pushBlockTx.Txs = append(pushBlockTx.Txs, pushtx)
		pusdata, err := json.Marshal(&pushBlockTx)
		if err == nil {
			neo.AddPushTask(blockData.Result.Index, txData.Result.Txid, tmpWatchList, pusdata)
		}
		break
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
