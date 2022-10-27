package rabbitmq

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

//rabbitMQ结构体
type RbMQ struct {
	// 连接
	conn *amqp.Connection
	// 通道
	channel *amqp.Channel
	// 队列名称
	QueueName string
	// 交换机名称
	Exchange string
	// bind Key 名称
	Key string
	// 连接信息
	Mqurl string
}

func (r *RbMQ) Error() error {
	return nil
}

// NewRabbitMQ 创建结构体实例
func NewRabbitMQ(queueName string, exchange string, key string, mqurl string) *RbMQ {
	return &RbMQ{QueueName: queueName, Exchange: exchange, Key: key, Mqurl: mqurl}
}

// Destory 断开channel 和 connection
func (r *RbMQ) Destory() {
	r.channel.Close()
	r.conn.Close()
}

// failOnErr 错误处理函数
func (r *RbMQ) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s", message, err)
		panic(fmt.Sprintf("%s:%s", message, err))
	}
}

// NewRabbitMQSimple 创建简单模式下RabbitMQ实例
func NewRabbitMQSimple(queueName string, exchangeName string, key string, mqurl string) *RbMQ {
	//创建RabbitMQ实例
	rabbitmq := NewRabbitMQ(queueName, exchangeName, key, mqurl)
	var err error
	//获取connection
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "failed to connect push!")
	//获取channel
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "failed to open a channel")
	return rabbitmq
}

// NewRabbitMQPubSub 订阅模式创建RabbitMQ实例
func NewRabbitMQPubSub(queueName string, exchangeName string, key string, mqurl string) *RbMQ {
	//创建RabbitMQ实例
	rabbitmq := NewRabbitMQ(queueName, exchangeName, key, mqurl)
	var err error
	//获取connection
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "failed to connect push!")
	//获取channel
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "failed to open a channel")
	return rabbitmq
}

// NewRabbitMQRouting 路由模式
//创建RabbitMQ实例
func NewRabbitMQRouting(queueName string, exchangeName string, routingKey string, mqurl string) *RbMQ {
	//创建RabbitMQ实例
	rabbitmq := NewRabbitMQ(queueName, exchangeName, routingKey, mqurl)
	var err error
	//获取connection
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "failed to connect push!")
	//获取channel
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "failed to open a channel")
	return rabbitmq
}

// NewRabbitMQTopic 话题模式
// 创建RabbitMQ实例
func NewRabbitMQTopic(queueName string, exchangeName string, routingKey string, mqurl string) *RbMQ {
	//创建RabbitMQ实例
	rabbitmq := NewRabbitMQ(queueName, exchangeName, routingKey, mqurl)
	var err error
	//获取connection
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "failed to connect push!")
	//获取channel
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "failed to open a channel")
	return rabbitmq
}
