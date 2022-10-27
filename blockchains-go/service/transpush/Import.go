package transpush

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/go-xorm/builder"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"github.com/group-coldwallet/blockchains-go/runtime/global"
	"github.com/shopspring/decimal"
)

type ExecImport struct {
	// nothing
}

var arrExclude = []string{"iotx", "sol", "btc", "eth", "etc", "eos",
	"trx", "heco", "hsc", "bsc", "zec", "ltc", "bnb", "wtc", "moac",
	"kai", "rbtc", "movr", "sep20", "optim", "brise-brise", "ftm",
	"welups", "rose", "one", "rev", "tkm", "ron", "neo", "icp", "flow",
	"uenc", "btm", "cspr", "matic-matic", "rei", "evmos", "aur", "dscc",
	"mob", "dscc1", "lat", "deso", "nodle", "hbar"}

func (r *ExecImport) Run(reqexit <-chan bool) {
	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		log.Error(err)
		WaitGroupTransPush.Done()
		return
	}
	defer redisHelper.Close()
	log.Debug("Run ExecImport")
	run := true
	for run {
		select {
		case s := <-reqexit:
			log.Error("ExecImport exit", s)
			run = false
			break

		default:
			var tmpdata []byte
			var err error
			var result = false
			item, err := redisHelper.Rpop("import_list_new")
			//log.Info("Rpop import_list_new")
			if err != nil {
				if strings.Contains(err.Error(), "nil") {
					time.Sleep(time.Second * 1)
					break
				}
				log.Error(err)
				time.Sleep(time.Second * 1)
				break
			}
			for {
				if item == nil {
					break
				}

				//log.Debug(string(item))
				var trans_data map[string]interface{}
				err = json.Unmarshal(item, &trans_data)
				if err != nil {
					log.Error(err)
					break
				}

				session := dao.TransPushGetSession()
				if session == nil {
					break
				}
				defer session.Close()
				session.Begin()

				datatype := int32(trans_data["type"].(float64))
				if datatype != 0 && datatype != 10 && datatype != 30 && datatype != 40 && datatype != 50 {
					redisHelper.LeftPush("confirm_list_new", string(item))
					break
				}

				data := trans_data["clear"]
				if data == nil {
					break
				}
				_data := data.(map[string]interface{})
				if _data["tx_fee"] != nil {
					switch _data["tx_fee"].(type) {
					case float64:
						_data["tx_fee"] = decimal.NewFromFloat(_data["tx_fee"].(float64)).String()
					case int64:
						_data["tx_fee"] = decimal.NewFromInt(_data["tx_fee"].(int64)).String()
					}
				}

				tmpdata, err = json.Marshal(_data)
				if err != nil {
					log.Error(err)
					break
				}

				txclear := &entity.FcTxClear{}
				err = json.Unmarshal(tmpdata, txclear)
				if err != nil {
					log.Error(err)
					break
				}
				//txclear.UpdateAt = time.Now().Unix()

				//todo 波卡类型存在批量交易，暂时先硬编码处理，后续更改结构
				//isPolk := false //波卡类型
				txc_id := int64(0)
				if in_array(txclear.CoinType, []interface{}{"dot", "azero", "sgb-sgb", "kar", "dhx", "nodle"}) {
					log.Infof("Import dao.FcTxClearFindList %s", txclear.TxId)
					tcs, _, err := dao.FcTxClearFindList(txclear.CoinType, txclear.TxId)
					log.Infof("Import dao.FcTxClearFindList 完成 %s", txclear.TxId)
					if err != nil {
						log.Error()
					}
					//isPolk = true
					if len(tcs) > 0 {
						txc_id = int64(tcs[0].Id)
					}
				}

				if txc_id == 0 {
					var insert_data builder.Eq
					_tmp, _ := json.Marshal(txclear)
					json.Unmarshal(_tmp, &insert_data)
					log.Infof("Import dao.insert fc_txclear %s", txclear.TxId)
					sql, err := builder.MySQL().Insert(insert_data).Into("fc_tx_clear").ToBoundSQL()
					log.Infof("Import dao.insert fc_txclear 完成 %s", txclear.TxId)

					if err != nil {
						log.Error(err)
						break
					}
					//log.Debug(sql)
					sqlRes, err := session.Exec(sql)
					if err != nil {
						log.Error(err)
						session.Rollback()
						if strings.Contains(err.Error(), "Duplicate") {
							result = true
						}
						break
					}
					txc_id, err = sqlRes.LastInsertId()
					if err != nil {
						log.Error(err)
						session.Rollback()
						break
					}
				}

				log.Debug(txc_id)

				if trans_data["push_data"] != nil {
					push_data := trans_data["push_data"].([]interface{})
					for _, _vv := range push_data {
						vv := _vv.(map[string]interface{})
						vv["txc_id"] = txc_id

						tmpdata, err = json.Marshal(vv)
						if err != nil {
							log.Error(err)
							break
						}
						txcleardetail := &entity.FcTxClearDetail{}
						err = json.Unmarshal(tmpdata, txcleardetail)
						if err != nil {
							log.Error(err)
							err = errors.New("json error")
							result = true //过滤掉json错误。不然会反复出错
							break
						}
						log.Infof("Import dao.insert fc_txclear_detail %s", txcleardetail.TxId)
						txcleardetail.UpdateAt = time.Now()
						insertid, err := session.InsertOne(txcleardetail)
						log.Infof("Import dao.insert fc_txclear_detail 完成 %s", txcleardetail.TxId)
						log.Debug(insertid)
						if err != nil {
							log.Error(err)
							break
						}
					}

					if err != nil {
						session.Rollback()
						break
					}
				}

				if trans_data["trans_push"] != nil {
					trans_data := trans_data["trans_push"].([]interface{})

					for _, _vv := range trans_data {
						vv := _vv.(map[string]interface{})
						if vv["fee"] != nil {
							switch vv["fee"].(type) {
							case float64:
								vv["fee"] = decimal.NewFromFloat(vv["fee"].(float64)).String()
							case int64:
								vv["fee"] = decimal.NewFromInt(vv["fee"].(int64)).String()
							}
						}
						tmpdata, err = json.Marshal(vv)
						if err != nil {
							log.Error(err)
							break
						}
						trans := &entity.FcTransPush{}
						err = json.Unmarshal(tmpdata, trans)
						if err != nil {
							log.Error(err)
							break
						}
						//前期先写死代币
						//eth 编号5，
						//eos 编号8
						//heco，826
						//bsc 640
						//hsc
						if global.CoinDecimal[trans.CoinType].Pid == 5 || global.CoinDecimal[trans.CoinType].Pid == 8 ||
							global.CoinDecimal[trans.CoinType].Pid == 826 || global.CoinDecimal[trans.CoinType].Pid == 640 ||
							global.CoinDecimal[trans.CoinType].Pid == 1012 || global.CoinDecimal[trans.CoinType].Pid == 37 ||
							global.CoinDecimal[trans.CoinType].Pid == 349 || util.IsInArrayStr(trans.CoinType, arrExclude) {
							log.Infof("币种：%s,无需写入trans_push", trans.CoinType)
						} else {
							log.Info("Import dao.insert trans_push")
							insertid, err := session.InsertOne(trans)
							log.Info("Import dao.insert trans_push 完成")
							log.Debug(insertid, err)
							if err != nil {
								log.Error(err)
								break
							}
						}
					}

					if err != nil {
						session.Rollback()
						break
					}
				}

				session.Commit()
				result = true
				break
			}

			if !result {
				redisHelper.LeftPush("import_list_new", string(item))
			}

			time.Sleep(time.Millisecond * 60)
			break
		}
	}

	WaitGroupTransPush.Done()
}
