package demo

import (
	"fmt"
	"gitee.com/tym_hmm/rabbitmq-pool-go"
	"sync"
	"testing"
)

//func init() {
//	// 初始化
//	initrabbitmq()
//}

func TestInitrabbitmq(t *testing.T) {
	initrabbitmq()
	rund()
}

var oncePool sync.Once
var instanceRPool *kelleyRabbimqPool.RabbitPool

func initrabbitmq() *kelleyRabbimqPool.RabbitPool {
	oncePool.Do(func() {
		instanceRPool = kelleyRabbimqPool.NewProductPool()
		err := instanceRPool.Connect("127.0.0.1", 5672, "guest", "guest")
		if err != nil {
			fmt.Println(err)
		}
	})
	return instanceRPool
}

func rund() {

	var wg sync.WaitGroup

	//wg.Add(1)
	//go func() {
	//	fmt.Println("aaaaaaaaaaaaaaaaaaaaaa")
	//	defer wg.Done()
	//	runtime.SetMutexProfileFraction(1)  // 开启对锁调用的跟踪
	//	runtime.SetBlockProfileRate(1)      // 开启对阻塞操作的跟踪
	//	err:= http.ListenAndServe("0.0.0.0:8080", nil)
	//	fmt.Println(err)
	//}()

	for i := 0; i < 100000; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			data := kelleyRabbimqPool.GetRabbitMqDataFormat("test", kelleyRabbimqPool.EXCHANGE_TYPE_FANOUT, "test", "test", fmt.Sprintf("这里是数据%d", num))
			err := instanceRPool.Push(data)
			if err != nil {
				fmt.Println(err)
			}
		}(i)
	}

	wg.Wait()

}
