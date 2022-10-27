package transpush

import (
	"encoding/json"
	"github.com/group-coldwallet/blockchains-go/dao"
	"github.com/group-coldwallet/blockchains-go/entity"
	"github.com/group-coldwallet/blockchains-go/log"
	"github.com/group-coldwallet/blockchains-go/pkg/util"
	"strings"
	"time"
)

// 确认数推送
type ExecConfirm struct {
	// nothing
}

func (r *ExecConfirm) Run(reqexit <-chan bool) {
	redisHelper, err := util.AllocRedisClient()
	if err != nil {
		log.Error(err)
		WaitGroupTransPush.Done()
		return
	}
	defer redisHelper.Close()
	log.Debug("Run ExecConfirm")
	run := true
	for run {
		select {
		case s := <-reqexit:
			log.Error("ExecConfirm exit", s)
			run = false
			break

		default:
			var err error
			var result = false
			redisHelper.Rpop()
			item, err := redisHelper.Rpop("confirm_list_new")
			//lenConfirm, _ := redisHelper.LLEN("confirm_list_new")
			//log.Infof("confirm_list_new 长度:%d", lenConfirm)
			if err != nil {
				if strings.Contains(err.Error(), "nil") {
					time.Sleep(time.Second * 1)
					break
				}
				log.Debug(err)
				time.Sleep(time.Second * 1)
				break
			}
			for {
				if item == nil {
					break
				}
				log.Debug(string(item))
				var trans_data map[string]interface{}
				err = json.Unmarshal(item, &trans_data)
				if err != nil {
					log.Debug(err)
					break
				}

				datatype := int32(trans_data["type"].(float64))
				if datatype != 1 && datatype != 11 && datatype != 31 && datatype != 41 {
					log.Error("datatype error")
					break
				}

				height := int64(trans_data["height"].(float64))
				hash := trans_data["hash"].(string)
				confirmations := int64(trans_data["confirmations"].(float64))
				coin_type := strings.ToLower(trans_data["coin"].(string))
				if coin_type == "btc-stx" {
					result = true
					break
				}

				coin_type_range := []interface{}{}
				if in_array(coin_type, []interface{}{"rub", "eth", "eos", "nas", "qtum", "hc"}) {
					coinset := &entity.FcCoinSet{}
					tmpresult, err := dao.TransPushGet(coinset, "select id from fc_coin_set where name = ? and pid = 0", coin_type)
					if !tmpresult || err != nil {
						log.Error(err)
						break
					}
					rub_id := coinset.Id

					everyone := make([]*entity.FcCoinSet, 0)
					err = dao.TransPushFind(&everyone, "select name from fc_coin_set where pid = ?", rub_id)
					if err != nil {
						log.Error(err)
						break
					}
					for _, v := range everyone {
						coin_type_range = append(coin_type_range, v.Name)
					}
					coin_type_range = append(coin_type_range, coin_type)
				} else if coin_type == "btc" {
					coin_type_range = append(coin_type_range, coin_type)
					coin_type_range = append(coin_type_range, "usdt")
				} else {
					coin_type_range = append(coin_type_range, coin_type)
				}

				//$list = Db::name("fc_tx_clear")->where("coin", $coin_type)->where(['block_height'=>$height, 'hash'=>$hash])->select();
				list := make([]*entity.FcTxClear, 0)
				err = dao.TransPushFind(&list, "select * from fc_tx_clear where coin = ? and block_height = ? and hash = ?", coin_type, height, hash)
				if err != nil {
					log.Error(err)
					break
				}

				if len(list) == 0 {
					if int64(trans_data["time"].(float64))+720 > time.Now().UTC().Unix() {
						if coin_type != "eth" {
							redisHelper.LeftPush("confirm_list_new", string(item))
						}
					}
					result = true
					log.Debug("还未写入数据", string(item))
					break
				}

				//yxc2019-09-17 开始
				_coinset := &entity.FcCoinSet{}
				tmpresult, err := dao.TransPushGet(_coinset, "select contrast_confirm from fc_coin_set where name = ?", coin_type)
				if !tmpresult || err != nil {
					log.Error(err)
					break
				}

				session := dao.TransPushGetSession()
				if session == nil {
					break
				}
				defer session.Close()
				session.Begin()

				contrast_confirm := int64(_coinset.ContrastConfirm)
				if contrast_confirm <= confirmations {
					contrast_time := time.Now().Unix()
					_, err := session.Exec("update fc_tx_clear set contrast_time = ? where coin = ? and block_height = ? and hash = ? and contrast_time = 0", contrast_time, coin_type, height, hash)
					if err != nil {
						log.Error(err)
						session.Rollback()
						break
					}
				}

				log.Debug(string(item), "_ok")
				session.Exec("update fc_tx_clear set confirmations = ? where coin = ? and block_height = ? and hash = ?", confirmations, coin_type, height, hash)
				session.In("coin_type", coin_type_range).Exec("update trans_push set confirmations = ? where block_height = ? and hash = ?", confirmations, height, hash)

				usdtTxids := make(map[string]struct{})
				for _, info := range list {
					if strings.ToLower(info.CoinType) == "usdt" {
						usdtTxids[info.TxId] = struct{}{}
						break
					}
				}

				for _, ko := range list {
					if _, ok := usdtTxids[ko.TxId]; ok && strings.ToLower(ko.CoinType) == "btc" {
						//存在usdt ,抛弃btc发送
						continue
					}
					m_list := make([]*entity.FcPushRecord, 0)
					log.Infof("确认数执行币种：%s", ko.CoinType)
					if strings.ToLower(ko.CoinType) == "usdt" {
						err = session.SQL("select app_id,coin_type,msg,url,tx_id,confirmations,status from fc_push_record2 where tx_id = ? and coin_type = ? group by msg", ko.TxId, ko.CoinType).Find(&m_list)
					} else {
						err = session.SQL("select app_id,coin_type,msg,url,tx_id,confirmations,status from fc_push_record2 where tx_id = ? group by msg", ko.TxId).Find(&m_list)
					}
					if err != nil {
						log.Error(err)
						break
					}
					temp_arr := map[string][]interface{}{}
					if len(m_list) == 0 {
						log.Debug("无推送记录")
						continue
					}

					add := false
					for _, vv := range m_list {
						//fix usdt需要做暂时切换处理
						//if strings.ToLower(vv.CoinType) == "usdt" {
						//	//暂时抛弃usdt数据，确认数
						//	continue
						//}
						var data map[string]interface{}
						json.Unmarshal([]byte(vv.Msg), &data)
						delete(data, "confirmations")
						if _, ok := data["confirm_time"]; ok {
							delete(data, "confirm_time")
						}
						temp, _ := json.Marshal(data)
						if !add {
							temp_arr[string(temp)] = make([]interface{}, 0)
							add = true
						}
						temp_arr[string(temp)] = append(temp_arr[string(temp)], vv)
					}

					if len(temp_arr) > 0 {
						for kk, _ := range temp_arr {
							var temp map[string]interface{}
							json.Unmarshal([]byte(kk), &temp)
							temp["confirmations"] = confirmations
							temp["confirm_time"] = time.Now().Unix()
							json_str, _ := json.Marshal(temp)
							//切换币种临时限制
							err := redisHelper.LeftPush("notice_list_new", string(json_str))
							if err == nil {
								log.Debug("更新确认数成功", string(item))
							} else {
								log.Debug("更新确认数失败", string(item))
							}
						}
					}
				}

				session.Commit()
				result = true
				break
			}
			//
			if !result {
				//redisHelper.LeftPush("confirm_list_new", string(item))
			}

			time.Sleep(time.Millisecond * 60)
			break
		}
	}

	WaitGroupTransPush.Done()
}
