package rabbitmq

/**
连接负载处理
*/
type RabbitLoadBalance struct {
}

func NewRabbitLoadBalance() *RabbitLoadBalance {
	return &RabbitLoadBalance{}
}

/**
负载均衡
轮循
*/
func (r *RabbitLoadBalance) RoundRobin(cIndex, max int32) int32 {
	if max == 0 {
		return 0
	}
	return (cIndex + 1) % max
}
