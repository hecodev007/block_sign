package transpush

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/model"
	"github.com/group-coldwallet/blockchains-go/model/address"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/shopspring/decimal"
	"xorm.io/builder"
)

// utxo模型
type UtxoModel struct {
	// nothing
}

func (r *UtxoModel) Run(reqexit <-chan bool) {
	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		log.Error(err)
		WaitGroupTransPush.Done()
		return
	}
	defer redisHelper.Close()

	log.Debug("Run UtxoModel")
	run := true
	for run {
		select {
		case s := <-reqexit:
			log.Error("UtxoModel exit", s)
			run = false
			break

		default:
			item, err := redisHelper.Rpop("ticket_list_new")
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
				if err := dispossUtxo(item, redisHelper); err != nil {
					redisHelper.LeftPush("ticket_list_new", string(item))
				}
			}
		}
	}
	WaitGroupTransPush.Done()
}

// utxo模型入账
//from 要么全部是内部地址 要么是外部地址，业务模型不会存在混用情况
//优先级是判断to的状况，如果不是外部地址或者是用户地址（此时from不能是同商户的内部地址），则视为内部转账，不进行推送
//外部地址为出账，用户地址为充值（此时from不能是同商户的内部地址），两者都需要进行推送
func dispossUtxo(postdata []byte, redisHelper *util.RedisClient) error {
	var (
		transData model.PushUtxoBlockInfo
		coinName  string //主链币币种
		//tokenName               string //代币币种
		payCoinName             string //rub是混合交易，因此只能做特殊般判断赋值
		isBase                  bool   //是否挖矿收入
		isCoinStake             bool   //qtum 合约退回
		isInternalAddrOfFrom    bool   //from是否为内部地址，如果from地址存在内部地址，那么要么是合并 要么是打手续费
		internalAddrOfFromMchId int    //如果from地址存在内部地址，则全部为这个商户的ID
		isPushToFrom            bool   //对于发送方通知一次即可
		feeCoinName             string //手续费币种 ，一般为主链币，有些特殊币种的话 特殊处理
		item_dump               map[string]interface{}
		data                    map[string]interface{}
		fee                     string //手续费

		fromIsUserACC bool
	)
	//isMerge from都为用户地址，不能掺杂其他地址，否则不推送，人工处理
	//isFee 不关心from,只要to地址是手续费地址

	err := json.Unmarshal(postdata, &transData)
	if err != nil {
		log.Debug(err)
		//返回err会进入死循环
		return nil
	}

	if len(transData.Txs) == 0 {
		return errors.New("param error")
	}

	coinName = transData.CoinName
	feeCoinName = transData.CoinName

	item_dump = make(map[string]interface{})
	data = make(map[string]interface{})

	//utxo目前只有一个交易 。多个交易会分开推送
	trans := &transData.Txs[0]
	txid := trans.Txid
	blockhash := transData.Hash
	switch trans.Fee.(type) {
	case float64:
		fee = decimal.NewFromFloat(trans.Fee.(float64)).String()
	case int64:
		fee = decimal.NewFromInt(trans.Fee.(int64)).String()
	case string:
		fee = trans.Fee.(string)
	}

	//是否是挖矿收入
	if trans.Coinbase {
		isBase = trans.Coinbase
	}
	if trans.CoinStake {
		isCoinStake = trans.CoinStake
	}

	data["coin"] = transData.CoinName
	var contract *model.PushContractTx = nil

	if len(trans.Contract) > 0 {
		contract = &trans.Contract[0]
	}

	if len(trans.Contract) > 1 {
		log.Error("目前不支持多个合约入账")
		contract = nil
	}

	in_list := trans.Vin
	out_list := trans.Vout

	//uenc,btm归集不推送
	formHasColdWallet := false
	toHasColdWallet := false
	if coinName == "uenc" || coinName == "btm" {
		coldAddrs, err := entity.FcGenerateAddressList{}.FindAddress(builder.Eq{
			"type":      address.AddressTypeCold,
			"status":    address.AddressStatusAlloc,
			"coin_name": coinName,
		})
		if err != nil {
			log.Infof("获取出账地址失败! \nfrom:[%v] \n to:[%v] \n", in_list, out_list)
			return nil
		}

		for _, addr := range coldAddrs {
			for _, in := range in_list {
				if in.Addresse == addr {
					formHasColdWallet = true
				}
			}
		}

		for _, addr := range coldAddrs {
			for _, out := range out_list {
				if out.Addresse == addr {
					toHasColdWallet = true
				}
			}
		}
	}

	isUencCollector := false
	isBtmCollector := false
	if !formHasColdWallet && toHasColdWallet {
		log.Infof("归集交易! \nfrom:[%v] \n to:[%v] \n", in_list, out_list)
		isUencCollector = true
		isBtmCollector = true
	}

	data["block_height"] = transData.Height
	data["timestamp"] = transData.Time
	data["tx_id"] = txid
	data["hash"] = blockhash
	data["tx_fee"] = fee
	data["memo"] = ""
	data["confirmations"] = transData.Confirmations
	data["create_at"] = time.Now().Unix()

	//if coinName == "usdt" {
	//	fee_coin = "btc"
	//} else {
	//	fee_coin = coin_name
	//}
	data["tx_fee_coin"] = feeCoinName
	item_dump["type"] = transData.Type //0:推送交易数据，1：推送块确认数更新
	push_data := []map[string]interface{}{}
	if len(in_list) > 0 {
		for _, vo := range in_list {
			//neo存在代币，暂时在数据服务过滤掉
			if coinName == "oneo" {
				if strings.ToLower(vo.AssetName) != "oneo" || vo.AssetId != "0xc56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b" {
					continue
				}
			}

			temp := map[string]interface{}{}
			res := &entity.FcGenerateAddressList{}
			result := false

			//输入地址
			fromAddr := vo.Addresse
			//输入金额
			fromAmount := vo.Value

			if fromAddr != "" {
				if in_array(strings.ToLower(coinName), []interface{}{"bch", "bsv", "btc-stx"}) {
					//bch bsv 需要查转换后的地址
					result, err = dao.TransPushGet(res, "select id, type,platform_id from fc_generate_address_list where compatible_address = ? ", fromAddr)
				} else {
					log.Infof("检索from地址：%s", fromAddr)
					result, err = dao.TransPushGet(res, "select id, type,platform_id from fc_generate_address_list where address = ?", fromAddr)
				}
				//if err != nil || !result {
				//	log.Infof("检索不到相关记录,外部地址：%s", vo.Addresse)
				//	res.PlatformId = 0
				//	res.Type = 0
				//}
			} else {
				//如果from地址为空,尝试从数据查询这个地址信息
				//_res := &entity.FcTxClearDetail{}
				result, err = dao.TransPushGet(res, "select id, addr_type as type,mch_id as platform_id,amount,addr from tx_clear_detail where tx_id = ? and dir = ? and tx_n = ?", vo.Txid, 1, vo.Vout)
				//if err != nil || !result {
				//	fromAddr = _res.Addr
				//	fromAmount = _res.Amount
				//} else {
				//	res.PlatformId = 0
				//	res.Type = 0
				//	fromAddr = ""
				//	fromAmount = "0"
				//}
			}
			log.Infof("检索from地址，结果：%t, ", result)
			if err != nil || !result {
				log.Infof("检索不到相关记录,外部地址：%s", vo.Addresse)
				res.PlatformId = 0
				res.Type = 0
			}

			//if vo.AssetId != nil && *vo.AssetId != "" {
			//	coinset := &entity.FcCoinSet{}
			//	result, err := dao.TransPushGet(coinset, "select name from fc_coin_set where token = ?", *vo.AssetId)
			//	if err != nil || !result {
			//		log.Debug(err)
			//		return err
			//	}
			//	tokenName = coinset.Name
			//}
			//主链币
			temp["mch_id"] = res.PlatformId
			temp["coin_type"] = coinName
			temp["tx_id"] = trans.Txid
			temp["hash"] = blockhash
			temp["dir"] = 2
			temp["tx_n"] = vo.Vout
			temp["addr"] = fromAddr
			temp["amount"] = fromAmount
			temp["addr_type"] = res.Type
			temp["from_tx_id"] = vo.Txid
			temp["is_spent"] = 1
			temp["create_at"] = time.Now().Unix()

			if res.Type != 0 || res.PlatformId != 0 {
				//只要地址存在商户ID，则from全部为内部地址，（业务模型决定）
				isInternalAddrOfFrom = true
				internalAddrOfFromMchId = res.PlatformId
			}

			if res.Type == 2 {
				fromIsUserACC = true
			}

			//主链信息附加
			push_data = append(push_data, temp)
			//处理合约数据
			//rub 代币的信息收藏在assets
			if vo.Assets != nil {
				coinset := &entity.FcCoinSet{}
				result, _ := dao.TransPushGet(coinset, "select name from fc_coin_set where token = ?", vo.Assets.AssetId)
				if !result || strings.ToLower(coinset.Name) != strings.ToLower(vo.Assets.Name) {
					//不支持的合约。整条数据抛弃
					return fmt.Errorf("不支持的utxo合约数据：%s", vo.Assets.AssetId)
				}
				//tokenName = vo.Assets.Name
				temp = map[string]interface{}{}
				temp["mch_id"] = res.PlatformId
				temp["coin_type"] = vo.Assets.Name
				temp["tx_id"] = trans.Txid
				temp["hash"] = blockhash
				temp["dir"] = 2
				temp["tx_n"] = vo.Vout
				temp["addr"] = fromAddr
				temp["amount"] = vo.Assets.AssetValue
				temp["addr_type"] = res.Type
				temp["from_tx_id"] = vo.Txid
				temp["is_spent"] = 1
				temp["create_at"] = time.Now().Unix()
				push_data = append(push_data, temp)
			}

			//todo 目前已上线代码目前没有使用
			//if vo.Assets != nil {
			//	bb := rrc20_in(vo.Assets, &trans_data, txid, blockhash, res, address, &vo, redisHelper)
			//	push_data = append(push_data, bb)
			//}
		}
	}

	//for i := len(push_data) - 1; i >= 0; i-- {
	//	if push_data[i]["addr_type"] == 0 && push_data[i]["mch_id"] == 0 {
	//		push_data = append(push_data[:i], push_data[i+1:]...)
	//	}
	//}

	//判断是否只是单纯的入账，而不是内部转账
	//is_out := false
	//if len(push_data) == 0 {
	//	is_out = true
	//}

	if len(out_list) > 0 {
		for _, vo := range out_list {
			//outAmount, _ := decimal.NewFromString(vo.Value)
			//if outAmount.LessThanOrEqual(decimal.Zero) {
			//	log.Errorf("txid =[%s],金额异常，原始金额[%s]", trans.Txid, vo.Value)
			//	continue
			//}

			if strings.ToLower(coinName) == "ckb" {
				if vo.CodeHash != "9bd7e06f3ecf4be0f2fcd2188b23f1b9fcc88e5d4b65a8637b17723bbda3cce8" &&
					vo.CodeHash != "5c5069eb0857efc65e1bca0c07df34c31663b3622fd3876c876320fc9634e2a8" {
					log.Infof("无法识别ckb codehash：%s", vo.CodeHash)
					continue
				}
			}

			if coinName == "oneo" {
				if strings.ToLower(vo.AssetName) != "oneo" || vo.AssetId != "0xc56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b" {
					log.Infof("无法识别neo assetName：%s,assetId:%s", vo.AssetName, vo.AssetId)
					continue
				}
			}

			result := false
			temp := map[string]interface{}{}
			res := &entity.FcGenerateAddressList{}
			toAddress := vo.Addresse
			//toAmount := vo.Value

			if in_array(coinName, []interface{}{"bch", "bsv", "btc-stx"}) {
				//bch bsv 需要查转换后的地址
				result, err = dao.TransPushGet(res, "select id, type,platform_id from fc_generate_address_list where compatible_address = ?", toAddress)
			} else {
				result, err = dao.TransPushGet(res, "select id, type,platform_id from fc_generate_address_list where address = ?", toAddress)
				//这里是判断stx，但是暂时线上没有该逻辑代码
				//if err == nil && result && in_array(coinName, []interface{}{"btc", "btc-stx"}) {
				//	result, err = dao.TransPushGet(res, "select id, coin_name, type,platform_id from fc_generate_address_list where compatible_address = ? ", toAddress)
				//	if err == nil && result && res.CoinName == "stx" {
				//		is_stx = true
				//	}
				//}
			}
			log.Infof("检索to地址【%s】，结果：%t, ", vo.Addresse, result)
			if err != nil || !result {
				res.PlatformId = 0
				res.Type = 0
			}

			temp["mch_id"] = res.PlatformId
			temp["coin_type"] = coinName
			temp["tx_id"] = trans.Txid
			temp["hash"] = blockhash
			temp["dir"] = 1 //1接收，2发送
			temp["tx_n"] = vo.N
			temp["addr"] = vo.Addresse
			temp["amount"] = vo.Value
			temp["addr_type"] = res.Type
			temp["from_tx_id"] = ""
			//搞不太懂 挖矿收入冻结？
			//if isBase && coinName != "btc" && res.Type > 0 {
			//	temp["is_spent"] = 2
			//} else {
			//	temp["is_spent"] = 0
			//}
			temp["is_spent"] = 0
			temp["create_at"] = time.Now().Unix()
			log.Infof("push_data add addr:%s", vo.Addresse)
			push_data = append(push_data, temp)
			//if vo.Assets != nil {
			//	bb := rrc20_out(vo.Assets, &trans_data, txid, blockhash, res, vo.Addresse, &vo, in_list, redisHelper)
			//	push_data = append(push_data, bb)
			//}
			//处理合约数据
			//rub 代币的信息收藏在assets
			if vo.Assets != nil {
				coinset := &entity.FcCoinSet{}
				result, _ := dao.TransPushGet(coinset, "select name from fc_coin_set where token = ?", vo.Assets.AssetId)
				if !result || strings.ToLower(coinset.Name) != strings.ToLower(vo.Assets.Name) {
					//不支持的合约。整条数据抛弃
					return fmt.Errorf("不支持的utxo合约数据：%s", vo.Assets.AssetId)
				}
				//tokenName = coinset.Name
				//toAmount = vo.Assets.AssetValue

				temp = map[string]interface{}{}
				temp["mch_id"] = res.PlatformId
				temp["coin_type"] = vo.Assets.Name
				temp["tx_id"] = trans.Txid
				temp["hash"] = blockhash
				temp["dir"] = 1 //1接收，2发送
				temp["tx_n"] = vo.N
				temp["addr"] = vo.Addresse
				temp["amount"] = vo.Assets.AssetValue
				temp["addr_type"] = res.Type
				temp["from_tx_id"] = ""
				temp["is_spent"] = 0
				temp["create_at"] = time.Now().Unix()
				push_data = append(push_data, temp)
				log.Infof("push_data Assets add addr:%s", vo.Addresse)
			}

			if contract != nil {
				if contract.Coin == "" {
					contractInfo, err := dao.FcCoinSetGetCoinByContract(contract.Contract)
					if err != nil {
						log.Infof("主链币种[%s],无法识别合约[%s]", transData.CoinName, contractInfo.Connect)
						continue
					}
					contract.Coin = contractInfo.Name
				}
				//如果是合约数据，则不需要推送主链手续费数据了,此处直接中断
				log.Infof("主链币种[%s],代币交易[%s]", transData.CoinName, contract.Coin)

			}

			//不推送数据
			if strings.ToLower(coinName) == "btc-stx" {
				continue
			}

			needPushUtxo := true //是否需要推送utxo数据

			//不推送usdt的btc数据
			if contract != nil && strings.ToLower(contract.Coin) == "usdt" {
				needPushUtxo = false
			}

			//推送部分
			item_arr := map[string]interface{}{}
			item_arr["coin"] = coinName
			if vo.Assets != nil {
				item_arr["coin_type"] = vo.Assets.Name
			} else {
				item_arr["coin_type"] = coinName
			}
			item_arr["is_in"] = 1
			item_arr["block_height"] = transData.Height
			item_arr["timestamp"] = transData.Time
			item_arr["transaction_id"] = trans.Txid
			item_arr["txid"] = trans.Txid
			item_arr["trx_n"] = vo.N
			item_arr["from_address"] = "" //推送的话，会清空from，用户最终需要在关注的只是出账
			if coinName == "btm" {
				if len(in_list) > 0 {
					item_arr["from_address"] = in_list[0].Addresse //这里只拿一个其实并不准确
				}
			}
			item_arr["to_address"] = vo.Addresse
			item_arr["to_raw_address"] = vo.RawAddresse
			if vo.Assets != nil {
				item_arr["amount"] = vo.Assets.AssetValue
			} else {
				item_arr["amount"] = vo.Value
			}

			if coinName == "uenc" {
				fromString, _ := decimal.NewFromString(vo.Value)
				if fromString.Shift(6).IntPart() == 0 {
					log.Infof("忽略金额为零的交易 %v", res)
					continue
				}
			}

			if coinName == "btm" {
				fromString, _ := decimal.NewFromString(vo.Value)
				if fromString.Shift(8).IntPart() == 0 {
					log.Infof("忽略金额为零的交易 %v", res)
					continue
				}
			}

			if coinName == "uenc" {
				if res.Type == 2 && fromIsUserACC {
					log.Infof("忽略不推送的交易 %v", res)
					continue
				}
			}

			if coinName == "btm" {
				if res.Type == 2 && fromIsUserACC {
					log.Infof("忽略不推送的交易 %v", res)
					continue
				}
			}

			item_arr["fee"] = fee
			//理论上来说confirmations应该是需要检查是否有过推送记录的，没有的话需要进行补全
			item_arr["confirmations"] = transData.Confirmations
			item_arr["confirm_time"] = transData.Time
			item_arr["user_sub_id"] = res.PlatformId
			item_arr["from_trx_id"] = ""
			item_arr["memo"] = ""
			item_arr["contract_address"] = ""

			dataConfirmations := transData.Confirmations
			//utxo模型限制
			if dataConfirmations > 6 {
				dataConfirmations = 6
			}

			//如果是出账, 过滤一下usdt的btc数据推送
			if res.Type == 0 {
				if isInternalAddrOfFrom {
					//设置出账币种名字
					payCoinName = item_arr["coin_type"].(string)
					//查询push表记录目前最大的确认数，例如：如果确认数为4  db不存在记录 则 补充1-4
					var dbConfirmations int64 = 0
					record, err := dao.FcPushRecordLast(trans.Txid, coinName, item_arr["coin_type"].(string), internalAddrOfFromMchId)
					if err != nil {
						log.Debugf("查询push表记录目前最大的确认数,设置为0，异常，error=[%s]", err.Error())
					}
					if record != nil {
						dbConfirmations = int64(record.Confirmations)
						if dbConfirmations > 6 {
							dbConfirmations = 6
						}
					}
					//基于数据库最大存入的确认数，补全丢失记录，例如推送12 db记录4 补全 4-12，如果已经是12 再推送一次12
					if dataConfirmations > dbConfirmations && needPushUtxo {
						for i := dbConfirmations; i < dataConfirmations; i++ {

							item_arr["confirmations"] = i + 1
							item_str, _ := json.Marshal(item_arr)
							//通知发送方
							redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, internalAddrOfFromMchId)))
						}
					} else if dataConfirmations == dbConfirmations && needPushUtxo {
						//作用于补推
						//通知发送方
						item_str, _ := json.Marshal(item_arr)
						redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, internalAddrOfFromMchId)))
					}
				}
			} else {
				log.Infof("isBase:%t,isCoinStake:%t", isBase, isCoinStake)
				//如果是挖矿收入,或者qtum gas退回不操作
				if isBase || isCoinStake {
					log.Infof("币种:%s，txid:%s,挖矿或者qtum gas退款，忽略", coinName, trans.Txid)
					//查询push表记录目前最大的确认数，例如：如果确认数为4  db不存在记录 则 补充1-4
					//var dbConfirmations int64 = 0
					//record, err := dao.FcPushRecordLast(trans.Txid, coinName, item_arr["coin_type"].(string), res.PlatformId)
					//if err != nil {
					//	log.Debugf("查询push表记录目前最大的确认数,设置为0，异常，error=[%s]", err.Error())
					//}
					//if record != nil {
					//	dbConfirmations = int64(record.Confirmations)
					//}
					////基于数据库最大存入的确认数，补全丢失记录，例如推送12 db记录4 补全 4-12，如果已经是12 再推送一次12
					//if dataConfirmations > dbConfirmations {
					//	for i := dbConfirmations; i < dataConfirmations; i++ {
					//		item_arr["confirmations"] = i + 1
					//		item_str, _ := json.Marshal(item_arr)
					//		//通知接收方
					//		redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, res.PlatformId)))
					//	}
					//} else if dataConfirmations == dbConfirmations {
					//	//作用于补推
					//	//通知接收方
					//	item_str, _ := json.Marshal(item_arr)
					//	redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, res.PlatformId)))
					//}
				} else {
					//如果是同一商户的内部地址不推送
					if res.PlatformId == internalAddrOfFromMchId {
						log.Infof("币种=[%s],txid=[%s],其中from和to地址[%s],为同一商户 [%d] 的内部地址", coinName, trans.Txid, vo.Addresse, internalAddrOfFromMchId)
						if (res.Type == 2 || res.Type == 3) && needPushUtxo {
							if in_array(strings.ToLower(coinName), []interface{}{"doge", "biw", "dash", "avax", "btc", "bsv", "ltc", "bch", "ckb", "zec", "hc", "dcr", "oneo", "bcha", "xec", "ada", "zen", "satcoin", "eac", "iota", "btm", "uenc"}) {
								item_str, _ := json.Marshal(item_arr)
								if !isPushToFrom {
									//通知发送方,只需要一次
									for i := 0; i < 2; i++ {
										//2次
										redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, internalAddrOfFromMchId)))
									}
									isPushToFrom = true
								}
								for i := 0; i < 2; i++ {
									//2次
									redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, res.PlatformId)))
								}

							}
						}
					} else {
						if internalAddrOfFromMchId != 0 && !isPushToFrom && needPushUtxo {
							//查询push表记录目前最大的确认数，例如：如果确认数为4  db不存在记录 则 补充1-4
							var dbConfirmations int64 = 0
							record, err := dao.FcPushRecordLast(trans.Txid, coinName, item_arr["coin_type"].(string), internalAddrOfFromMchId)
							if err != nil {
								log.Debugf("查询push表记录目前最大的确认数,设置为0，异常，error=[%s]", err.Error())
							}
							if record != nil {
								dbConfirmations = int64(record.Confirmations)
								if dbConfirmations > 6 {
									dbConfirmations = 6
								}
							}
							//基于数据库最大存入的确认数，补全丢失记录，例如推送12 db记录4 补全 4-12，如果已经是12 再推送一次12
							if dataConfirmations > dbConfirmations && needPushUtxo {
								for i := dbConfirmations; i < dataConfirmations; i++ {
									item_arr["confirmations"] = i + 1
									item_str, _ := json.Marshal(item_arr)
									//通知发送方
									redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, internalAddrOfFromMchId)))
								}
							} else if dataConfirmations == dbConfirmations && needPushUtxo {
								//作用于补推
								//通知发送方
								item_str, _ := json.Marshal(item_arr)
								log.Infof("grep\n%v", string(format_decimal(item_str, 2, internalAddrOfFromMchId)))
								redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, internalAddrOfFromMchId)))

							}

						}

						//过滤掉冷地址推送
						if res.Type != 2 && res.Type != 0 && res.Type != 3 {
							continue
						}

						if isUencCollector || isBtmCollector {
							continue
						}
						log.Infof("开始执行推送：%s", trans.Txid)
						//查询push表记录目前最大的确认数，例如：如果确认数为4  db不存在记录 则 补充1-4
						var dbConfirmations int64 = 0
						record, err := dao.FcPushRecordLast(trans.Txid, coinName, item_arr["coin_type"].(string), res.PlatformId)
						if err != nil {
							log.Debugf("查询push表记录目前最大的确认数,设置为0，异常，error=[%s]", err.Error())
						}
						if record != nil {
							dbConfirmations = int64(record.Confirmations)
						}
						//基于数据库最大存入的确认数，补全丢失记录，例如推送12 db记录4 补全 4-12，如果已经是12 再推送一次12
						if dataConfirmations > dbConfirmations && needPushUtxo {
							for i := dbConfirmations; i < dataConfirmations; i++ {
								item_arr["confirmations"] = i + 1
								item_str, _ := json.Marshal(item_arr)
								//通知接收方
								redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, res.PlatformId)))
							}
						} else if dataConfirmations <= dbConfirmations && needPushUtxo {
							//补推
							//通知接收方
							item_str, _ := json.Marshal(item_arr)
							redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, res.PlatformId)))
						}

					}
				}
			}

			////如果是挖矿收入
			//if isBase && res.Type != 0 {
			//	//通知接收方
			//	redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, res.PlatformId)))
			//}

		}
	}

	//开始处理合约数据，双币种模型
	if contract != nil {
		payCoinName = contract.Coin
		reb := make([]map[string]interface{}, 0)
		if strings.ToLower(transData.CoinName) == "qtum" {
			reb = parsingContract(contract, &transData, txid, blockhash, redisHelper, fee)
		} else {
			reb = parsingContract(contract, &transData, txid, blockhash, redisHelper, "")
		}
		if reb != nil {
			for _, vv := range reb {
				push_data = append(push_data, vv)
			}
		}
	}

	item_dump["push_data"] = push_data
	trans_push := []map[string]interface{}{}
	if push_data != nil {
		for _, vv := range push_data {
			jump := map[string]interface{}{}

			jump["coin_type"] = vv["coin_type"]
			jump["is_in"] = vv["dir"]
			jump["block_height"] = transData.Height
			jump["timestamp"] = transData.Time
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
			jump["confirmations"] = transData.Confirmations
			jump["app_id"] = vv["mch_id"]
			jump["user_sub_id"] = vv["mch_id"]
			jump["type"] = vv["addr_type"]
			jump["push_state"] = 1
			trans_push = append(trans_push, jump)
		}
	}
	item_dump["trans_push"] = trans_push

	tx_in_amount := ""
	tx_in_num := 0
	if len(in_list) > 0 {
		tx_in_amount = bcadd_multi_utxoinput(func(val interface{}) string { _tx := val.(model.PushTxInput); return _tx.Value }, in_list)
		tx_in_num = len(in_list)
	}

	tx_out_amount := ""
	tx_out_num := 0
	if len(out_list) > 0 {
		tx_out_amount = bcadd_multi_utxooutput(
			func(val interface{}) string { _tx := val.(model.PushTxOutput); return _tx.Value }, out_list)
		tx_out_num = len(out_list)
	}
	if payCoinName != "" {
		data["coin_type"] = payCoinName
	} else {
		data["coin_type"] = coinName
	}
	data["tx_in_amount"] = tx_in_amount
	data["tx_in_num"] = tx_in_num
	data["tx_out_amount"] = tx_out_amount
	data["tx_out_num"] = tx_out_num
	item_dump["clear"] = data
	if item_dump != nil {
		dump_str, _ := json.Marshal(item_dump)
		//log.Debug(string(dump_str))
		//1次
		redisHelper.LeftPush("import_list_new", string(dump_str))
	}

	return nil
}

