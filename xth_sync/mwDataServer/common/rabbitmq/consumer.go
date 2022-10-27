package rabbitmq

import (
	"mwDataServer/common/log"
	"fmt"
	"github.com/streadway/amqp"
	"sync"
)

// A Consumer represents a RabbitMQ consumer of messages against a set of
// queues. It implements logic to have always a connection available where to
// receive messages from.
type Consumer struct {
	server  *Server
	channel *amqp.Channel
	dqueue  amqp.Queue

	exchange  ExParams
	queue     QueueParams
	cnsparams ConsumerParams

	listeners []chan amqp.Delivery
	messages  <-chan amqp.Delivery
	qosCount  int
	m         sync.Mutex
}

// New returns a Consumer or an error if it's not in a working state. A Server
// argument is required, if it's missing it will return a rabbit.ENOCONN and if
// it's an incorrect value it will return a ENOSERVER.
func NewConsumer(r *Server, e ExParams, q QueueParams, cp ConsumerParams) (*Consumer, error) {
	c := &Consumer{
		server:    r,
		exchange:  e,
		queue:     q,
		cnsparams: cp,
		m:         sync.Mutex{},
	}

	if err := c.validate(); err != nil {
		return nil, err
	}

	if err := c.init(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Consumer) Close() error {
	return c.channel.Close()
}

func (c *Consumer) Subscribe() <-chan amqp.Delivery {
	c.m.Lock()
	defer c.m.Unlock()
	ch := make(chan amqp.Delivery, 1)

	c.listeners = append(c.listeners, ch)

	return ch
}

// RoutingKeys returns a slice of all the routing keys being used against the
// queue in this consumer.
func (c *Consumer) RoutingKeys() []string {
	var rk []string
	for _, b := range c.queue.Binds {
		rk = append(rk, b.Key)
	}

	return rk
}

// ConsumeMessage accepts a handler function and only consumes one message
// stream from RabbitMq
func (c *Consumer) GetMsg() (*amqp.Delivery, error) {
	if !c.server.IsOpen() {
		return nil, fmt.Errorf("don't connect rabbitmq")
	}

	message, ok, err := c.channel.Get(c.dqueue.Name, c.cnsparams.AutoAck)
	if err != nil {
		return nil, err
	}

	if !ok {
		log.Info("No message received")
		return nil, fmt.Errorf("no message received")
	}

	// TODO maybe we should return ok too?
	return &message, nil
}

// Consume accepts a handler function for every message streamed from RabbitMq
// will be called within this handler func
func (c *Consumer) Consume(handler func(delivery *amqp.Delivery)) error {
	var err error
	c.messages, err = c.channel.Consume(
		c.dqueue.Name,
		c.cnsparams.Tag,
		c.cnsparams.AutoAck,
		c.cnsparams.Exclusive,
		c.cnsparams.Internal,
		c.cnsparams.NoWait,
		c.cnsparams.Args,
	)

	if err != nil {
		return err
	}

	log.Info("Consume handle: deliveries channel starting")

	closing := c.server.NotifyClose()
	for c.server.Loop() {
		select {
		case d := <-c.messages:
			//for _, ch := range c.listeners {
			//	ch <- d
			//}

			if handler != nil {
				handler(&d)
			}
		case <-closing:
			log.Info("receive close msg")
			c.Close()
			return nil
		}
	}

	log.Info("handle: deliveries channel closed")
	return nil
}

// init returns an error if any of the processes fail. The processes ran here
// are: channel creation, Exchange declaration, Queues declarations and Queue
// bindings declarations.
func (c *Consumer) init() error {
	c.m.Lock()
	defer c.m.Unlock()

	var err error

	if c.channel != nil {
		c.channel.Close()
	}

	c.channel, err = c.server.Channel()
	if err != nil {
		return err
	}

	if c.qosCount > 0 {
		err = c.channel.Qos(c.qosCount, 0, false)
		if err != nil {
			return err
		}
	}

	err = c.channel.ExchangeDeclare(
		c.exchange.Name,
		c.exchange.Kind,
		c.exchange.Durable,
		c.exchange.AutoDelete,
		c.exchange.Internal,
		c.exchange.NoWait,
		c.exchange.Args,
	)
	if err != nil {
		return err
	}

	c.dqueue, err = c.channel.QueueDeclare(
		c.queue.Name,
		c.queue.Durable,
		c.queue.AutoDelete,
		c.queue.Exclusive,
		c.queue.NoWait,
		c.queue.Args,
	)
	if err != nil {
		return err
	}

	for _, b := range c.queue.Binds {
		err := c.channel.QueueBind(
			c.dqueue.Name,
			b.Key,
			c.exchange.Name,
			b.NoWait,
			b.Args,
		)

		if err != nil {
			return err
		}
	}

	//if err := c.consume(); err != nil {
	//	return err
	//}

	return nil
}

// validate returns an error if the server is not declared (ENOSERVER), the
// exchange is not declared (ENOEXCHANGE) or the Queue was not passed
// (ENOQUEUE).
func (c *Consumer) validate() error {
	if c.server == nil {
		return fmt.Errorf(ENOSERVER)
	}

	if c.exchange.Name == "" {
		return fmt.Errorf(ENOEXCHANGE)
	}

	//if !c.qpassed {
	//	return fmt.Errorf(ENOQUEUE)
	//}

	return nil
}
