package transpush

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/shopspring/decimal"
)

// 账号模型
type AccountModel struct {
	// nothing
}

var retryError = errors.New("need retry")

func (r *AccountModel) Run(reqexit <-chan bool) {
	defer WaitGroupTransPush.Done()
	tempFunc := func() {
		redisHelper, err := util.AllocRedisClient()
		if err != nil {
			log.Error(err)
			// WaitGroupTransPush.Done()
			return
		}
		defer redisHelper.Close()

		log.Debug("Run AccountModel")
		run := true
		for run {
			select {
			case s := <-reqexit:
				log.Error("AccountModel exit", s)
				run = false
				break
			default:
				item, err := redisHelper.Rpop("account_list_new")
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
					log.Infof("准备执行dispossAccount %s", string(item))
					// dispossAccount(item, redisHelper)
					if err := dispossAccount(item, redisHelper); err != nil {
						if err == retryError {
							redisHelper.LeftPush("account_list_new", string(item))
						}
					}
				}
			}
		}
	}
	go tempFunc()
	go tempFunc()
	go tempFunc()
	go tempFunc()
	go tempFunc()
	go tempFunc()
	tempFunc()
	// WaitGroupTransPush.Done()
}

func setCsprToAddr(transData *model.PushAccountBlockInfo) {
	var txs []model.PushAccountTx
	for _, tx := range transData.Txs {
		if tx.To != "" {
			// 说明已经有to地址
			log.Infof("setCsprToAddr txId=%s to地址不为空，跳过 %s", tx.Txid, tx.To)
			txs = append(txs, tx)
			continue
		}
		hotOrder, err := dao.FcOrderHotFindTxId(tx.Txid)
		if err != nil {
			log.Errorf("getCsprToAddr出错:%s", err.Error())
			txs = append(txs, tx)
			continue
		}
		log.Infof("setCsprToAddr txId=%s 设置了to地址 %s", tx.Txid, hotOrder.ToAddress)
		newTx := tx
		newTx.To = hotOrder.ToAddress
		txs = append(txs, newTx)
	}
	transData.Txs = txs
}

