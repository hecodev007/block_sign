package mqpool

import (
	"fmt"
	"gitee.com/tym_hmm/rabbitmq-pool-go"
	"sync"
)

func init() {

}

func Consume() {
	nomrl := &kelleyRabbimqPool.ConsumeReceive{
		ExchangeName: "testChange31", //队列名称
		ExchangeType: kelleyRabbimqPool.EXCHANGE_TYPE_DIRECT,
		Route:        "",
		QueueName:    "testQueue31",
		IsTry:        true, //是否重试
		MaxReTry:     5,    //最大重试次数
		EventFail: func(code int, e error, data []byte) {
			fmt.Printf("error:%s", e)
		},
		EventSuccess: func(data []byte, header map[string]interface{}, retryClient kelleyRabbimqPool.RetryClientInterface) bool { //如果返回true 则无需重试
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
		err := instanceConsumePool.Connect("192.168.1.80", 5672, "fnadmin", "Fn123456")
		if err != nil {
			fmt.Println(err)
		}
	})
	return instanceConsumePool
}
