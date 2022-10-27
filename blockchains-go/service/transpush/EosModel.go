package transpush

import (
	"encoding/json"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

// eos模型
type EosModel struct {
	// nothing
}

func (r *EosModel) Run(reqexit <-chan bool) {
	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		log.Error(err)
		WaitGroupTransPush.Done()
		return
	}
	defer redisHelper.Close()

	log.Debug("Run EosModel")
	run := true
	for run {
		select {
		case s := <-reqexit:
			log.Error("EosModel exit", s)
			run = false
			break

		default:
			item, err := redisHelper.Rpop("eos_push_list_new")
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
				log.Debug(string(item))
				if err := dispossEos(item, redisHelper); err != nil {
					redisHelper.LeftPush("eos_push_list_new", string(item))
				}
			}
		}
	}
	WaitGroupTransPush.Done()
}

// 账号模型入账
func dispossEos(postdata []byte, redisHelper *util.RedisClient) error {

	eosCoinInfo, err := dao.FcCoinSetGetCoinId("eos", "")
	if err != nil {
		log.Debug("没有找到eos相关信息")
		return err
	}

	var trans_data model.PushEosBlockInfo
	err = json.Unmarshal(postdata, &trans_data)
	if err != nil {
		log.Debug(err)
		return err
	}

	for _, trans := range trans_data.Txs {
		txid := trans.Txid
		fee := ""
		switch trans.Fee.(type) {
		case float64:
			fee = decimal.NewFromFloat(trans.Fee.(float64)).String()
		case int64:
			fee = decimal.NewFromInt(trans.Fee.(int64)).String()
		case string:
			fee = trans.Fee.(string)
		}
		blockhash := trans_data.Hash

		for ko, action := range trans.Actions {
			item_dump := map[string]interface{}{}
			push_data := []map[string]interface{}{}
			trans_push := []map[string]interface{}{}
			data := map[string]interface{}{}

			lower_coin_name := strings.ToLower(trans_data.CoinName)
			data["coin"] = lower_coin_name
			data["coin_type"] = lower_coin_name
			data["block_height"] = trans_data.Height
			data["timestamp"] = trans_data.Time
			data["tx_id"] = txid
			data["hash"] = blockhash
			data["tx_fee"] = fee
			data["trans_status"] = trans.Status
			data["memo"] = action.Memo
			data["tx_n"] = ko
			data["confirmations"] = trans_data.Confirmations
			data["create_at"] = time.Now().Unix()
			data["tx_fee_coin"] = lower_coin_name
			item_dump["type"] = trans_data.Type
			contract := ""
			if action.Contract != "" {
				contract = action.Contract
			}

			contract_token := ""
			//if action.Token != "" {
			//	contract_token = action.Token
			//}

			coin_name := ""
			if strings.ToLower(contract) != "eosio.token" {
				if contract != "" {
					coinset := &entity.FcCoinSet{}
					result, _ := dao.TransPushGet(coinset, "select name from fc_coin_set where   token = ? and pid =? and (name = ? or name = ?) ",
						contract,
						eosCoinInfo.Id,
						action.Token,
						action.Token+"-EOS")
					if !result {
						log.Debug("不支持的合约")
						continue
					}
					//if coinset.Pid < 0 {
					//	log.Debug("不支持的合约")
					//	continue
					//}
					contract_token = coinset.Name
				} else {
					log.Debug("合约不能为空")
					continue
				}
				if contract_token != "" {
					if lower_coin_name == "eos" {
						switch contract_token {
						case "usdt", "USDT":
							coin_name = "usdt-eos"
						case "dili":
							coin_name = "dili-eos"
						default:
							coin_name = strings.ToLower(contract_token)
						}
					} else {
						coin_name = strings.ToLower(contract_token)
					}
				} else {
					log.Debug("代币名称不能为空")
					continue
				}
			} else {
				coin_name = strings.ToLower(action.Token)
			}

			amount := action.Amount
			if action.From != "" {
				address := action.From
				temp := map[string]interface{}{}
				res := &entity.FcGenerateAddressList{}
				result, err := dao.TransPushGet(res, "select type,platform_id from fc_generate_address_list where address = ? and coin_name = ?", address, lower_coin_name)
				if !result || err != nil {
					res.PlatformId = 0
					res.Type = 0
				}

				temp["mch_id"] = res.PlatformId
				temp["coin_type"] = coin_name
				temp["tx_id"] = trans.Txid
				temp["hash"] = blockhash
				temp["dir"] = 2
				temp["tx_n"] = ko
				temp["addr"] = address
				temp["amount"] = amount
				temp["addr_type"] = res.Type
				temp["from_tx_id"] = ""
				temp["create_at"] = time.Now().Unix()
				temp["contract_address"] = action.Contract
				push_data = append(push_data, temp)
			}

			if action.To != "" {
				address := action.To
				decamount, _ := decimal.NewFromString(action.Amount)
				temp := map[string]interface{}{}
				res := &entity.FcGenerateAddressList{}
				result, err := dao.TransPushGet(res, "select type,platform_id from fc_generate_address_list where address = ? and coin_name = ?", address, lower_coin_name)
				if !result || err != nil {
					res.PlatformId = 0
					res.Type = 0
				}

				temp["mch_id"] = res.PlatformId
				temp["coin_type"] = coin_name
				temp["tx_id"] = trans.Txid
				temp["hash"] = blockhash
				temp["dir"] = 1
				temp["tx_n"] = ko
				temp["addr"] = address
				temp["amount"] = amount
				temp["addr_type"] = res.Type
				temp["from_tx_id"] = ""
				temp["create_at"] = time.Now().Unix()
				temp["contract_address"] = action.Contract
				push_data = append(push_data, temp)

				if in_array(res.Type, []interface{}{0, 2, 6}) || (in_array(lower_coin_name, []interface{}{"eos", "fo"}) && temp["addr_type"] == 1) {
					item_arr := map[string]interface{}{}
					item_arr["coin"] = lower_coin_name
					item_arr["coin_type"] = coin_name
					item_arr["is_in"] = 1
					item_arr["block_height"] = trans_data.Height
					item_arr["timestamp"] = trans_data.Time
					item_arr["transaction_id"] = trans.Txid
					item_arr["txid"] = trans.Txid
					item_arr["trx_n"] = ko
					item_arr["from_address"] = action.From
					item_arr["to_address"] = address
					item_arr["amount"] = amount
					item_arr["fee"] = fee
					item_arr["confirmations"] = trans_data.Confirmations
					item_arr["confirm_time"] = trans_data.Time
					item_arr["user_sub_id"] = temp["mch_id"]
					item_arr["from_trx_id"] = ""
					item_arr["memo"] = action.Memo
					item_arr["contract_address"] = contract
					item_str, _ := json.Marshal(item_arr)

					address_info2 := &entity.FcGenerateAddressList{}
					count, _ := dao.TransPushCount(address_info2, "select id from fc_generate_address_list where type = 3 and status = 2 and address = ?", item_arr["from_address"].(string))
					if count == 0 && item_arr["from_address"] != item_arr["to_address"] {
						from_address := action.From
						addr := &entity.FcGenerateAddressList{}
						if from_address != "" {
							result, err := dao.TransPushGet(addr, "select type,platform_id from fc_generate_address_list where address = ? and coin_name = ?", from_address, lower_coin_name)
							if !result || err != nil {
								addr.PlatformId = 0
								addr.Type = 0
							}
						} else {
							addr.PlatformId = 0
							addr.Type = 0
						}
						if res.Type > 0 && decamount.Cmp(decimal.NewFromInt(0)) > 0 && res.PlatformId != addr.PlatformId {
							if in_array(lower_coin_name, []interface{}{"eth", "nas", "etc", "eos", "fo"}) && trans_data.Confirmations >= 6 {
								for i := 1; i <= 6; i++ {
									item_arr["confirmations"] = i
									item_str, _ := json.Marshal(item_arr)
									redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, res.PlatformId)))
								}
							} else {
								redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, res.PlatformId)))
							}
						}

						if from_address != "" && decamount.Cmp(decimal.NewFromInt(0)) > 0 && res.PlatformId != addr.PlatformId {
							if addr.Type > 0 && addr.PlatformId > 0 {
								if in_array(strings.ToLower(trans_data.CoinName), []interface{}{"eth", "nas", "etc", "eos", "fo"}) && trans_data.Confirmations >= 6 {
									for i := 1; i <= 6; i++ {
										item_arr["confirmations"] = i
										item_str, _ := json.Marshal(item_arr)
										redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, addr.PlatformId)))
									}
								} else {
									redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, addr.PlatformId)))
								}
							}
						}
					}
				}
			}

			data["tx_in_amount"] = amount
			data["tx_in_num"] = 1
			data["tx_out_amount"] = amount
			data["tx_out_num"] = 1
			if coin_name != "" {
				data["coin_type"] = coin_name
			}
			item_dump["push_data"] = push_data
			for _, vv := range push_data {
				jump := map[string]interface{}{}
				jump["coin_type"] = vv["coin_type"]
				jump["is_in"] = vv["dir"]
				jump["block_height"] = trans_data.Height
				jump["timestamp"] = trans_data.Time
				jump["transaction_id"] = vv["tx_id"]
				jump["txid"] = vv["tx_id"]
				jump["hash"] = blockhash
				jump["trx_n"] = vv["tx_n"]
				jump["from_trx_id"] = vv["from_tx_id"]
				if vv["dir"] == 2 {
					jump["from_address"] = vv["addr"]
					jump["to_address"] = ""
					jump["address"] = vv["addr"]
				} else {
					jump["from_address"] = ""
					jump["to_address"] = vv["addr"]
					jump["address"] = vv["addr"]
				}
				jump["amount"] = vv["amount"]
				jump["fee"] = fee
				jump["confirmations"] = trans_data.Confirmations
				jump["app_id"] = vv["mch_id"]
				jump["user_sub_id"] = vv["mch_id"]
				jump["type"] = vv["addr_type"]
				jump["push_state"] = 1
				jump["contract_address"] = vv["contract_address"]
				trans_push = append(trans_push, jump)
			}

			item_dump["trans_push"] = trans_push
			item_dump["clear"] = data
			if item_dump != nil {
				dump_str, _ := json.Marshal(item_dump)
				log.Debug(string(dump_str))
				redisHelper.LeftPush("import_list_new", string(dump_str))
			}
		}
	}
	return nil
}
