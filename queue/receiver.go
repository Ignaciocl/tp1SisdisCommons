package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Ignaciocl/tp1SisdisCommons/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

type receiver[Receivable any] struct {
	connectionEnable
	channelConsuming <-chan amqp.Delivery
	topicName        string
	key              string
	closingChannel   <-chan *amqp.Error
}

func (r *receiver[R]) ReceiveMessage() (R, uint64, error) {
	var receivable R
	received := make(chan error)
	var receivedMessage []byte
	var idMessage uint64

	go func() {
		if r.channelConsuming == nil {
			err := r.ch.ExchangeDeclare(r.topicName, "topic", true, false, false, false, nil)
			utils.LogError(err, "couldn't declare exchange")
			err = r.ch.QueueBind(r.queue.Name, r.key, "", false, nil) // ToDo check this
			utils.LogError(err, "could not bind queue")

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
		select {
		case d := <-r.channelConsuming:
			rm := d
			receivedMessage = rm.Body
			idMessage = rm.DeliveryTag
		case _ = <-r.closingChannel:
			received <- nil
		}
	}()
	err := <-received
	if err != nil {
		if !errors.Is(err, amqp.ErrClosed) {
			utils.LogError(err, "Failed to receive a message")
		}
		return receivable, idMessage, err
	}
	if err = json.Unmarshal(receivedMessage, &receivable); err != nil {
		utils.LogError(err, fmt.Sprintf("message %s couldn't be parsed to type %v", string(receivedMessage), receivable))
		return receivable, idMessage, errors.New(fmt.Sprint("couldn't unmarshall message received"))
	}
	return receivable, idMessage, nil
}

func (r *receiver[R]) AckMessage(messageId uint64) error {
	return r.ch.Ack(messageId, false)
}

func (r *receiver[R]) RejectMessage(messageId uint64) error {
	return r.ch.Nack(messageId, false, true)
}

func InitializeReceiver[R any](queueName string, connection string, key string, topicName string, cr ConnectionRetrievable) (Receiver[R], error) {
	c, err := initializeConnectionRabbit(cr, connection)
	if err != nil {
		utils.LogError(err, "could not establish connection")
		return nil, err
	}
	conn, ch := c.conn, c.ch
	var r receiver[R]
	r.conn = conn
	r.ch = ch

	q, err := ch.QueueDeclare(
		fmt.Sprintf("%s.%s", queueName, key), // name
		true,                                 // durable
		false,                                // delete when unused
		false,                                // exclusive
		false,                                // no-wait
		nil,                                  // arguments
	)
	if err != nil {
		r.Close()
		utils.FailOnError(err, "Failed to declare a queue")
		return nil, err
	}
	errChannel := make(chan *amqp.Error, 1)
	ch.NotifyClose(errChannel)
	r.queue = &q
	r.key = key
	r.closingChannel = errChannel
	r.topicName = topicName
	return &r, nil
}
