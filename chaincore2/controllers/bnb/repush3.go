package bnb

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/binance-chain/go-sdk/types/tx"
	"github.com/group-coldwallet/chaincore2/common"
	"github.com/group-coldwallet/chaincore2/models"
	"github.com/group-coldwallet/chaincore2/service/bnb"
	service "github.com/group-coldwallet/chaincore2/service/bnb"
	"github.com/group-coldwallet/common/log"
	"time"
)

/*
write by jun
2020/6/23
*/

type RePushController3 struct {
	beego.Controller
}

func (c *RePushController3) Post() {

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

	var req ReqParams
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &req)
	if err != nil {
		set_resp(1, fmt.Sprintf("parse req params error.err=[%v]", err))
		c.Data["json"] = resp
		c.ServeJSON()
		return
	}
	log.Debug("接受推送")
	log.Debug(string(c.Ctx.Input.RequestBody))
	for true {
		if req.Height == 0 || req.Uid == 0 || req.Txid == "" {
			set_resp(1, "params have null")
			break
		}
		if bnb.UserWatchList[req.Uid] == nil {
			set_resp(1, "user not found")
			break
		}
		//nodeUrl := beego.AppConfig.DefaultString("nodeUrl", "bnb.rylink.io:30147")
		//if strings.HasPrefix(nodeUrl, "http://") {
		//	nodeUrl = strings.TrimPrefix(nodeUrl, "http://")
		//}
		nodeUrl := "dataseed1.binance.org"
		blockUrl := fmt.Sprintf("http://%s/block?height=%d", nodeUrl, req.Height)
		blockReq := httplib.Get(blockUrl)
		blockRespData, err := blockReq.Bytes()
		if err != nil {
			set_resp(1, fmt.Sprintf("get block error,Err=[%v]", err))
			break
		}
		//fmt.Println(string(blockRespData))
		log.Debug(string(blockRespData))
		var block map[string]interface{}
		err2 := json.Unmarshal(blockRespData, &block)
		if err2 != nil {
			set_resp(1, fmt.Sprintf("parse block error,err=[%v]", err2))
			break
		}
		result := block["result"].(map[string]interface{})
		block_meta := result["block_meta"].(map[string]interface{})
		block_id := block_meta["block_id"].(map[string]interface{})
		header := block_meta["header"].(map[string]interface{})
		blocktime := header["time"].(string)
		blockHash := block_id["hash"].(string)

		url := fmt.Sprintf("http://%s/tx_search?query=\"tx.height=%d\"&per_page=100", nodeUrl, req.Height)
		reqs := httplib.Get(url)
		respdata, err := reqs.Bytes()
		if err != nil {
			log.Debug(err)
			break
		}
		var respD RespData
		err = json.Unmarshal(respdata, &respD)
		if err != nil {
			set_resp(1, fmt.Sprintf("parse resp data error,err=[%v]", err))
			break
		}
		if len(respD.Result.Txs) == 0 {
			set_resp(1, fmt.Sprintf("height=[%d] oo not find any tx", req.Height))
			break
		}
		for _, t := range respD.Result.Txs {
			if t.Hash == req.Txid {
				//处理当前这笔txid
				_tx := t.Tx
				b, _ := base64.StdEncoding.DecodeString(_tx)
				var stdtx tx.StdTx
				err = tx.Cdc.UnmarshalBinaryLengthPrefixed(b, &stdtx)
				if err != nil {
					set_resp(1, fmt.Sprintf("parse tx error,Err=[%v]", err))
					break
				}
				if len(stdtx.Msgs) == 0 {
					set_resp(1, "stdtx.Msgs is null")
					break
				}
				if stdtx.Msgs[0].Type() != "send" {
					set_resp(1, "invalid transaction")
					break
				}

				d1, err := tx.Cdc.MarshalJSON(stdtx.Msgs[0])
				if err != nil {
					log.Debug(err)
					set_resp(1, "parse msg error")
					break
				}
				//log.Debug(string(d1))

				var sendmsg SendMsgTx
				err = json.Unmarshal(d1, &sendmsg)
				if err != nil {
					log.Debug(err)
					set_resp(1, "parse send msg error")
					break
				}

				//log.Debug(stdtx, stdtx.Msgs[0].Type(), stdtx.Msgs[0].Route())
				//log.Debug(sendmsg)

				// check coin
				denom := sendmsg.Msg.Inputs[0].Coins[0].Denom
				if service.WatchContractList[denom] == nil {
					set_resp(1, fmt.Sprintf("do not find this contractList: %s", denom))
					break
				}
				var pushtxs []models.PushAccountTx
				var tmpWatchList map[string]bool = make(map[string]bool)
				{
					from := sendmsg.Msg.Inputs[0].Address
					to := sendmsg.Msg.Outputs[0].Address
					amount := common.StrToFloat64(sendmsg.Msg.Inputs[0].Coins[0].Amount)
					memo := service.GetMemo(hex.EncodeToString(b))
					if service.WatchAddressList[from] != nil {
						tmpWatchList[from] = true
					}
					if service.WatchAddressList[to] != nil {
						tmpWatchList[to] = true
					}

					if len(tmpWatchList) > 0 {
						var b models.PushAccountTx
						b.Txid = req.Txid
						b.Fee = 0.000375
						b.From = from
						b.To = to
						b.Amount = bnb.GetValueStr(amount)
						b.Memo = memo
						b.Contract = denom
						pushtxs = append(pushtxs, b)
					}
				}
				pushBlockTx := new(models.PushAccountBlockInfo)
				pushBlockTx.Type = models.PushTypeAccountTX
				pushBlockTx.Height = req.Height
				pushBlockTx.Hash = blockHash
				pushBlockTx.CoinName = beego.AppConfig.String("coin")
				pushBlockTx.Confirmations = beego.AppConfig.DefaultInt64("confirmations", 12)
				_time, _ := time.Parse(time.RFC3339Nano, blocktime)
				pushBlockTx.Time = _time.Unix()
				pushBlockTx.Txs = pushtxs
				pusdata, err := json.Marshal(&pushBlockTx)
				if err == nil {
					log.Debug("------------repush3----------------->")
					log.Debug(string(pusdata))
					bnb.AddPushTask(req.Height, blockHash, tmpWatchList, pusdata)
				}
			}
		}
		break
	}
	c.Data["json"] = resp
	c.ServeJSON()
}