// 账号模型入账
func dispossAccount(postdata []byte, redisHelper *util.RedisClient) error {
	log.Infof("dispossAccount处理数据 %s", string(postdata))
	var (
		fuleData       map[string]interface{} // 燃料数据，针对双币链的 额外存储手续费消耗,需要在推送之后添加到数据库
		trans_data     model.PushAccountBlockInfo
		coin_type_name string // 暂时替换使用，coin_name容易导致填充默认值,且尚未梳理完整，暂不替换，2020年04月25日
		main_coin_name string // 暂时替换使用，coin_name容易导致填充默认值，且尚未梳理完整，暂不替换，2020年04月25日
		fromMchId      int    // 一般来说账户模型 from只有一个地址，因此存储商户ID
		fromAddrType   int    // 一般来说账户模型 from只有一个地址，因此地址类型
		isNeoAirDropTx bool   // 是否是neo的空投手续费交易
	)
	err := json.Unmarshal(postdata, &trans_data)
	if err != nil {
		log.Debug(err)
		return err
	}
	log.Info("推送问题1")
	if len(trans_data.Txs) == 0 {
		return errors.New("param error")
	}

	coin_name := strings.ToLower(trans_data.CoinName)
	main_coin_name = strings.ToLower(trans_data.CoinName)

	if "cspr" == coin_name {
		setCsprToAddr(&trans_data)
		log.Infof("setCsprToAddr 设置to地址完成 %v", trans_data)
	}

	// 查找主链币中是否存在
	dbMainCoinInfo, err := dao.FcCoinSetGetByName(main_coin_name, 1)
	// result, err := dao.TransPushGet(&entity.FcCoinSet{}, "select  name from fc_coin_set where name = ?", coin_name)
	if err != nil || dbMainCoinInfo == nil {
		log.Debug(fmt.Sprintf("don't find coinname :%s", coin_name), err)
		return nil
	}
	log.Info("推送问题2")

	var item_dump map[string]interface{} = make(map[string]interface{})
	var data map[string]interface{} = make(map[string]interface{})
	main_trans := &trans_data.Txs[0]
	txid := main_trans.Txid
	blockhash := trans_data.Hash
	fee := ""
	switch main_trans.Fee.(type) {
	case float64:
		fee = decimal.NewFromFloat(main_trans.Fee.(float64)).String()
	case int64:
		fee = decimal.NewFromInt(main_trans.Fee.(int64)).String()
	case string:
		fee = main_trans.Fee.(string)
	}
	data["coin"] = coin_name
	data["block_height"] = trans_data.Height
	data["timestamp"] = trans_data.Time
	data["tx_id"] = txid
	data["hash"] = blockhash
	data["tx_fee"] = fee
	// if coin_name == "hnt" {
	//	//回查order表替换
	//	orderInfo, err := dao.FcOrderHotGetByTxid(txid, 4)
	//	if err != nil {
	//		log.Errorf("HNT 回查手续费异常，txid：%s", txid)
	//		return nil
	//	}
	//	fee = decimal.NewFromInt(orderInfo.Fee).Shift(-8).String()
	//	data["tx_fee"] = fee
	// } else {
	//	data["tx_fee"] = fee
	// }

	data["memo"] = main_trans.Memo
	data["confirmations"] = trans_data.Confirmations
	data["create_at"] = time.Now().Unix()
	var fee_coin string
	if coin_name == "usdt" {
		fee_coin = "btc"
	} else if coin_name == "ont" {
		fee_coin = "ong"
	} else if coin_name == "neo" {
		fee_coin = "gas"
	} else if coin_name == "vet" {
		fee_coin = "vtho"
	} else {
		fee_coin = coin_name
	}

	data["tx_fee_coin"] = fee_coin

	item_dump["type"] = trans_data.Type
	// var is_ong bool = false
	var is_ignore_ong_fee bool = false // 是否已经忽略过ong手续费,只需要处理一次
	log.Infof("推送问题3 %s", txid)

	push_data := []map[string]interface{}{}
	{
		// ont交易一般第二条是手续费支付
		// if len(trans_data.Txs) > 1 && trans_data.CoinName == "ont" {
		//	is_ong = false
		// } else {
		//	is_ong = true
		// }
		log.Infof("推送问题4 %s", txid)
		// php
		for kk, trans := range trans_data.Txs {
			// hx资产ID必须要1.3.0
			if coin_name == "hx" {
				if trans.Assetid != "1.3.0" {
					return fmt.Errorf("不支持币种hx的资产ID：%s", trans.Assetid)
				}
			}
			log.Infof("推送问题5 %s", txid)

			// 抛弃掉ont的第二笔手续费ong推送
			if strings.ToLower(trans_data.CoinName) == "ont" &&
				trans.Contract == "0200000000000000000000000000000000000000" && !is_ignore_ong_fee {
				dataFee, _ := decimal.NewFromString(fee)
				contractAmount, _ := decimal.NewFromString(trans.Amount)
				if !dataFee.IsZero() && dataFee.Equals(contractAmount) {
					// 如果手续费和金额相等，忽略
					is_ignore_ong_fee = true
				}
			}

			// 防止ont替换在后面
			contract := trans.Contract
			coin_type_name = trans.Contract

			// bnb因为历史遗留问题，暂时没法改动表数据增加主链币token,
			// 暂时怕影响php的数据统计流程，因为在bnb的链，主链币也是合约代币的一种
			if strings.ToLower(trans_data.CoinName) == "bnb" && trans.Contract == "BNB" {
				contract = ""
				coin_type_name = ""
			}
			if strings.ToLower(trans_data.CoinName) == "zvc" && trans.Contract == "zv0000000000000000000000000000000000000000000000000000000000000000" {
				contract = ""
				coin_type_name = ""
			}
			if strings.ToLower(trans_data.CoinName) == "ont" && trans.Contract == "0100000000000000000000000000000000000000" {
				contract = ""
				coin_type_name = ""
			}
			if strings.ToLower(trans_data.CoinName) == "mtr" && trans.Contract == "0" {
				contract = ""
				coin_type_name = ""
			}
			if strings.ToLower(trans_data.CoinName) == "fio" && trans.Contract == "fio.token" {
				contract = ""
				coin_type_name = ""
			}
			if strings.ToLower(trans_data.CoinName) == "sol" && trans.Contract == "11111111111111111111111111111111" {
				contract = ""
				coin_type_name = ""
			}

			if strings.ToLower(trans_data.CoinName) == "avax" && trans.Contract == "FvwEAhmxKfeiG8SnEvq42hc6whRyY3EFYAvebMqDNDGCgxN5Z" {
				contract = ""
				coin_type_name = ""
			}

			if strings.ToLower(trans_data.CoinName) == "tlos" && trans.Contract == "eosio.token" {
				contract = ""
				coin_type_name = ""
			}
			if strings.ToLower(trans_data.CoinName) == "bos" && trans.Contract == "eosio.token" {
				contract = ""
				coin_type_name = ""
			}
			if strings.ToLower(trans_data.CoinName) == "iost" && trans.Contract == "iost" {
				contract = ""
				coin_type_name = ""
			}

			if dbMainCoinInfo.Id == 0 && contract != "" {
				// 理论上如果有合约那么pid就不会等于0
				log.Infof("pid 为0,主链币：%s 合约：%s", coin_name, contract)
				return nil
			}

			if contract != "" {
				coinset := &entity.FcCoinSet{}
				log.Infof("推送问题 dao.TransPushGet %s", txid)
				result, err := dao.TransPushGet(coinset, "select pid, name from fc_coin_set where token = ? and pid = ?", contract, dbMainCoinInfo.Id)
				log.Infof("推送问题 dao.TransPushGet 完成 %s", txid)
				if err != nil {
					log.Debug(err, result)
					return retryError
				}
				if !result {
					// 如果找不到该币种,直接放弃详情
					log.Debug(fmt.Errorf("don't find token contract %s", contract), result)
					continue
				}
				// 8 是eos
				if coinset.Pid > 0 && coinset.Pid != 8 {
					// 名字替换
					coin_name = coinset.Name
					coin_type_name = coinset.Name
				}
			}

			contract_token := "" // strings.ToLower(trans_data.Token)
			if contract_token != "" && trans_data.CoinName == "eos" {
				switch contract_token {
				case "usdt", "USDT":
					coin_name = "usdt-eos"
					coin_type_name = "usdt-eos"
				case "dili":
					coin_name = "dili-eos"
					coin_type_name = "dili-eos"
				default:
					coin_name = contract_token
					coin_type_name = contract_token
				}
			}

			// {"coin":"bnb","coin_type":"bnb","is_in":1,"block_height":84009121,"timestamp":1588013626,"transaction_id":"B2163A8A4D9E2DD4C4A286D12325760E9A4818B56AE2BF64B1B1B295C5581C4D","trx_n":0,"from_address":"bnb1mdwxfew666vtmsgvc9g9n53jl0rnhlgn3j0fnt","to_address":"bnb1kcalyfje6wk7mjfrcjv9grtzfrudq0906fcxe7","amount":"0.00010000","fee":"0.000375","confirmations":1,"confirm_time":1588013626,"user_sub_id":1,"from_trx_id":"","memo":"","contract_address":"BNB"}"
			if trans.From != "" {
				address := trans.From
				amount := trans.Amount
				temp := map[string]interface{}{}
				address_info := &entity.FcGenerateAddressList{}
				log.Infof("推送问题L318 dao.TransPushGet %s", txid)
				result, err := dao.TransPushGet(address_info, "select type,platform_id from fc_generate_address_list where address = ? and coin_name = ?", address, strings.ToLower(trans_data.CoinName))
				log.Infof("推送问题L318 dao.TransPushGet 完成 %s", txid)
				if !result || err != nil {
					address_info.PlatformId = 0
					address_info.Type = 0
				} else {
					// 填充商户ID,避免特殊类型，只填充一次
					if fromMchId == 0 {
						fromMchId = address_info.PlatformId
						fromAddrType = address_info.Type
					}
				}

				if trans_data.CoinName == "neo" && address == "Any" {
					isNeoAirDropTx = true
				} else {
					isNeoAirDropTx = false
				}

				temp["mch_id"] = address_info.PlatformId
				// temp["coin_type"] = coin_name
				if coin_type_name == "" {
					temp["coin_type"] = main_coin_name
				} else {
					temp["coin_type"] = coin_type_name
				}
				temp["contract_address"] = trans.Contract
				temp["tx_id"] = trans.Txid
				temp["hash"] = blockhash
				temp["dir"] = 2
				temp["tx_n"] = kk
				if strings.ToLower(coin_name) == "crab" {
					// 目前只有一笔
					temp["tx_n"] = 0
				}

				temp["addr"] = address
				temp["amount"] = amount
				temp["addr_type"] = address_info.Type
				temp["from_tx_id"] = ""
				temp["create_at"] = time.Now().Unix()
				temp["memo_encrypt"] = trans.MemoEncrypt
				if main_coin_name == "hx" {
					temp["memo_encrypt"] = trans.Memo
					memoDecodeStr := strings.TrimLeft(trans.Memo, "0")
					memoDecode, _ := hex.DecodeString(memoDecodeStr)
					temp["memo"] = string(memoDecode)
				}
				temp["feepayer"] = trans.FeePayer

				push_data = append(push_data, temp)

				// hnt 特殊处理
				if address_info.PlatformId != 0 && coin_name == "hnt" {
					// 回查order表替换
					log.Infof("推送问题 dao.FcOrderHotGetByTxid %s", txid)
					orderInfo, err := dao.FcOrderHotGetByTxid(txid, 4)
					log.Infof("推送问题 dao.FcOrderHotGetByTxid 完成 %s", txid)

					if err != nil {
						log.Errorf("HNT 回查手续费异常，txid：%s", txid)
						return nil
					}
					fee = decimal.NewFromInt(orderInfo.Fee).Shift(-8).String()
					data["tx_fee"] = fee
				}

				// 针对双币链,特殊处理
				fuleData = splitFuelCoin(temp, fee)
				if fuleData != nil {
					// 手续费置为空，不然按php模式会双重扣费
					data["tx_fee"] = "0"
				}

			}

			if trans.To != "" {
				log.Infof("推送问题5 %s", txid)
				address := trans.To
				amount := trans.Amount
				decamount, _ := decimal.NewFromString(trans.Amount)
				temp := map[string]interface{}{}
				res := &entity.FcGenerateAddressList{}
				log.Infof("推送问题6 coinName=%s %s", trans_data.CoinName, txid)
				result, err := dao.TransPushGet(res, "select type,platform_id from fc_generate_address_list where address = ? and coin_name = ?", address, strings.ToLower(trans_data.CoinName))
				log.Infof("推送问题6 完成 %s", trans_data.CoinName, txid)

				// log.Debug(result, err)
				if !result || err != nil {
					res.PlatformId = 0
					res.Type = 0
				}

				temp["mch_id"] = res.PlatformId
				// temp["coin_type"] = coin_name
				if coin_type_name == "" {
					temp["coin_type"] = main_coin_name
				} else {
					temp["coin_type"] = coin_type_name
				}

				temp["tx_id"] = trans.Txid
				temp["hash"] = blockhash
				temp["dir"] = 1
				temp["tx_n"] = kk
				if strings.ToLower(coin_name) == "crab" {
					// 目前只有一笔
					temp["tx_n"] = 0
				}

				temp["contract_address"] = trans.Contract
				temp["addr"] = address
				temp["amount"] = amount
				temp["addr_type"] = res.Type
				temp["from_tx_id"] = ""
				temp["create_at"] = time.Now().Unix()
				temp["memo_encrypt"] = trans.MemoEncrypt
				if main_coin_name == "hx" {
					temp["memo_encrypt"] = trans.Memo
					memoDecodeStr := strings.TrimLeft(trans.Memo, "0")
					memoDecode, _ := hex.DecodeString(memoDecodeStr)
					temp["memo"] = string(memoDecode)
				}
				push_data = append(push_data, temp)

				// to地址
				// bnb暂时添加手续费提醒
				log.Infof("推送问题7 %s", txid)
				log.Info("推送问题7")

				log.Info(res.OutOrderid, "   res.Type: ", res.Type)
				log.Info(res.OutOrderid, "   fromMchId: ", fromMchId)
				log.Info(res.OutOrderid, "   res.PlatformId: ", res.PlatformId)
				log.Info(res.OutOrderid, "   trans_data.CoinName: ", trans_data.CoinName)
				log.Info(res.OutOrderid, "   fromAddrType: ", fromAddrType)

				if in_array(res.Type, []interface{}{0, 2, 6}) ||
					(fromMchId != res.PlatformId && fromAddrType == 1 && in_array(strings.ToLower(trans_data.CoinName), []interface{}{"kava", "luna", "lunc", "ksm", "crab", "bsc", "crust"})) ||
					(in_array(strings.ToLower(trans_data.CoinName), []interface{}{"bos", "tlos", "eos", "fo", "mdu", "stg", "cocos", "gxc", "bnb", "atom", "yta", "cfx", "xlm", "dip", "stx", "heco", "trx", "nyzo", "xdag", "iost"}) && temp["addr_type"] == 1) ||
					(in_array(strings.ToLower(trans_data.CoinName), []interface{}{"iotx", "matic-matic", "wtc", "moac", "hsc", "okt", "waves", "glmr", "avaxcchain", "eth", "bnb", "mtr", "cfx", "bsc", "kai", "rbtc", "movr", "sep20", "optim", "brise-brise", "ftm", "welups", "rose", "one", "rev", "tkm", "ron", "neo", "sol", "icp", "flow", "uenc", "btm", "cspr", "pcx", "trx", "rei", "aur", "dscc", "mob", "dscc1", "lat", "deso", "nodle"}) && res.Type == 3) ||
					(in_array(strings.ToLower(trans_data.CoinName), []interface{}{"iotx", "matic-matic", "wtc", "moac", "hsc", "okt", "waves", "glmr", "avaxcchain", "eth", "pcx", "heco", "trx", "kai", "rbtc", "movr", "sep20", "optim", "brise-brise", "ftm", "welups", "rose", "one", "rev", "tkm", "ron", "neo", "sol", "icp", "flow", "uenc", "btm", "cspr", "rei", "aur", "dscc", "mob", "dscc1", "lat", "deso", "nodle"}) && fromMchId != res.PlatformId) {

					log.Infof("in")

					if trans_data.CoinName == "neo" {
						log.Infof("trans_data.CoinName ==  neo")
						if isNeoAirDropTx {
							log.Infof("trans_data.CoinName ==  neo  ===> isNeoAirDropTx")
							goto SKIP_EX
						}
						if address == "neo-coinbase" {
							log.Infof("2.trans_data.CoinName ==  neo  ===> isNeoAirDropTx")
							goto SKIP_EX
						}
					}

					if trans_data.CoinName == "sol" {
						if address == "create" || address == "fee" {
							log.Infof("trans_data.CoinName ==  sol  ===> skip ex")
							goto SKIP_EX
						}
					}

					if trans_data.CoinName == "neo" || trans_data.CoinName == "okt" || trans_data.CoinName == "waves" ||
						trans_data.CoinName == "glmr" || trans_data.CoinName == "icp" || trans_data.CoinName == "flow" ||
						trans_data.CoinName == "rbtc" || trans_data.CoinName == "sol" || trans_data.CoinName == "tkm" ||
						trans_data.CoinName == "movr" || trans_data.CoinName == "sep20" || trans_data.CoinName == "ccn" ||
						trans_data.CoinName == "optim" || trans_data.CoinName == "rev" || trans_data.CoinName == "one" ||
						trans_data.CoinName == "rose" || trans_data.CoinName == "welups" || trans_data.CoinName == "ftm" ||
						trans_data.CoinName == "ron" || trans_data.CoinName == "brise-brise" || trans_data.CoinName == "rei" ||
						trans_data.CoinName == "aur" || trans_data.CoinName == "dscc" || trans_data.CoinName == "mob" || trans_data.CoinName == "dscc1" ||
						trans_data.CoinName == "lat" || trans_data.CoinName == "deso" || trans_data.CoinName == "nodle" {
						if res.Type == 1 && fromAddrType == 0 {
							goto SKIP_EX
						}
					}
					log.Info(res.OutOrderid, "==> in.............")
					// asko sta 暂时抛弃指定的地址
					if strings.ToLower(trans_data.CoinName) == "eth" {
						if strings.ToLower(trans.Contract) == "0xeeee2a622330e6d2036691e983dee87330588603" {
							// asko 回收交易忽略
							if address == "0xe346a9fa3414645d4ac383d47705b8f663a37b58" {
								continue
							}
						} else if strings.ToLower(trans.Contract) == "0xa7de087329bfcda5639247f96140f9dabe3deed1" {
							// sta 销毁忽略
							if address == "0x0000000000000000000000000000000000000000" {
								continue
							}
						}
					}

					if "bsc" == strings.ToLower(trans_data.CoinName) {
						if isBscDestroyTx(trans.Contract, trans.To) {
							log.Infof("txId=%s to地址=%s amount=%s BSC销毁交易，不推给交易所", trans.Txid, trans.To, trans.Amount)
							continue
						}
					}

					item_arr := map[string]interface{}{}
					item_arr["coin"] = strings.ToLower(trans_data.CoinName)
					// item_arr["coin_type"] = coin_name
					if coin_type_name == "" {
						item_arr["coin_type"] = main_coin_name
					} else {
						item_arr["coin_type"] = coin_type_name
					}

					item_arr["is_in"] = 1
					item_arr["block_height"] = trans_data.Height
					item_arr["timestamp"] = trans_data.Time
					item_arr["transaction_id"] = trans.Txid
					item_arr["txid"] = trans.Txid
					item_arr["trx_n"] = kk
					item_arr["from_address"] = trans.From
					item_arr["to_address"] = address
					item_arr["contract_address"] = trans.Contract
					item_arr["amount"] = amount
					switch trans.Fee.(type) {
					case float64:
						item_arr["fee"] = decimal.NewFromFloat(trans.Fee.(float64)).String()
					case int64:
						item_arr["fee"] = decimal.NewFromInt(trans.Fee.(int64)).String()
					case string:
						item_arr["fee"] = trans.Fee
					}
					item_arr["confirmations"] = trans_data.Confirmations
					item_arr["confirm_time"] = trans_data.Time
					item_arr["user_sub_id"] = temp["mch_id"]
					item_arr["from_trx_id"] = ""
					item_arr["memo"] = trans.Memo
					item_arr["memo_encrypt"] = trans.MemoEncrypt
					if main_coin_name == "hx" {
						item_arr["memo_encrypt"] = trans.Memo
						memoDecodeStr := strings.TrimLeft(trans.Memo, "0")
						memoDecode, _ := hex.DecodeString(memoDecodeStr)
						item_arr["memo"] = string(memoDecode)
					}
					log.Info("推送问题8")
					// 以下三个为 入金 风控新增的字段 2021.06.24
					item_arr["is_risk"] = trans.IsRisk
					item_arr["risk_level"] = trans.RiskLevel
					item_arr["risk_msg"] = trans.RiskMsg

					item_str, _ := json.Marshal(item_arr)

					address_info2 := &entity.FcGenerateAddressList{}
					log.Infof("推送问题 dao.TransPushCount %s", txid)
					count, _ := dao.TransPushCount(address_info2, "select id from fc_generate_address_list where type = 3 and status = 2 and address = ?", item_arr["from_address"].(string))
					log.Infof("推送问题 dao.TransPushCount 完成%s", txid)
					if count == 0 && item_arr["from_address"] != item_arr["to_address"] {
						// if strings.ToLower(coin_name) == "ong" && !is_ong {
						if 1 != 1 {
							// 暂时修改ong逻辑，让他执行false
							// 屏蔽ong是手续费的情况，币种只会叫ont
						} else {
							from_address := trans.From
							addr := &entity.FcGenerateAddressList{}
							if from_address != "" {
								log.Infof("推送问题L546 dao.TransPushGet %s", txid)
								result, err := dao.TransPushGet(addr, "select type,platform_id from fc_generate_address_list where address = ? and coin_name = ?", from_address, strings.ToLower(trans_data.CoinName))
								log.Infof("推送问题L546 dao.TransPushGet 完成 %s", txid)
								if !result || err != nil {
									addr.PlatformId = 0
									addr.Type = 0
								}
							} else {
								addr.PlatformId = 0
								addr.Type = 0
							}
							// if res.Type > 0 && decamount.Cmp(decimal.NewFromInt(0)) > 0 && res.PlatformId != addr.PlatformId {
							// 增加新逻辑,币种为BNB,to为冷地址的时候不推送 2020年05月11日
							// to推送
							// memo模型存在1类型
							if in_array(res.Type, []interface{}{1, 2}) && decamount.Cmp(decimal.NewFromInt(0)) > 0 && res.PlatformId != addr.PlatformId {
								if in_array(strings.ToLower(trans_data.CoinName), []interface{}{"eos", "stg", "cocos", "kava", "luna", "lunc", "seek", "ont", "ar", "tlos"}) && trans_data.Confirmations > 6 {
									if !is_ignore_ong_fee {
										nums := 6
										for i := 4; i <= nums; i++ {
											item_arr["confirmations"] = i
											item_str, _ := json.Marshal(item_arr)
											log.Info("塞入数据:", item_arr["transaction_id"])

											redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, res.PlatformId)))
										}
									}

								} else if in_array(strings.ToLower(trans_data.CoinName), []interface{}{"trx"}) && trans_data.Confirmations > 25 {
									nums := 25
									for i := 23; i <= nums; i++ {
										item_arr["confirmations"] = i
										item_str, _ := json.Marshal(item_arr)
										log.Info("塞入数据:", item_arr["transaction_id"])
										redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, res.PlatformId)))
									}
								} else if in_array(strings.ToLower(trans_data.CoinName), []interface{}{"matic-matic", "wtc", "moac", "hsc", "eth", "etc",
									"gxc", "bnb", "cds", "ksm", "bnc", "crab", "celo", "mtr", "fio", "nas", "bsc", "fil", "wd", "crust",
									"near", "atom", "yta", "cfx", "star", "fis", "atp", "cph-cph", "xlm", "pcx", "dip",
									"algo", "ori", "bos", "okt", "glmr", "avaxcchain", "heco", "waves", "nyzo", "xdag", "iost", "dom", "rbtc", "movr", "sep20", "ccn", "optim", "brise-brise", "ftm", "welups", "rose", "one", "rev", "tkm", "ron", "kai", "neo", "icp", "flow", "uenc", "btm", "cspr", "rei", "aur", "dscc", "mob", "dscc1", "lat", "deso", "nodle"}) && trans_data.Confirmations > 12 {
									nums := 12
									for i := 10; i <= nums; i++ {
										item_arr["confirmations"] = i
										item_str, _ := json.Marshal(item_arr)
										log.Info("塞入数据:", item_arr["transaction_id"])
										redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, res.PlatformId)))
									}
								} else if in_array(strings.ToLower(trans_data.CoinName), []interface{}{"stx"}) && trans_data.Confirmations > 3 {
									nums := 3
									for i := 1; i <= nums; i++ {
										item_arr["confirmations"] = i
										item_str, _ := json.Marshal(item_arr)
										log.Info("塞入数据:", item_arr["transaction_id"])
										redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, res.PlatformId)))
									}
								} else if in_array(strings.ToLower(trans_data.CoinName), []interface{}{"hnt", "sol", "tlos", "dot", "azero", "sgb-sgb", "kar", "mw", "dhx"}) && trans_data.Confirmations > 6 {
									// 新逻辑。重推的时候确认数+1
									nums := 6
									for i := 4; i <= nums; i++ {
										item_arr["confirmations"] = i
										item_str, _ := json.Marshal(item_arr)
										log.Info("塞入数据:", item_arr["transaction_id"])
										redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, res.PlatformId)))
									}

								} else {
									if !is_ignore_ong_fee {
										log.Info("塞入数据:", item_arr["transaction_id"])
										redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, res.PlatformId)))
									}
								}
							}

							// from推送
							if from_address != "" && decamount.Cmp(decimal.NewFromInt(0)) > 0 && res.PlatformId != addr.PlatformId {
								if addr.Type > 0 && addr.PlatformId > 0 {
									if in_array(strings.ToLower(trans_data.CoinName), []interface{}{"eos", "stg", "cocos", "kava", "luna", "lunc", "seek", "cocos", "ont", "tlos"}) && trans_data.Confirmations > 6 {
										if !is_ignore_ong_fee {
											for i := 4; i <= 6; i++ {
												item_arr["confirmations"] = i
												item_str, _ := json.Marshal(item_arr)
												log.Info("塞入数据:", item_arr["transaction_id"])
												redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, addr.PlatformId)))
											}
										}
									} else if in_array(strings.ToLower(trans_data.CoinName), []interface{}{"trx"}) && trans_data.Confirmations > 25 {
										for i := 23; i <= 25; i++ {
											item_arr["confirmations"] = i
											item_str, _ := json.Marshal(item_arr)
											log.Info("塞入数据:", item_arr["transaction_id"])
											redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, addr.PlatformId)))
										}
									} else if in_array(strings.ToLower(trans_data.CoinName), []interface{}{"matic-matic", "wtc", "moac", "hsc", "eth", "etc",
										"gxc", "bnb", "cds", "ksm", "bnc", "crab", "celo", "mtr", "fio", "nas", "bsc", "fil", "wd", "crust",
										"near", "atom", "yta", "cfx", "star", "fis", "atp", "cph-cph", "xlm", "pcx", "dip",
										"algo", "ori", "bos", "okt", "waves", "glmr", "avaxcchain", "heco", "nyzo", "xdag", "iost", "dom", "rbtc", "movr", "sep20", "ccn", "optim", "brise-brise", "ftm", "welups", "rose", "one", "rev", "tkm", "ron", "kai", "neo", "icp", "flow", "uenc", "btm", "cspr", "rei", "aur", "dscc", "mob", "dscc1", "lat", "deso", "nodle"}) && trans_data.Confirmations > 12 {
										for i := 10; i <= 12; i++ {
											item_arr["confirmations"] = i
											item_str, _ := json.Marshal(item_arr)
											log.Info("塞入数据:", item_arr["transaction_id"])
											redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, addr.PlatformId)))
										}
									} else if in_array(strings.ToLower(trans_data.CoinName), []interface{}{"stx"}) && trans_data.Confirmations > 3 {
										for i := 1; i <= 3; i++ {
											item_arr["confirmations"] = i
											item_str, _ := json.Marshal(item_arr)
											log.Info("塞入数据:", item_arr["transaction_id"])
											redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, addr.PlatformId)))
										}
									} else if in_array(strings.ToLower(trans_data.CoinName), []interface{}{"hnt", "sol", "tlos", "dot", "azero", "sgb-sgb", "kar", "mw", "dhx"}) && trans_data.Confirmations > 6 {
										// 新逻辑。重推的时候确认数+1
										for i := 4; i <= 6; i++ {
											item_arr["confirmations"] = i
											item_str, _ := json.Marshal(item_arr)
											log.Info("塞入数据:", item_arr["transaction_id"])
											redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, addr.PlatformId)))
										}
									} else {
										if !is_ignore_ong_fee {
											redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, addr.PlatformId)))
										}
									}
								}
							}

							// 新增逻辑  专门处理内部出账给手续费地址
							if res.PlatformId == addr.PlatformId && res.Type == 3 && addr.Type == 1 && in_array(strings.ToLower(trans_data.CoinName), []interface{}{"wtc", "moac", "hsc", "eth", "cfx", "bnb", "mtr", "stx", "bsc", "trx", "heco", "cspr"}) {
								log.Infof("：%s手续费出账", trans_data.CoinName)
								// 这种行为视为出账即可
								if trans_data.Confirmations > 12 && in_array(strings.ToLower(trans_data.CoinName), []interface{}{"bnb", "mtr", "heco"}) {
									for i := 10; i <= 12; i++ {
										item_arr["confirmations"] = i
										item_str, _ := json.Marshal(item_arr)
										log.Info("塞入数据:", item_arr["transaction_id"])
										redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, addr.PlatformId)))
									}
								} else if trans_data.Confirmations > 3 && in_array(strings.ToLower(trans_data.CoinName), []interface{}{"stx"}) {
									for i := 1; i <= 3; i++ {
										item_arr["confirmations"] = i
										item_str, _ := json.Marshal(item_arr)
										log.Info("塞入数据:", item_arr["transaction_id"])
										redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, addr.PlatformId)))
									}
								} else if trans_data.Confirmations > 25 && in_array(strings.ToLower(trans_data.CoinName), []interface{}{"trx"}) {
									for i := 23; i <= 25; i++ {
										item_arr["confirmations"] = i
										item_str, _ := json.Marshal(item_arr)
										log.Info("塞入数据:", item_arr["transaction_id"])
										redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, addr.PlatformId)))
									}
								} else {
									redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, addr.PlatformId)))
								}
							}

							// 临时处理  专门处理内部出账给用户地址
							//addr是 from, res是to
							//1 归集地址（冷地址）  2 用户地址  3 手续费地址  4 热地址  5 商户余额地址,6是接收地址
							//res.Type == 1
							if res.PlatformId == addr.PlatformId && res.Type == 2 && addr.Type == 1 {
								log.Infof("：%s内部出账", trans_data.CoinName)
								// 同时推送入账和出账
								if trans_data.Confirmations > 12 &&
									in_array(strings.ToLower(trans_data.CoinName), []interface{}{"matic-matic", "wtc", "moac", "hsc", "eth", "ksm", "fil", "wd", "bnb", "bsc", "near", "crust",
										"yta", "cfx", "fis", "atp", "cph-cph", "xlm", "pcx", "dip", "algo", "dot", "azero", "sgb-sgb", "kar", "ori", "okt", "glmr", "avaxcchain", "waves", "heco", "nyzo", "xdag", "dhx", "dom", "kai", "rbtc", "movr", "sep20", "ccn", "ftm", "optim", "brise-brise", "welups", "rose", "one", "rev", "tkm", "ron", "iost", "neo", "sol", "icp", "flow", "uenc", "btm", "cspr", "rei", "aur", "dscc", "mob", "dscc1", "lat", "deso", "nodle"}) {
									for i := 10; i <= 12; i++ {
										item_arr["confirmations"] = i
										item_str, _ := json.Marshal(item_arr)
										log.Info("塞入数据:", item_arr["transaction_id"])
										redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, addr.PlatformId)))
										redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, addr.PlatformId)))
									}
								} else if trans_data.Confirmations > 25 &&
									in_array(strings.ToLower(trans_data.CoinName), []interface{}{"trx"}) {

									for i := 23; i <= 25; i++ {
										item_arr["confirmations"] = i
										item_str, _ := json.Marshal(item_arr)
										log.Info("塞入数据:", item_arr["transaction_id"])
										redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, addr.PlatformId)))
										redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, addr.PlatformId)))
									}

								} else if trans_data.Confirmations > 3 &&
									in_array(strings.ToLower(trans_data.CoinName), []interface{}{"stx"}) {

									for i := 1; i <= 3; i++ {
										item_arr["confirmations"] = i
										item_str, _ := json.Marshal(item_arr)
										log.Info("塞入数据:", item_arr["transaction_id"])
										redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, addr.PlatformId)))
										redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, addr.PlatformId)))
									}

								} else {
									for i := 0; i < 2; i++ {
										log.Info("塞入数据:", item_arr["transaction_id"])
										redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, addr.PlatformId)))
										redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, addr.PlatformId)))
									}
								}
							}

						}
					}
				}
			SKIP_EX:
			}
		}
	}
	// if trans_data.Type == model.PushTypeAccountTX && len(push_data) == 0 {
	//	return fmt.Errorf("dispossAccount don't find any tx")
	// }

	if fuleData != nil {
		push_data = append(push_data, fuleData)
	}

	item_dump["push_data"] = push_data
	trans_push := []map[string]interface{}{}
	if push_data != nil {

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
			jump["contract_address"] = vv["contract_address"]
			jump["fee"] = fee
			jump["confirmations"] = trans_data.Confirmations
			jump["app_id"] = vv["mch_id"]
			jump["user_sub_id"] = vv["mch_id"]
			jump["type"] = vv["addr_type"]
			jump["push_state"] = 1
			jump["feepayer"] = vv["feepayer"]
			trans_push = append(trans_push, jump)
		}
	}

	item_dump["trans_push"] = trans_push
	tx_in_amount := bcadd_multi_account(func(val interface{}) string { _tx := val.(model.PushAccountTx); return _tx.Amount }, trans_data.Txs)
	tx_in_num := len(trans_data.Txs)
	tx_out_amount := bcadd_multi_account(func(val interface{}) string { _tx := val.(model.PushAccountTx); return _tx.Amount }, trans_data.Txs)
	tx_out_num := len(trans_data.Txs)
	data["tx_in_amount"] = tx_in_amount
	data["tx_in_num"] = tx_in_num
	data["tx_out_amount"] = tx_out_amount
	data["tx_out_num"] = tx_out_num
	if coin_type_name == "" {
		data["coin_type"] = data["coin"]
	} else {
		data["coin_type"] = coin_type_name
	}
	item_dump["clear"] = data

	// 查询是否存在该币种设置
	log.Infof("推送问题L819 dao.TransPushGet %s", txid)
	result, err := dao.TransPushGet(&entity.FcCoinSet{}, "select  name from fc_coin_set where name = ?", data["coin_type"])
	log.Infof("推送问题L819 dao.TransPushGet 完成 %s", txid)

	if err != nil || !result {
		log.Debug(fmt.Errorf("don't find coinname  %s", data["coin_type"]), result)
		return nil
	}

	if item_dump != nil {
		dump_str, _ := json.Marshal(item_dump)
		log.Infof("加入到import_list_new %s", txid)
		// log.Debug(string(dump_str))
		log.Infof("import_list_new str: ", string(dump_str))
		redisHelper.LeftPush("import_list_new", string(dump_str))
	} else {
		log.Infof("item_dump 为空 不加入 import_list_new %s", txid)
	}

	return nil
}

