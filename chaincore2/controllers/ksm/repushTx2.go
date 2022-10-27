package ksm

import (
	"encoding/json"
	"fmt"
	"github.com/JFJun/substrate-go/rpc"
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/chaincore2/common"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/chaincore2/service/ksm"
	"github.com/group-coldwallet/common/log"
	"github.com/shopspring/decimal"
	"math/big"
)

type KsmRepushTx2Controller struct {
	beego.Controller
	client *rpc.Client
}

func (c *KsmRepushTx2Controller) Post() {
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
	errJ := json.Unmarshal(c.Ctx.Input.RequestBody, &jsonObj)
	if errJ != nil {
		log.Error(errJ)
		set_resp(1, fmt.Sprintf("%v", errJ))
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}
	log.Debug(jsonObj)
	for true {
		if jsonObj["txid"] == nil || jsonObj["uid"] == nil || jsonObj["height"] == nil {
			set_resp(1, "param error")
			break
		}
		// 读取 txid
		txid := jsonObj["txid"].(string)
		height := int64(jsonObj["height"].(float64))
		uid := int64(jsonObj["uid"].(float64))

		if ksm.UserWatchList[uid] == nil {
			set_resp(1, "user not found")
			break
		}
		var hash string = ""
		var pushtxs []models.PushAccountTx
		var timestamp int64
		var err error

		if c.client == nil {
			c.client, err = rpc.New(beego.AppConfig.String("nodeurl"), "", "")
			if err != nil {
				log.Errorf("repush tx init client error,Err=[%v]", err)
				set_resp(1, fmt.Sprintf("repush tx init client error,Err=[%v]", err))
				break
			}
		}

		var tmpWatchList map[string]bool = make(map[string]bool)
		{
			block, err2 := c.client.GetBlockByNumber(height)
			if err2 != nil {
				log.Errorf("repush tx get block error,Err=[%v]", err)
				continue
			}
			height = block.Height
			hash = block.BlockHash
			timestamp = block.Timestamp
			if len(block.Extrinsic) > 0 {
				for _, extrinsic := range block.Extrinsic {
					if extrinsic.Txid == txid {
						if extrinsic.Status == "success" && extrinsic.Type == "transfer" {
							from, to := extrinsic.FromAddress, extrinsic.ToAddress

							aDec, _ := decimal.NewFromString(extrinsic.Amount)
							amount := aDec.Shift(-1 * models.KSM_DECIMAL).String()
							var b models.PushAccountTx
							feeBI, _ := new(big.Int).SetString(extrinsic.Fee, 10)

							fee := common.CoinToFloat(feeBI, int32(models.KSM_DECIMAL))
							b.Txid = txid
							b.Fee = fee
							b.From = from
							b.To = to
							b.Amount = amount
							b.Memo = ""
							b.Contract = ""
							pushtxs = append(pushtxs, b)
							if ksm.WatchAddressList[to] != nil {
								tmpWatchList[to] = true
							}
							if ksm.WatchAddressList[from] != nil {
								tmpWatchList[from] = true
							}
						}
					}
				}
			}
		}
		if len(tmpWatchList) == 0 {
			set_resp(1, "do not find any ours address")
			break
		}
		pushBlockTx := new(models.PushAccountBlockInfo)
		pushBlockTx.Type = models.PushTypeAccountTX
		pushBlockTx.Height = height
		pushBlockTx.Hash = hash
		pushBlockTx.CoinName = beego.AppConfig.String("coin")
		pushBlockTx.Confirmations, _ = beego.AppConfig.Int64("confirmations")
		pushBlockTx.Time = timestamp
		pushBlockTx.Txs = pushtxs
		pusdata, err := json.Marshal(&pushBlockTx)
		log.Debug("Push Data", string(pusdata))
		if err == nil {
			ksm.AddPushTask(height, txid, tmpWatchList, pusdata)
		}
		break
	}
	c.Data["json"] = resp
	c.ServeJSON()
}
