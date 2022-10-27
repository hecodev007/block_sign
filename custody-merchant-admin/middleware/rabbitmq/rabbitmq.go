package rabbitmq

import (
	. "custody-merchant-admin/config"
	"fmt"
	"github.com/labstack/echo/v4"
)

type mqconf struct {
	Mtype        string
	Mqueue       string
	Key          string
	ExchangeName string
}

const (
	WORK = "WORK"

	SIMPLE = "SIMPLE"

	PUBLISH = "PUBLISH"

	ROUTING = "ROUTING"

	TOPIC = "TOPIC"
)

var (
	//Murl = "amqps://hoocustody-mq:CsMa8GkLn7hjHPJ4@b-59fecdb5-5f87-4fd4-b099-121f293555ea.mq.ap-northeast-1.amazonaws.com:5671"
	Murl = ""
)

func InitConf() {
	Murl = fmt.Sprintf("%s://%s:%s@%s/", Conf.RabbitMQ.Prefix, Conf.RabbitMQ.MQUser, Conf.RabbitMQ.MQPassword, Conf.RabbitMQ.MQUrl)
}

var (
	// DefaultMQConfig is the default Secure middleware config.
	DefaultMQConfig = mqconf{
		Mtype:  WORK,
		Mqueue: "default_simple",
	}
)

func NewMQQueue(m mqconf) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			InitConf()
			switch m.Mtype {
			case WORK:
				NewRabbitMQSimple(m.Mqueue, "", "", Murl).PublishSimple(c.QueryString())
				break
			case SIMPLE:
				NewRabbitMQSimple(m.Mqueue, "", "", Murl).PublishSimple(c.QueryString())
				break
			case PUBLISH:
				NewRabbitMQPubSub(m.Mqueue, m.ExchangeName, m.Key, Murl).PublishPub(c.QueryString())
				break
			case ROUTING:
				NewRabbitMQRouting(m.Mqueue, m.ExchangeName, m.Key, Murl).PublishRouting(c.QueryString())
				break
			case TOPIC:
				NewRabbitMQTopic(m.Mqueue, m.ExchangeName, m.Key, Murl).PublishTopic(c.QueryString())
				break
			default:
				return nil
			}
			return next(c)
		}
	}
}

func NewMQ(m mqconf) *RbMQ {
	InitConf()
	switch m.Mtype {
	case WORK:
		return NewRabbitMQSimple(m.Mqueue, "", "", Murl)
	case SIMPLE:
		return NewRabbitMQSimple(m.Mqueue, "", "", Murl)
	case PUBLISH:
		return NewRabbitMQPubSub(m.Mqueue, m.ExchangeName, m.Key, Murl)
	case ROUTING:
		return NewRabbitMQRouting(m.Mqueue, m.ExchangeName, m.Key, Murl)
	case TOPIC:
		return NewRabbitMQTopic(m.Mqueue, m.ExchangeName, m.Key, Murl)
	default:
		return nil
	}
}