//解析合约数据
func parsingContract(trans *model.PushContractTx, trans_data *model.PushUtxoBlockInfo, tx_id string, hash string, redisHelper *util.RedisClient, fixFeeFloat string) []map[string]interface{} {
	var (
		coinName  string
		tokenName string
		jumpData  []map[string]interface{}
		fromMchId int
		toMchId   int
	)

	jumpData = make([]map[string]interface{}, 0)
	coinName = strings.ToLower(trans_data.CoinName)

	//usdt有无效数据
	if coinName == "btc" && !trans.Valid {
		//usdt 无效交易
		return nil
	}
	if coinName == "qtum" {
		trans.Coin = "qtum"
	}
	if trans.Coin != "" {
		tokenName = trans.Coin
	}

	if trans.Contract != "" {
		coinset := &entity.FcCoinSet{}
		result, err := dao.TransPushGet(coinset, "select name from fc_coin_set where token = ?", trans.Contract)
		if !result || err != nil {
			log.Debug(err)
			return nil
		}
		tokenName = coinset.Name
	}
	if tokenName == "" {
		//无法查询代币信息，直接忽略
		return nil
	}

	if trans.From != "" {
		address := trans.From
		amount := trans.Amount
		temp := map[string]interface{}{}
		res := &entity.FcGenerateAddressList{}
		result, err := dao.TransPushGet(res, "select id,type,platform_id from fc_generate_address_list where address = ?", address)
		if !result || err != nil {
			res.PlatformId = 0
			res.Type = 0
		}
		fromMchId = res.PlatformId
		temp["mch_id"] = res.PlatformId
		temp["coin_type"] = tokenName
		temp["tx_id"] = tx_id
		temp["hash"] = hash
		temp["dir"] = 2
		temp["tx_n"] = 0
		temp["addr"] = address
		temp["amount"] = amount
		temp["addr_type"] = res.Type
		temp["from_tx_id"] = ""
		temp["is_spent"] = 0
		temp["create_at"] = time.Now().Unix()
		jumpData = append(jumpData, temp)
	}

	if trans.To != "" {
		address := trans.To
		amount := trans.Amount
		temp := map[string]interface{}{}
		res := &entity.FcGenerateAddressList{}
		//result, err := dao.TransPushGet(address, "select type,platform_id from fc_generate_address_list where address = ? and coin_name = ?", res)
		//查找主链币地址
		result, err := dao.TransPushGet(res, "select type,platform_id from fc_generate_address_list where address = ? and coin_name = ?", address, trans.Coin)
		if !result || err != nil {
			res.PlatformId = 0
			res.Type = 0
		}
		toMchId = res.PlatformId
		temp["mch_id"] = res.PlatformId
		temp["coin_type"] = tokenName
		temp["tx_id"] = tx_id
		temp["hash"] = hash
		temp["dir"] = 1
		temp["tx_n"] = 0
		temp["addr"] = address
		temp["amount"] = amount
		temp["addr_type"] = res.Type
		temp["from_tx_id"] = ""
		temp["is_spent"] = 0
		temp["create_at"] = time.Now().Unix()
		jumpData = append(jumpData, temp)
		//if res.Type == 5 && lower_coin_name == "usdt" {
		//	mch_recharge(temp)
		//}
		fee := ""
		switch trans.Fee.(type) {
		case float64:
			fee = decimal.NewFromFloat(trans.Fee.(float64)).String()
		case int64:
			fee = decimal.NewFromInt(trans.Fee.(int64)).String()
		case string:
			fee = trans.Fee.(string)
		}

		if coinName == "qtum" {
			fee = fixFeeFloat
		}

		item_arr := map[string]interface{}{}
		item_arr["coin"] = coinName
		item_arr["coin_type"] = tokenName
		item_arr["is_in"] = 1
		item_arr["block_height"] = trans_data.Height
		item_arr["timestamp"] = trans_data.Time
		item_arr["transaction_id"] = tx_id
		item_arr["txid"] = tx_id
		item_arr["trx_n"] = 0
		item_arr["from_address"] = trans.From
		item_arr["to_address"] = address
		item_arr["amount"] = amount
		item_arr["fee"] = fee
		item_arr["fee_coin"] = coinName
		item_arr["confirmations"] = trans_data.Confirmations
		item_arr["confirm_time"] = trans_data.Time
		item_arr["user_sub_id"] = toMchId
		item_arr["from_trx_id"] = ""
		item_arr["memo"] = ""
		if trans.Memo != "" {
			item_arr["memo"] = trans.Memo
		}
		item_arr["contract_address"] = ""
		if trans.Contract != "" {
			item_arr["contract_address"] = trans.Contract
		}
		//item_str, _ := json.Marshal(item_arr)

		dataConfirmations := trans_data.Confirmations
		if dataConfirmations > 6 {
			dataConfirmations = 6
		}
		//如果是出账,
		if res.Type == 0 {
			if fromMchId != 0 {
				//查询push表记录目前最大的确认数，例如：如果确认数为4  db不存在记录 则 补充1-4
				var dbConfirmations int64 = 0
				record, err := dao.FcPushRecordLast(tx_id, coinName, tokenName, fromMchId)
				if err != nil {
					log.Debugf("查询push表记录目前最大的确认数,设置为0，异常，error=[%s]", err.Error())
				}
				if record != nil {
					dbConfirmations = int64(record.Confirmations)
				}
				//基于数据库最大存入的确认数，补全丢失记录，例如推送12 db记录4 补全 4-12，如果已经是12 再推送一次12
				if dataConfirmations > dbConfirmations {
					for i := dbConfirmations; i < dataConfirmations; i++ {
						item_arr["confirmations"] = i + 1
						item_str, _ := json.Marshal(item_arr)
						//通知发送方
						redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, fromMchId)))
					}
				} else if dataConfirmations == dbConfirmations {
					//通知发送方
					item_str, _ := json.Marshal(item_arr)
					redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, fromMchId)))
				}

			}
		} else {
			if res.PlatformId == fromMchId && res.Type != 3 {
				log.Infof("币种=[%s],代币[%s],txid=[%s],其中from和to地址[%s],为同一商户 [%d] 的内部地址,无需推送", coinName, tokenName, tx_id, trans.To, fromMchId)
			} else {
				if fromMchId != 0 {

					//查询push表记录目前最大的确认数，例如：如果确认数为4  db不存在记录 则 补充1-4
					var dbConfirmations int64 = 0
					record, err := dao.FcPushRecordLast(tx_id, coinName, tokenName, fromMchId)
					if err != nil {
						log.Debugf("查询push表记录目前最大的确认数,设置为0，异常，error=[%s]", err.Error())
					}
					if record != nil {
						dbConfirmations = int64(record.Confirmations)
					}
					//基于数据库最大存入的确认数，补全丢失记录，例如推送12 db记录4 补全 4-12，如果已经是12 再推送一次12
					if dataConfirmations > dbConfirmations {
						for i := dbConfirmations; i < dataConfirmations; i++ {
							item_arr["confirmations"] = i + 1
							item_str, _ := json.Marshal(item_arr)
							//通知发送方
							redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, fromMchId)))
						}
					} else if dataConfirmations == dbConfirmations {
						//补推
						//通知发送方
						item_str, _ := json.Marshal(item_arr)
						redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 2, fromMchId)))
					}
				}

				if fromMchId == 0 && res.Type == 1 {
					return nil
				}

				//查询push表记录目前最大的确认数，例如：如果确认数为4  db不存在记录 则 补充1-4
				var dbConfirmations int64 = 0
				record, err := dao.FcPushRecordLast(tx_id, coinName, tokenName, res.PlatformId)
				if err != nil {
					log.Debugf("查询push表记录目前最大的确认数,设置为0，异常，error=[%s]", err.Error())
				}
				if record != nil {
					dbConfirmations = int64(record.Confirmations)
				}
				//基于数据库最大存入的确认数，补全丢失记录，例如推送12 db记录4 补全 4-12，如果已经是12 再推送一次12
				if dataConfirmations > dbConfirmations {
					for i := dbConfirmations; i < dataConfirmations; i++ {
						item_arr["confirmations"] = i + 1
						item_str, _ := json.Marshal(item_arr)
						//通知接收方
						redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, res.PlatformId)))
					}
				} else if dataConfirmations == dbConfirmations {
					//补推
					//通知接收方
					item_str, _ := json.Marshal(item_arr)
					redisHelper.LeftPush("notice_list_new", string(format_decimal(item_str, 1, res.PlatformId)))
				}
			}
		}
	}
	return jumpData
}
