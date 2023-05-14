package commons

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"strconv"
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
	IsEmpty() bool
	GetChannel() *amqp.Channel
	GetQueue() *amqp.Queue
}

type rabbitQueue[S any, R any] struct {
	conn             *amqp.Connection
	ch               *amqp.Channel
	queue            *amqp.Queue
	channelConsuming <-chan amqp.Delivery
	consumerName     string
	maxAmount        int
	currentAmount    int
	key              string
}

func (r *rabbitQueue[S, R]) GetChannel() *amqp.Channel {
	return r.ch
}

func (r *rabbitQueue[S, R]) GetQueue() *amqp.Queue {
	return r.queue
}
func (r *rabbitQueue[S, R]) SendMessage(message S) error {
	data, err := json.Marshal(message)
	if err != nil {
		FailOnError(err, "error while marshalling")
		return errors.New(fmt.Sprintf("data object: %v is not marshable", message))
	}
	ctx := context.Background()
	key := ""
	if r.maxAmount > 0 {
		key = strconv.Itoa(r.currentAmount%r.maxAmount + 1)
		r.currentAmount = r.currentAmount + 1%r.maxAmount
	}
	return r.ch.PublishWithContext(ctx,
		r.consumerName, // exchange
		key,            // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Transient,
			ContentType:  "text/plain",
			Body:         data,
		},
	)
}

func (r *rabbitQueue[S, R]) ReceiveMessage() (R, error) {
	var receivable R
	received := make(chan error)
	var receivedMessage []byte

	go func() {
		if r.channelConsuming == nil {
			err := r.ch.ExchangeDeclare(r.consumerName, "topic", true, false, false, false, nil)
			FailOnError(err, "couldn't declare exchange")
			err = r.ch.QueueBind(r.queue.Name, r.key, r.consumerName, false, nil)
			FailOnError(err, "coult not bind queue")

			msgs, err := r.ch.Consume(
				r.queue.Name, // queue
				"",           // consumer
				false,        // auto-ack
				false,        // exclusive
				false,        // no-local
				false,        // no-wait
				nil,          // args
			)
			if err != nil {
				received <- err
				return
			}
			r.channelConsuming = msgs
		}
		rm := <-r.channelConsuming
		receivedMessage = rm.Body
		received <- r.ch.Ack(rm.DeliveryTag, false)
	}()
	err := <-received
	if err != nil {
		FailOnError(err, "Failed to receive a message")
		return receivable, err
	}
	if err = json.Unmarshal(receivedMessage, &receivable); err != nil {
		FailOnError(err, fmt.Sprintf("message %s couldn't be parsed to type %v", string(receivedMessage), receivable))
		return receivable, errors.New(fmt.Sprint("couldn't unmarshall message received"))
	}
	return receivable, nil
}

func (r *rabbitQueue[S, R]) Close() error {
	r.ch.Close()
	return r.conn.Close()
}

func (r *rabbitQueue[S, R]) IsEmpty() bool {
	_, ok, _ := r.ch.Get(r.queue.Name, false)
	return len(r.channelConsuming) == 0 && !ok
}

func InitializeRabbitQueue[S, R any](queueName string, connection string, key string, amountToPublish int) (Queue[S, R], error) {
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
		conn:          conn,
		ch:            ch,
		consumerName:  queueName,
		currentAmount: 0,
		maxAmount:     amountToPublish,
	}

	q, err := ch.QueueDeclare(
		"",    // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		r.Close()
		FailOnError(err, "Failed to declare a queue")
		return nil, err
	}
	r.queue = &q
	r.key = key
	return &r, nil
}
