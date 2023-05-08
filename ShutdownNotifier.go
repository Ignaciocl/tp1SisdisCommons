package commons

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

type GracielaManager interface {
	WaitForSigterm()
	SignalSigterm()
	Close()
}

type graceful struct {
	ch         *amqp.Channel
	connection *amqp.Connection
	queue      amqp.Queue
}

func (g graceful) WaitForSigterm() {
	msgs, err := g.ch.Consume(
		g.queue.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		FailOnError(err, "error initializing sigterm waiter")
	}
	for _ = range msgs {
		return
	}
}

func (g graceful) SignalSigterm() {
	ctx := context.Background()
	g.ch.PublishWithContext(ctx,
		"sigterm", // exchange
		"",
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte("let the galaxy burn"),
		},
	)
}

func (g graceful) Close() {
	g.ch.Close()
	g.connection.Close()
}

func CreateGracefulManager(connection string) (GracielaManager, error) {
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
		conn.Close()
		return nil, err
	}
	_ = ch.ExchangeDeclare(
		"sigterm", // name
		"fanout",  // type
		true,      // durable
		false,     // auto-deleted
		false,     // internal
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		FailOnError(err, "Failed to declare an exchange")
		ch.Close()
		conn.Close()
		return nil, err
	}
	q, _ := ch.QueueDeclare("", true, false, false, false, nil)
	ch.QueueBind(q.Name, "", "sigterm", false, nil)
	return graceful{
		ch:         ch,
		connection: conn,
		queue:      q,
	}, nil
}
