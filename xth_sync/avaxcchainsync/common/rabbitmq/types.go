package rabbitmq

import "github.com/streadway/amqp"

// ENOSERVER is returned when the server.Server passed is incorrect, nil or
// there was no server.Server passed. Check the method Server() under the
// Option type.
const (
	ENOCONN = "no connection available"
	// ENOEXCHANGE is returned when no server.Exchange is properly declared. The
	// default is not enough, it needs a name.
	ENOSERVER   = "no valid server in rabbit"
	ENOEXCHANGE = "no exchange declared"
	ENOQUEUE    = "no queue declared, can't consume without queues"
)

// Exchange holds the definition of an AMQP exchange.
type ExParams struct {
	Name       string
	Kind       string // Possible options are direct, fanout, topic and headers.
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       amqp.Table
}

// Queue hold definition of AMQP queue
type QueueParams struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table

	Binds []BindParams
}

// Binding used to declare binding between AMQP Queue and AMQP Exchange
type BindParams struct {
	Key    string
	NoWait bool
	Args   amqp.Table
}

type ConsumerParams struct {
	// The consumer is identified by a string that is unique and scoped for all
	// consumers on this channel.
	Tag string
	// When autoAck (also known as noAck) is true, the server will acknowledge
	// deliveries to this consumer prior to writing the delivery to the network.  When
	// autoAck is true, the consumer should not call Delivery.Ack
	AutoAck bool // autoAck
	// Check Queue struct documentation
	Exclusive bool // exclusive
	// When noLocal is true, the server will not deliver publishing sent from the same
	// connection to this consumer. (Do not use Publish and Consume from same channel)
	Internal bool // noLocal
	// Check Queue struct documentation
	NoWait bool // noWait
	// Check Exchange comments for Args
	Args amqp.Table // arguments
}

type PublishParams struct {
	// The key that when publishing a message to a exchange/queue will be only delivered to
	// given routing key listeners
	RoutingKey string
	// Publishing tag
	Tag string
	// Queue should be on the server/broker
	Mandatory bool
	// Consumer should be bound to server
	Immediate bool
}
