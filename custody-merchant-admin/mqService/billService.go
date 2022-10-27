package mqService

import (
	"custody-merchant-admin/internal/domain"
	"custody-merchant-admin/internal/service"
	"custody-merchant-admin/middleware/rabbitmq"
	"custody-merchant-admin/module/log"
	"encoding/json"
	"fmt"
)

// BillConsume
// 账单
func BillConsume() {
	nomrl := ConsumeReceiveFunc("m_data",
		"m_data",
		"m_data",
		"direct",
		func(code int, e error, data []byte) {
			// 错误处理
			log.Errorf("账单，MQ接收消息错误 code：%d; error：%v; data：%v \n", code, e.Error(), string(data))
			//fmt.Printf("账单，MQ接收消息错误 data：%v \n", e.Error())
		},
		func(data []byte, header map[string]interface{}, retryClient rabbitmq.RetryClientInterface) bool {
			// 正确处理
			for key, value := range header {
				log.Errorf("账单，MQ接收消息 header：%s=>%s \n", key, value)
			}
			mp := map[string]interface{}{}
			err := json.Unmarshal(data, &mp)
			if err != nil {
				fmt.Println(err.Error())
				log.Error(err.Error())
				return false
			}
			msg := ""
			if v, ok := mp["msg"]; ok {
				msg = v.(string)
				fmt.Printf("%s,data:%v", msg, string(data))
				log.Infof(" MQ接收消息 data：%v", msg, string(data))
			} else {
				fmt.Println("缺少msg参数")
				log.Error("缺少msg参数")
				return false
			}

			if v, ok := mp["type"]; ok {
				billData := domain.BillInfo{}
				b, err := json.Marshal(mp["data"])
				if err != nil {
					log.Error(err.Error())
					return false
				}
				json.Unmarshal(b, &billData)
				params := mp["params"].(map[string]interface{})
				switch v {
				case "re_push":
					// TODO 重推,这里写商户回调逻辑
					if billData.SerialNo == "" {
						log.Error("重推失败，订单号 为null")
						return false
					}
					err = service.PushDataByUrl(billData.SerialNo)
					if err != nil {
						log.Error(err.Error())
						return false
					}
					break
				case "withdrawal":
					// TODO 提现,这里写提现审核逻辑
					billStatus := params["bill_status"].(float64)
					err = service.MerchantWithdrawal(billData, int(billStatus))
					if err != nil {
						return false
					}
				}
			}
			return true
		},
	)
	instanceConsumePool.RegisterConsumeReceive(nomrl)
	err := instanceConsumePool.RunConsume()
	if err != nil {
		log.Errorf("开启MQ，消费失败：%v", err.Error())
		fmt.Println(err)
	}
}
