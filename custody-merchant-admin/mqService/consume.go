package mqService

import (
	. "custody-merchant-admin/config"
	"custody-merchant-admin/middleware/rabbitmq"
	"fmt"
	"sync"
)

var (
	onceConsumePool     sync.Once
	instanceConsumePool *rabbitmq.RabbitPool
)

func initConsumerabbitmq() *rabbitmq.RabbitPool {
	onceConsumePool.Do(func() {
		instanceConsumePool = rabbitmq.NewConsumePool()
		//instanceConsumePool.SetMaxConsumeChannel(100)
		err := instanceConsumePool.Connect(Conf.RabbitMQ.Prefix, Conf.RabbitMQ.MQUrl, Conf.RabbitMQ.MQUser, Conf.RabbitMQ.MQPassword)
		if err != nil {
			fmt.Println(err)
		}
	})
	return instanceConsumePool
}

func ConsumeReceiveFunc(exchangeName, route, queueName, exchangeType string,
	callFail func(code int, e error, data []byte),
	callSuccess func(data []byte, header map[string]interface{}, retryClient rabbitmq.RetryClientInterface) bool) *rabbitmq.ConsumeReceive {

	return &rabbitmq.ConsumeReceive{
		// 定义消费者事件
		ExchangeName: exchangeName, //队列名称
		ExchangeType: exchangeType,
		Route:        route,
		QueueName:    queueName,
		IsTry:        true,                  //是否重试
		MaxReTry:     Conf.RabbitMQ.Reconns, //最大重试次数
		EventFail:    callFail,

		/***
		 * 参数说明
		 * @param data []byte 接收的rabbitmq数据
		 * @param header map[string]interface{} 原rabbitmq header
		 * @param retryClient RabbitmqPool.RetryClientInterface 自定义重试数据接口，重试需return true 防止数据重复提交
		 ***/
		EventSuccess: callSuccess,
	}
}
