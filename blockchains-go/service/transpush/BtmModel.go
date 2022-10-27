package transpush

import (
	"encoding/json"
	"errors"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

// utxo模型
type BtmModel struct {
	// nothing
}

func (r *BtmModel) Run(reqexit <-chan bool) {
	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		log.Error(err)
		WaitGroupTransPush.Done()
		return
	}
	defer redisHelper.Close()

	log.Debug("Run BtmModel")
	run := true
	for run {
		select {
		case s := <-reqexit:
			log.Error("BtmModel exit", s)
			run = false
			break

		default:
			item, err := redisHelper.Rpop("btm_push_list_new")
			if err != nil {
				if strings.Contains(err.Error(), "nil") {
					time.Sleep(time.Second * 1)
					break
				}
				log.Error(err)
				time.Sleep(time.Second * 1)
				break
			}
			if item != nil {
				if err := dispossBtm(item, redisHelper); err != nil {
					redisHelper.LeftPush("btm_push_list_new", string(item))
				}
			}
		}
	}
	WaitGroupTransPush.Done()
}

func dispossBtm(postdata []byte, redisHelper *util.RedisClient) error {
	var (
		transData model.PushBtmBlockInfo
		coinName  string //主链币币种

	)
	//isMerge from都为用户地址，不能掺杂其他地址，否则不推送，人工处理
	//isFee 不关心from,只要to地址是手续费地址
	err := json.Unmarshal(postdata, &transData)
	if err != nil {
		log.Debug(err)
		return err
	}
	log.Info("btm 数据处理")

	//信息存储结构
	push_data := []map[string]interface{}{}
	coinName = transData.CoinName
	if len(transData.Txs) == 0 {
		return errors.New("param error")
	}
	for _, tx := range transData.Txs {
		var (
			isFeeTx           bool //是否是手续费交易
			isCollectOrChange bool //是否是归集交易或者找零
			isTransfer        bool //是否出账
			isPushFrom        bool //是否已经推送过给发送方
			item_dump         map[string]interface{}
			data              map[string]interface{}
			fromMchId         int
		)

		isFail := tx.StatusFail
		if isFail {
			//失败的交易 直接忽略
			continue
		}
		txid := tx.Txid
		feeFloat := tx.FeeFloat
		muxId := tx.MuxId
		isCoinbase := tx.Coinbase
		item_dump = make(map[string]interface{})
		item_dump["type"] = transData.Type  //30交易 31确认数
		data = make(map[string]interface{}) //用于 import_list_new
		data["coin"] = coinName
		data["block_height"] = transData.Height
		data["timestamp"] = transData.Time
		data["tx_id"] = txid
		data["hash"] = transData.Hash
		data["tx_fee"] = feeFloat.String()
		data["confirmations"] = transData.Confirmations
		data["create_at"] = time.Now().Unix()
		data["tx_fee_coin"] = coinName
		//非挖矿交易解析vin
		fromAmountTotalFloat := decimal.Zero
		//fromAddrs := make([]string, 0)
		if !isCoinbase {
			for _, vin := range tx.Vin {
				if vin.Type != model.BtmSpend {
					//vin.AssetId != "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" || vin.AssetName != "btm" {
					//vin.AssetId != "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"  {
					log.Info("btm in 不关心的交易类型")
					continue
				}
				fromAmountTotalFloat = fromAmountTotalFloat.Add(vin.AmountFloat)
				//检查地址是否关注
				res, _ := dao.FcGenerateAddressGet(vin.Address)
				temp := map[string]interface{}{}
				if res != nil {
					log.Infof("btm存在关注的地址 from：%s ", res.Address)
					if fromMchId == 0 {
						//存在关注地址,记录商户，业务规定from只存在一个商户
						fromMchId = res.PlatformId
					}
					temp["mch_id"] = res.PlatformId
					temp["addr_type"] = res.Type
				} else {
					temp["mch_id"] = 0
					temp["addr_type"] = 0
				}
				temp["coin_type"] = coinName
				temp["tx_id"] = txid
				temp["hash"] = transData.Hash
				temp["dir"] = 2
				temp["vout_id"] = vin.SpentOutputId
				temp["mux_id"] = muxId
				//temp["tx_n"] = t
				temp["addr"] = vin.Address
				temp["amount"] = vin.AmountFloat.String()
				temp["from_tx_id"] = vin.SpentOutputId
				temp["is_spent"] = 1
				temp["create_at"] = time.Now().Unix()
				push_data = append(push_data, temp)

			}
		}

		toAmountTotalFloat := decimal.Zero
		for _, vout := range tx.Vout {
			if vout.Type != model.BtmControl {
				//vout.AssetId != "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" || vout.AssetName != "btm" {
				//vout.AssetId != "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"  {
				log.Info("btm out 不关心的交易类型")
				continue
			}
			toAmountTotalFloat = toAmountTotalFloat.Add(vout.AmountFloat)
			//多次推送
			//检查接收地址
			voutInfo, _ := dao.FcGenerateAddressGet(vout.Address)
			if voutInfo == nil {
				isTransfer = true
				//重置0
				voutInfo = new(entity.FcGenerateAddressList)
			} else {
				log.Info("btm存在关注的地址 to :%s ", voutInfo.Address)
				if voutInfo.Type == 3 {
					//手续费
					isFeeTx = true
				} else if voutInfo.Type == 1 {
					//归集或者找零
					//todo 后续需要补充a商户往b商户冷地址充值问题
					isCollectOrChange = true
				} else if voutInfo.Type == 2 {
					//常规出账
					isTransfer = true
				}
			}
			temp := map[string]interface{}{}
			temp["mch_id"] = voutInfo.PlatformId
			temp["addr_type"] = voutInfo.Type
			temp["coin_type"] = coinName
			temp["tx_id"] = txid
			temp["hash"] = transData.Hash
			temp["dir"] = 1 //1接收，2发送
			temp["tx_n"] = vout.Position
			temp["addr"] = vout.Address
			temp["amount"] = vout.AmountFloat.String()

			temp["from_tx_id"] = ""
			temp["is_spent"] = 0
			temp["create_at"] = transData.Time
			temp["mux_id"] = muxId
			temp["vout_id"] = vout.VoutId
			push_data = append(push_data, temp)

			item_arr := map[string]interface{}{} //推送结构
			item_arr["coin"] = coinName
			item_arr["coin_type"] = vout.AssetName //暂无代币，目前是btm
			//item_arr["is_in"] = 2 //1接收 2发送
			item_arr["block_height"] = transData.Height
			item_arr["timestamp"] = transData.Time
			item_arr["transaction_id"] = txid
			item_arr["txid"] = txid
			item_arr["trx_n"] = vout.Position
			item_arr["from_address"] = "" //推送的话，会清空from，用户最终需要在关注的只是出账
			item_arr["to_address"] = vout.Address
			item_arr["fee"] = tx.FeeFloat.String()
			item_arr["confirmations"] = transData.Confirmations
			item_arr["confirm_time"] = transData.Time
			item_arr["from_trx_id"] = ""
			item_arr["memo"] = ""
			item_arr["contract_address"] = ""
			item_arr["amount"] = vout.AmountFloat.String()
			dataConfirmations := transData.Confirmations
			//utxo模型限制
			if isTransfer {
				//首先给发送方推送
				if !isPushFrom && fromMchId != 0 {
					item_arr["user_sub_id"] = fromMchId
					item_arr["is_in"] = 2
					if dataConfirmations > 6 {
						for i := 1; i <= 6; i++ {
							item_arr["confirmations"] = i
							item_str, _ := json.Marshal(item_arr)
							redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, fromMchId)))
						}
					} else {
						item_str, _ := json.Marshal(item_arr)
						redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, fromMchId)))
					}
				}
				//接收方
				item_arr["is_in"] = 1
				item_arr["user_sub_id"] = voutInfo.PlatformId
				if voutInfo.PlatformId != 0 {
					if dataConfirmations > 6 {
						for i := 1; i <= 6; i++ {
							item_arr["confirmations"] = i
							item_str, _ := json.Marshal(item_arr)
							redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, voutInfo.PlatformId)))
						}
					} else {
						item_str, _ := json.Marshal(item_arr)
						redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, voutInfo.PlatformId)))
					}
				}

			} else if isFeeTx {
				log.Infof("手续费交易，忽略：%s", transData.Txs)
				break
			} else if isCollectOrChange {
				log.Infof("归集或者找零的地址，忽略：%s", vout.Address)
				continue
			}
		}
		item_dump["push_data"] = push_data
		trans_push := []map[string]interface{}{}
		if len(push_data) > 0 {
			for _, vv := range push_data {
				jump := map[string]interface{}{}
				jump["coin_type"] = vv["coin_type"]
				jump["is_in"] = vv["dir"]
				jump["block_height"] = transData.Height
				jump["timestamp"] = transData.Time
				jump["transaction_id"] = vv["tx_id"]
				jump["txid"] = vv["tx_id"]
				jump["hash"] = transData.Hash
				jump["trx_n"] = vv["tx_n"]
				jump["from_trx_id"] = vv["from_tx_id"]
				jump["to_address"] = vv["addr"]
				jump["address"] = vv["addr"]
				jump["amount"] = vv["amount"]
				jump["fee"] = feeFloat.String()
				jump["confirmations"] = transData.Confirmations
				jump["app_id"] = vv["mch_id"]
				jump["user_sub_id"] = vv["mch_id"]
				jump["type"] = vv["addr_type"]
				jump["push_state"] = 1
				jump["mux_id"] = vv["mux_id"]
				jump["vout_id"] = vv["vout_id"]
				trans_push = append(trans_push, jump)
			}
		}
		item_dump["trans_push"] = trans_push
		tx_in_amount := fromAmountTotalFloat.String()
		tx_in_num := len(tx.Vin)

		tx_out_amount := toAmountTotalFloat.String()
		tx_out_num := len(tx.Vout)
		data["coin_type"] = coinName
		data["tx_in_amount"] = tx_in_amount
		data["tx_in_num"] = tx_in_num
		data["tx_out_amount"] = tx_out_amount
		data["tx_out_num"] = tx_out_num
		item_dump["clear"] = data
		if item_dump != nil {
			dump_str, _ := json.Marshal(item_dump)
			redisHelper.LeftPush("import_list_new", string(dump_str))
		}
	}
	return nil

}
