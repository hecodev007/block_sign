package rabbitmq

/**
发送数据
消息发送
*/
type RabbitMqData struct {
	ExchangeName string //交换机名称
	ExchangeType string //交换机类型 见RabbitmqPool.go 常量
	QueueName    string //队列名称
	Route        string //路由
	Data         string //发送数据
}

/**
获取发送数据模板
@param exChangeName 交换机名称
@param exChangeType 交换机类型
@param queueName string 队列名称
@param route string 路由
@param data string 发送的数据
*/
func GetRabbitMqDataFormat(exChangeName string, exChangeType string, queueName string, route string, data string) *RabbitMqData {
	return &RabbitMqData{
		ExchangeName: exChangeName,
		ExchangeType: exChangeType,
		QueueName:    queueName,
		Route:        route,
		Data:         data,
	}
}

/**
获取发送数据模板
过期设置(死信队列)
@param exChangeName 交换机名称
@param exChangeType 交换机类型
@param queueName string 队列名称
@param route string 路由
@param data string 发送的数据
*/
func GetRabbitMqDataFormatExpire(exChangeName string, exChangeType string, queueName string, route string, data string) *RabbitMqData {
	return &RabbitMqData{
		ExchangeName: exChangeName,
		ExchangeType: exChangeType,
		QueueName:    queueName,
		Route:        route,
		Data:         data,
	}
}
