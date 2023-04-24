package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

type Queue[S any, R any] interface {
	SendMessage(message S) error
	ReceiveMessage() (R, error) // blocking until message is received
	Close() error
}

type rabbitQueue[S any, R any] struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue *amqp.Queue
}

func (r *rabbitQueue[S, R]) SendMessage(message S) error {
	data, err := json.Marshal(message)
	if err != nil {
		FailOnError(err, "error while marshalling")
		return errors.New(fmt.Sprintf("data object: %v is not marshable", message))
	}
	ctx := context.Background()
	return r.ch.PublishWithContext(ctx,
		"",           // exchange
		r.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		},
	)
}

func (r *rabbitQueue[S, R]) ReceiveMessage() (R, error) {
	var receivable R
	received := make(chan error)
	var receivedMessage []byte

	go func() {
		msgs, err := r.ch.Consume(
			r.queue.Name, // queue
			"",           // consumer
			true,         // auto-ack
			false,        // exclusive
			false,        // no-local
			false,        // no-wait
			nil,          // args
		)
		if err != nil {
			received <- err
			return
		}
		fmt.Println("waiting for messages")
		rm := <-msgs
		fmt.Printf("message received: %v\n", rm)
		receivedMessage = rm.Body
		received <- err
	}()
	fmt.Println("waiting for message 1")
	err := <-received
	fmt.Println("answered")
	if err != nil {
		FailOnError(err, "Failed to receive a message")
		return receivable, err
	}
	if err = json.Unmarshal(receivedMessage, &receivable); err != nil {
		FailOnError(err, fmt.Sprintf("message %s couldn't be parsed to type %v", string(receivedMessage), receivable))
		return receivable, errors.New(fmt.Sprint("couldn't unmarshall message received"))
	}
	fmt.Printf("returning: %v\n", receivable)
	return receivable, nil
}

func (r *rabbitQueue[S, R]) Close() error {
	r.ch.Close()
	return r.conn.Close()
}

func InitializeRabbitQueue[S, R any](queueName string, connection string) (Queue[S, R], error) {
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
	r := rabbitQueue[S, R]{
		conn: conn,
		ch:   ch,
	}
	q, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		r.Close()
		FailOnError(err, "Failed to declare a queue")
		return nil, err
	}
	r.queue = &q
	return &r, nil
}
