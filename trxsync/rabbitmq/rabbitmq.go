package rabbitmq

import (
	"encoding/json"
	"fmt"
	"github.com/group-coldwallet/trxsync/common"
	"github.com/group-coldwallet/trxsync/common/log"
	"github.com/group-coldwallet/trxsync/conf"
	"github.com/group-coldwallet/trxsync/models/bo"
	"github.com/group-coldwallet/trxsync/models/po"
)

type AddrInfo struct {
	AddressInfo  bo.UserAddressInfo `json:"addressInfo"`
	ContractInfo po.ContractInfo    `json:"contractInfo"`
}

func initConsumerabbitmq(conf *conf.Config) *RabbitPool {

	instanceConsumePool := NewConsumePool()
	//instanceConsumePool.SetMaxConsumeChannel(100)
	err := instanceConsumePool.Connect("amqps", conf.Mq.HostPort, conf.Mq.Username, conf.Mq.Password)
	if err != nil {
		fmt.Println(err)
	}

	return instanceConsumePool
}
func Consume(watch *common.WatchControl, conf *conf.Config, isAddr bool) {
	instanceConsumePool := initConsumerabbitmq(conf)
	queueName := "iota_addr"
	if !isAddr {
		queueName = "iota_contract"
	}
	nomrl := &ConsumeReceive{
		// 定义消费者事件
		ExchangeName: queueName, //队列名称
		ExchangeType: EXCHANGE_TYPE_FANOUT,
		Route:        queueName,
		QueueName:    queueName,
		IsTry:        true, //是否重试
		MaxReTry:     5,    //最大重试次数
		EventFail: func(code int, e error, data []byte) {
			fmt.Printf("error:%s", e)
		},

		/***
		 * 参数说明
		 * @param data []byte 接收的rabbitmq数据
		 * @param header map[string]interface{} 原rabbitmq header
		 * @param retryClient Rabbitrabbitmq.RetryClientInterface 自定义重试数据接口，重试需return true 防止数据重复提交
		 ***/
		EventSuccess: func(data []byte, header map[string]interface{}, retryClient RetryClientInterface) bool {
			//如果返回true 则无需重试
			fmt.Printf("data:%s\n", string(data))
			if isAddr {
				addrInfos := make([]bo.UserAddressInfo, 0)
				err := json.Unmarshal(data, &addrInfos)
				if err != nil {
					log.Error(err)
					return false
				}
				for _, addrInfo := range addrInfos {
					watch.InsertWatchAddress(addrInfo.UserID, addrInfo.Address, addrInfo.NotifyUrl)
				}

			} else {
				contractInfos := make([]po.ContractInfo, 0)
				err := json.Unmarshal(data, &contractInfos)
				if err != nil {
					log.Error(err)
					return false
				}
				//name, contractaddr, cointype string, decimal int
				for _, contractInfo := range contractInfos {
					watch.InsertWatchContract(contractInfo.Name, contractInfo.ContractAddress, contractInfo.CoinType, contractInfo.Decimal)
				}
			}

			return true
		},
	}

	instanceConsumePool.RegisterConsumeReceive(nomrl)
	err := instanceConsumePool.RunConsume()
	if err != nil {
		log.Error(err)
	}
}
