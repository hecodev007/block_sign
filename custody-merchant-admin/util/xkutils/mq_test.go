package xkutils

import (
	"custody-merchant-admin/middleware/rabbitmq"
	"fmt"

	"testing"
)

func TestSum(t *testing.T) {
	rmq := rabbitmq.NewMQ(rabbitmq.DefaultMQConfig)
	rmq.ConsumeSimple(func(body []byte) {
		fmt.Println(string(body))
	})

	rmq.PublishSimple("22222222")
	rmq.PublishSimple("you know")
	return
}
