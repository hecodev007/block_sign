package common

import (
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/common/log"
	"github.com/group-coldwallet/common/rabbitmq"
	"time"
)

var TestPublisher *rabbitmq.Publisher
var ProdPublisher *rabbitmq.Publisher

func ConnectMQ(url string, env string) bool {
	limit := 10
	queueName := beego.AppConfig.String("coin")
	exName := "dataservice"
	routingKey := queueName

	rmq, err := rabbitmq.NewServer(url, limit)
	if err != nil {
		log.Errorf("rabbitmq NewServer err : %v ", err)
		return false
	}

	if env == "test" {
		TestPublisher, err = rabbitmq.NewPublisher(rmq,
			rabbitmq.ExParams{
				Name:    exName,
				Kind:    "direct",
				Durable: true},
			rabbitmq.PublishParams{
				RoutingKey: routingKey,
			})
	} else if env == "prod" {
		ProdPublisher, err = rabbitmq.NewPublisher(rmq,
			rabbitmq.ExParams{
				Name:    exName,
				Kind:    "direct",
				Durable: true},
			rabbitmq.PublishParams{
				RoutingKey: routingKey,
			})
	}
	if err != nil {
		log.Errorf("NewPublisher err : %v ", err)
		return false
	}

	{
		consumer, err := rabbitmq.NewConsumer(rmq,
			rabbitmq.ExParams{
				Name:    exName,
				Kind:    "direct",
				Durable: true},
			rabbitmq.QueueParams{
				Name:    queueName,
				Durable: true,
				Binds:   []rabbitmq.BindParams{{Key: routingKey}},
			},
			rabbitmq.ConsumerParams{
				Tag: "data-server"},
		)
		if err != nil {
			log.Fatalf("NewConsumer err : %v ", err)
		}
		time.Sleep(time.Second * 1)
		consumer.Close()
	}
	return true
}

func InitMQ() bool  {
	if beego.AppConfig.DefaultBool("enablemq", false) && beego.AppConfig.DefaultBool("enableprodmq", false) && beego.AppConfig.String("prodmqurl") != "" {
		if !ConnectMQ(beego.AppConfig.String("prodmqurl"), "prod") {
			return false
		}
	}

	if beego.AppConfig.DefaultBool("enablemq", false) && beego.AppConfig.DefaultBool("enabletestmq", false) && beego.AppConfig.String("testmqurl") != "" {
		if !ConnectMQ(beego.AppConfig.String("testmqurl"), "test") {
			return false
		}
	}
	return true
}
