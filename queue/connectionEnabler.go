package queue

import (
	"fmt"
	"github.com/Ignaciocl/tp1SisdisCommons/rabbitconfigfactory"
	"github.com/Ignaciocl/tp1SisdisCommons/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

type connectionEnable struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue *amqp.Queue
}

func (c connectionEnable) GetChannel() *amqp.Channel {
	return c.ch
}

func (c connectionEnable) GetQueue() *amqp.Queue {
	return c.queue
}

func (c connectionEnable) Close() error {
	if c.conn != nil {
		if err := c.ch.Close(); err != nil {
			return err
		}
		return c.conn.Close()
	}
	return nil
}

func initializeConnectionRabbit(retrievable ConnectionRetrievable, connection string) (connectionEnable, error) {
	var c connectionEnable
	if retrievable != nil {
		c.ch = retrievable.GetChannel()
		c.conn = nil
	} else {
		url := fmt.Sprintf(rabbitconfigfactory.GetConnectionString(), connection)
		conn, err := amqp.Dial(url)
		if err != nil {
			utils.LogError(err, "Failed to connect to RabbitMQ")
			return c, err
		}

		ch, err := conn.Channel()
		if err != nil {
			utils.LogError(err, "Failed to connect to RabbitMQ")
			_ = conn.Close()
			return c, err
		}

		c.ch = ch
		c.conn = conn
	}
	return c, nil
}

func InitializeConnectionRabbit(retrievable ConnectionRetrievable, connection string) (ConnectionRetrievable, error) {
	return initializeConnectionRabbit(retrievable, connection) // mask value behind interface for external purposes
}