// 账号模型确认数推送
func DispossAccountConfir(data []byte) {

}

// 双币链拆分，单独在存储一笔手续费的记录
// 注意某些币种有代付手段，目前先处理vet
func splitFuelCoin(tmpRaw map[string]interface{}, fee string) map[string]interface{} {
	var (
		tpType   int
		coinType string
		feePayer string
		tmp      map[string]interface{}
	)

	feeCoinMap := map[string]string{
		"vet": "vtho",
		"neo": "gas",
	}
	tmp = util.MapCopy(tmpRaw)
	tpType, _ = strconv.Atoi(fmt.Sprintf("%v", tmp["dir"]))
	feePayer = fmt.Sprintf("%v", tmp["feepayer"])
	// 1接收，2发送
	if tpType != 2 {
		return nil
	}
	if in_array(tmp["coin_type"], []interface{}{"vet", "neo"}) {
		coinType = fmt.Sprintf("%v", tmp["coin_type"])
		coinName := feeCoinMap[coinType]
		if coinName == "" {
			log.Errorf("缺少map配置手续费币种")
			return nil
		}
		// 金额替换 币种替换
		tmp["coin_type"] = coinName
		// tmp["amount"] = fee
		tmp["amount"] = fee
		if feePayer != "" {
			tmp["addr"] = feePayer
		}
		return tmp
	}
	return nil
}

func isBscDestroyTx(contract string, toAdd string) bool {
	contract = strings.ToLower(contract)
	log.Infof("判断是否BSC 销毁地址 contract=%s to地址=%s", contract, toAdd)
	if "0x6e59913074cd836c2904ddd5b81e85ec11bd0e02" != contract && "0xc748673057861a797275cd8a068abb95a902e8de" != contract {
		log.Infof("contract 非销毁币合约 %s 可以推送给交易所", contract)
		return false
	}
	toAdd = strings.ToLower(toAdd)
	if "0x06e8019f083368febd06f14dab3f1b1487962c1d" == toAdd || "0x0000000000000000000000000000000000000000" == toAdd {
		return true
	}

	log.Infof("不是BSC销毁地址交易，可以推送给交易所")
	return false
}
