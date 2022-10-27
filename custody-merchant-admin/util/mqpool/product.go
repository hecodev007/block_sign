package mqpool

import (
	"fmt"
	"gitee.com/tym_hmm/rabbitmq-pool-go"
	"net/http"
	"runtime"
	"sync"
)

func init() {
	// 初始化
	initrabbitmq()
}

var oncePool sync.Once
var instanceRPool *kelleyRabbimqPool.RabbitPool

func initrabbitmq() *kelleyRabbimqPool.RabbitPool {
	oncePool.Do(func() {
		instanceRPool = kelleyRabbimqPool.NewProductPool()
		err := instanceRPool.Connect("b-59fecdb5-5f87-4fd4-b099-121f293555ea.mq.ap-northeast-1.amazonaws.com", 5671, "hoocustody", "#M12$59gjegC8tza4Eg1bgUV")
		if err != nil {
			fmt.Println(err)
		}
	})
	return instanceRPool
}

func rund() {

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		fmt.Println("aaaaaaaaaaaaaaaaaaaaaa")
		defer wg.Done()
		runtime.SetMutexProfileFraction(1) // 开启对锁调用的跟踪
		runtime.SetBlockProfileRate(1)     // 开启对阻塞操作的跟踪
		err := http.ListenAndServe("0.0.0.0:8080", nil)
		fmt.Println(err)
	}()

	for i := 0; i < 100000; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			data := kelleyRabbimqPool.GetRabbitMqDataFormat("fn_fnout_test", kelleyRabbimqPool.EXCHANGE_TYPE_FANOUT, "", "", fmt.Sprintf("这里是数据%d", num))
			err := instanceRPool.Push(data)
			if err != nil {
				fmt.Println(err)
			}
		}(i)
	}

	wg.Wait()
}
