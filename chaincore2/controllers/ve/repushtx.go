package ve

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
	"github.com/group-coldwallet/chaincore2/common"
	dao "github.com/group-coldwallet/chaincore2/dao/dao_ve"
	"github.com/group-coldwallet/chaincore2/models"
	service "github.com/group-coldwallet/chaincore2/service/ve"
	"github.com/group-coldwallet/common/log"
	"github.com/shopspring/decimal"
	"github.com/vechain/thor/thor"
	//"github.com/vechain/thor/thor"
)

type RepushTxController struct {
	beego.Controller
}

func (c *RepushTxController) Post() {
	var (
		height         = int64(0) //区块高度
		hash           = ""       //区块hash
		blockTimestamp = int64(0)
		pushBlockTx    = new(models.PushAccountBlockInfo)
	)
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

		var pushtxs []models.PushAccountTx
		// 从节点拿数据推送
		var tmpWatchList map[string]bool = make(map[string]bool)
		{
			client := service.NewRskBlock(beego.AppConfig.String("nodeurl"))
			transaction, err := client.GetTransactionsByID(txid)
			if err != nil {
				set_resp(1, err.Error())
				break
			}
			if transaction.Reverted { //交易失败不做处理
				set_resp(1, "this transaction is not a success trans") //返回错误信息，这不是一笔正常的交易
				break
			}
			for i := 0; i < len(transaction.Outputs); i++ {
				for _, event := range transaction.Outputs[i].Events {
					blocktx := dao.NewBlockTX()
					blocktx.Height = transaction.Meta.BlockNumber
					blocktx.Hash = transaction.Meta.BlockID
					blocktx.Txid = txid
					has, contactInfo := service.HasContact(event.Address)
					if !has || len(event.Topics) <= 0 {
						continue
					}
					paid, _ := common.StrBaseToBigInt(transaction.Paid, 16)
					paid_ := decimal.NewFromBigInt(paid, -int32(contactInfo.Decimal))
					blocktx.Sysfee = paid_.String() //金额，18位小数
					blocktx.FeeName = contactInfo.Name
					blocktx.GasUsed = transaction.GasUsed
					blocktx.ContractAddress = event.Address
					addrs := make([]string, 0, 3)
					for _, topic := range event.Topics {
						if len(topic) > 0 {
							a, _ := thor.ParseBytes32(topic)
							c := thor.BytesToAddress(a.Bytes())
							addrs = append(addrs, c.String())
						}
					}

					blocktx.From = transaction.Meta.TxOrigin
					//下面这段代码是通过请求另一个获取交易data数据来比较，找出to地址
					toAddr, gas, gasPrice, err := client.GetTransactionsByPending(txid, addrs)
					if err != nil {
						log.Debug(err)
						blocktx.To = addrs[len(addrs)-1]
					} else {
						blocktx.To = toAddr
					}

					blocktx.Gas = gas
					blocktx.GasPrice = gasPrice
					if transaction.Meta.TxOrigin != transaction.GasPayer {
						blocktx.GasPayer = transaction.GasPayer
					}

					ammount, _ := common.StrBaseToBigInt(event.Data, 16)
					d := decimal.NewFromBigInt(ammount, -int32(contactInfo.Decimal))
					blocktx.Amount = d.String() //获得金额

					//var tmpWatchList map[string]bool = make(map[string]bool)
					log.Debug(blocktx.From, blocktx.To)
					if service.WatchAddressList[blocktx.From] != nil {
						log.Debug("watchaddr", blocktx.From)
						tmpWatchList[blocktx.From] = true
					}
					if service.WatchAddressList[blocktx.To] != nil {
						log.Debug("watchaddr", blocktx.To)
						tmpWatchList[blocktx.To] = true
					}
					if len(tmpWatchList) > 0 {
						var b models.PushAccountTx
						b.Txid = blocktx.Txid
						b.Fee = common.StrToFloat64(blocktx.Sysfee)
						b.From = blocktx.From
						b.To = blocktx.To
						b.Amount = blocktx.Amount
						b.Memo = blocktx.Memo
						b.Contract = blocktx.ContractAddress
						if blocktx.GasPayer != "" {
							b.FeePayer = blocktx.GasPayer
						}
						pushtxs = append(pushtxs, b)

						if exist, _ := blocktx.Exist(txid, blocktx.From, blocktx.To); !exist {
							num, err := blocktx.Insert()
							if num <= 0 || err != nil {
								log.Error(err)
							}
						}
					}
				}
				for _, transfer := range transaction.Outputs[i].Transfers {
					blocktx := dao.NewBlockTX()
					blocktx.Height = transaction.Meta.BlockNumber
					blocktx.Hash = transaction.Meta.BlockID
					blocktx.Txid = txid

					if transaction.Meta.BlockNumber > 0 {
						height = transaction.Meta.BlockNumber
					}
					if transaction.Meta.BlockID != "" {
						hash = transaction.Meta.BlockID
					}
					if transaction.Meta.BlockTimestamp > 0 {
						blockTimestamp = transaction.Meta.BlockTimestamp
					}

					ammount, _ := common.StrBaseToBigInt(transfer.Amount, 16)
					d := decimal.NewFromBigInt(ammount, -18)
					blocktx.Amount = d.String() //金额，18位小数

					paid, _ := common.StrBaseToBigInt(transaction.Paid, 16)
					paid_ := decimal.NewFromBigInt(paid, -18)
					blocktx.Sysfee = paid_.String() //金额，18位小数
					blocktx.FeeName = beego.AppConfig.String("fee_coin")

					blocktx.GasUsed = transaction.GasUsed
					if transaction.GasPayer != transfer.Sender {
						blocktx.GasPayer = transaction.GasPayer
					}
					blocktx.From = transfer.Sender  //发送者
					blocktx.To = transfer.Recipient //接收者

					//var tmpWatchList map[string]bool = make(map[string]bool)
					if service.WatchAddressList[blocktx.From] != nil {
						log.Debug("watchaddr", blocktx.From)
						tmpWatchList[blocktx.From] = true
					}

					if service.WatchAddressList[blocktx.To] != nil {
						log.Debug("watchaddr", blocktx.To)
						tmpWatchList[blocktx.To] = true
					}

					if len(tmpWatchList) > 0 {
						var b models.PushAccountTx
						b.Txid = blocktx.Txid
						b.Fee = common.StrToFloat64(blocktx.Sysfee)
						b.From = blocktx.From
						b.To = blocktx.To
						b.Amount = blocktx.Amount
						b.Memo = blocktx.Memo
						b.Contract = blocktx.ContractAddress
						if blocktx.GasPayer != "" {
							b.FeePayer = blocktx.GasPayer
						}
						pushtxs = append(pushtxs, b)

						if exist, _ := blocktx.Exist(txid, blocktx.From, blocktx.To); !exist {
							num, err := blocktx.Insert()
							if num <= 0 || err != nil {
								log.Error(err)
							}
						}
					}
				}
			}
		}

		if len(pushtxs) <= 0 {
			set_resp(1, "there is no data to push")
			break
		}

		pushBlockTx.Type = models.PushTypeAccountTX
		pushBlockTx.CoinName = beego.AppConfig.String("coin")
		pushBlockTx.Confirmations = beego.AppConfig.DefaultInt64("confirmations", 6) + 1
		pushBlockTx.Time = blockTimestamp

		pushBlockTx.Type = models.PushTypeAccountTX
		pushBlockTx.Height = height
		pushBlockTx.Hash = hash
		pushBlockTx.Txs = pushtxs
		pusdata, err := json.Marshal(&pushBlockTx)
		//beego.Debug("==============================================+++++++++++>")
		//beego.Debug(string(pusdata))
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
		if jsonObj["txid"] == nil || jsonObj["uid"] == nil || jsonObj["height"] == nil {
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

			transaction, err2 := client.GetTransaction(context.TODO(), txid)
			if err2 != nil {
				log.Error(err)
				break
			}
			if transaction != nil {
				var b models.PushAccountTx
				b.Txid = txid

				data, err := base64.RawURLEncoding.DecodeString(transaction.Owner()) //获取From字段
				h := sha256.New()
				h.Write(data)
				b.From = utils.EncodeToBase64(h.Sum(nil))
				b.To = transaction.Target() //转到哪而去

				fee, err := decimal.NewFromString(transaction.Reward()) //获取手续费
				if err != nil {
					log.Error(err)
					break
				}
				b.Fee, _ = fee.Div(decimal.New(1, 12)).Float64()

				//获取交易金额
				quantity, err := decimal.NewFromString(transaction.Quantity())
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
				blockInfo, err := client2.GetBlockByHeight(context.TODO(), height)
				if err != nil {
					log.Error(err)
					break
				}
				hash = blockInfo.IndepHash
				var hasTxt = false
				for _, v := range blockInfo.Txs { //这段代码是遍历区块的交易ID查询传入的交易ID是否在区块交易ID列表中
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

				if len(tmpWatchList) > 0 {
					has, err := blocktx.Exist(txid, blocktx.From, blocktx.To)
					if err != nil {
						beego.Error(err)
						break
					}
					if !has { //有则更新，无则插入
						num, err := blocktx.Insert()
						if num <= 0 || err != nil {
							beego.Error(err)
						}
					}
				}
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
			resp["code"] = 0
			resp["message"] = "ok"
		}

		break
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
