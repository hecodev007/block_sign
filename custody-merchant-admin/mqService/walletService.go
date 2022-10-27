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
func WalletConsume() {
	nomrl := ConsumeReceiveFunc("wallet_send_data",
		"wallet_send_data",
		"wallet_send_data",
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
			if _, ok := mp["data"]; ok {
				// 判断是充值还是提现
				// 充值
				if true {
					// 判断是否为未确认
					// 冻结金额
					if true {
						// 未确认
						service.ReceiveBillDetail(domain.MqWalletInfo{})
					} else {
						// 已经确认
						err := service.UpdateReceiveBillDetail("1", 1)
						if err != nil {
							return false
						}
					}
				}
				// 提现
				if true {
					service.FreezeBillDetailState(domain.MqWalletInfo{}, 1)
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
