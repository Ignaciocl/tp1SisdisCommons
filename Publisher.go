package commons

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Publish(targetPublic string, body []byte, key string) error
	Close()
}

type publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func (p *publisher) Publish(targetPublic string, body []byte, key string, kindOfPublish string) error {
	err := p.ch.ExchangeDeclare(
		targetPublic,  // name
		kindOfPublish, // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	FailOnError(err, "Failed to declare an exchange")
	ctx := context.Background()
	return p.ch.PublishWithContext(ctx,
		targetPublic, // exchange
		key,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		},
	)
}

func (p *publisher) Close() {
	p.Close()
	p.ch.Close()
}

func CreatePublisher(connection string) (Publisher, error) {

	url := fmt.Sprintf("amqp://guest:guest@%s:5672/", connection)
	conn, err := amqp.Dial(url)
	if err != nil {
		FailOnError(err, "Failed to connect to RabbitMQ")
		conn.Close()
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		FailOnError(err, "Failed to connect to RabbitMQ")
		ch.Close()
		conn.Close()
		return nil, err
	}
	return &publisher{ch: ch, conn: conn}, nil
}
