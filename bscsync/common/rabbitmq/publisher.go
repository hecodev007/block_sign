package rabbitmq

import (
	"dataserver/log"
	"fmt"
	"github.com/streadway/amqp"
	"sync"
)

// A Publisher represents a RabbitMQ publisher of messages against an specific
// queue. It implements logic to have always a connection available before
// trying to send a message and reconnect as possible.
type Publisher struct {
	server  *Server
	channel *amqp.Channel

	exchange ExParams
	pubparam PublishParams

	m sync.Mutex
}

// New returns a Publisher or an error if is not in a working state. A Server
// argument is required, if it's missing it will return a server.ENOCONN and if
// it's incorrect it will return a ENOSERVER. If no Exchange is passed it will
// return a ENOEXCHANGE.
func NewPublisher(r *Server, e ExParams, pp PublishParams) (*Publisher, error) {
	p := &Publisher{
		server:   r,
		exchange: e,
		pubparam: pp,
		m:        sync.Mutex{},
	}

	if err := p.validate(); err != nil {
		return nil, err
	}

	if err := p.init(); err != nil {
		return nil, err
	}

	return p, nil
}

// init returns an error if it's not possible to initialize a *amqp.Channel
// from the server or something fails during the Exchange declaration.
func (p *Publisher) init() error {
	p.m.Lock()
	defer p.m.Unlock()

	if p.channel != nil {
		p.channel.Close()
	}

	var err error
	p.channel, err = p.server.Channel()
	if err != nil {
		return err
	}

	// if p.tpl.ContentType == "" {
	// 	p.tpl.ContentType = "text/plain"
	// }

	err = p.channel.ExchangeDeclare(
		p.exchange.Name,
		p.exchange.Kind,
		p.exchange.Durable,
		p.exchange.AutoDelete,
		p.exchange.Internal,
		p.exchange.NoWait,
		p.exchange.Args,
	)
	if err != nil {
		return err
	}

	return nil
}

// validate returns an error if the server is not initialized or the exchange
// lacks at least a name.
func (p *Publisher) validate() error {
	if p.server == nil {
		return fmt.Errorf(ENOSERVER)
	}

	if p.exchange.Name == "" {
		return fmt.Errorf(ENOEXCHANGE)
	}

	return nil
}

// RoutingKey returns the routing key to which is publishing against the
// exchange.
func (p *Publisher) RoutingKey() string {
	return p.pubparam.RoutingKey
}

// Send returns an error if something goes wrong while publishing. If it's
// awaiting for a reconnection then it will block until either a reconnection
// succeeds or a close notification is sent, in which case it will return the
// closing error and it won't try to send the message. On subsequent calls it
// will return a ENOSERVER becase server.Server.Loop() in that case will be
// false.
func (p *Publisher) Send(body []byte) error {
	pub := amqp.Publishing{
		ContentType: "text/plain",
		Body:        body,
	}

	// If this is false it will block until either the reconnection attempt is
	// good or the connection get's closed in which case we will return a
	// ENOSERVER.
	if !p.server.IsOpen() {
		if !p.server.Loop() {
			return fmt.Errorf(ENOSERVER)
		}

		reconn := p.server.NotifyReconnect()
		closer := p.server.NotifyClose()

		select {
		case err := <-closer:
			return err
		case <-reconn:
			if err := p.init(); err != nil {
				return err
			}
		}
	}

	return p.channel.Publish(
		p.exchange.Name,
		p.pubparam.RoutingKey,
		p.pubparam.Mandatory,
		p.pubparam.Immediate,
		pub,
	)
}

func (p *Publisher) SendTx(body []byte) error {
	pub := amqp.Publishing{
		ContentType: "text/plain",
		Body:        body,
	}

	if !p.server.IsOpen() {
		if !p.server.Loop() {
			return fmt.Errorf(ENOSERVER)
		}

		reconn := p.server.NotifyReconnect()
		closer := p.server.NotifyClose()

		select {
		case err := <-closer:
			return err
		case <-reconn:
			if err := p.init(); err != nil {
				return err
			}
		}
	}

	// 开启事务
	if err := p.channel.Tx(); err != nil {
		return err
	}

	if err := p.channel.Publish(p.exchange.Name,
		p.pubparam.RoutingKey,
		p.pubparam.Mandatory,
		p.pubparam.Immediate,
		pub); err != nil {
		if err := p.channel.TxRollback(); err != nil {
			return err
		}
		return err
	}

	if err := p.channel.TxCommit(); err != nil {
		if err := p.channel.TxRollback(); err != nil {
			return err
		}
		return err
	}
	return nil
}

// Close will call the method Close against the *amqp.Channel.
func (p *Publisher) Close() error {
	log.Info("publish close")
	return p.channel.Close()
}
