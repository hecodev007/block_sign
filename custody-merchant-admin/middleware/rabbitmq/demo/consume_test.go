package demo

import (
	"fmt"
	"gitee.com/tym_hmm/rabbitmq-pool-go"
	"sync"
	"testing"
)

func TestConsume(t *testing.T) {
	// 初始化
	initConsumerabbitmq()
	Consume()
}

func Consume() {

	nomrl := &kelleyRabbimqPool.ConsumeReceive{
		// 定义消费者事件
		ExchangeName: "test", //队列名称
		ExchangeType: kelleyRabbimqPool.EXCHANGE_TYPE_FANOUT,
		Route:        "test",
		QueueName:    "test",
		IsTry:        true, //是否重试
		MaxReTry:     5,    //最大重试次数
		EventFail: func(code int, e error, data []byte) {
			fmt.Printf("error:%s", e)
		},

		/***
		 * 参数说明
		 * @param data []byte 接收的rabbitmq数据
		 * @param header map[string]interface{} 原rabbitmq header
		 * @param retryClient RabbitmqPool.RetryClientInterface 自定义重试数据接口，重试需return true 防止数据重复提交
		 ***/
		EventSuccess: func(data []byte, header map[string]interface{}, retryClient kelleyRabbimqPool.RetryClientInterface) bool {
			//如果返回true 则无需重试
			fmt.Printf("data:%s\n", string(data))
			return true
		},
	}

	instanceConsumePool.RegisterConsumeReceive(nomrl)
	err := instanceConsumePool.RunConsume()
	if err != nil {
		fmt.Println(err)
	}
}

var onceConsumePool sync.Once

var instanceConsumePool *kelleyRabbimqPool.RabbitPool

func initConsumerabbitmq() *kelleyRabbimqPool.RabbitPool {
	onceConsumePool.Do(func() {
		instanceConsumePool = kelleyRabbimqPool.NewConsumePool()
		//instanceConsumePool.SetMaxConsumeChannel(100)
		err := instanceConsumePool.Connect("127.0.0.1", 5672, "guest", "guest")
		if err != nil {
			fmt.Println(err)
		}
	})
	return instanceConsumePool
}
