package queue

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Ignaciocl/tp1SisdisCommons/rabbitconfigfactory"
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
			exchangeDeclarationConfig := rabbitconfigfactory.NewExchangeDeclarationConfig(r.topicName, "topic")
			err := r.ch.ExchangeDeclare(
				exchangeDeclarationConfig.Name,
				exchangeDeclarationConfig.Type,
				exchangeDeclarationConfig.Durable,
				exchangeDeclarationConfig.AutoDeleted,
				exchangeDeclarationConfig.Internal,
				exchangeDeclarationConfig.NoWait,
				exchangeDeclarationConfig.Arguments,
			)
			if err != nil {
				utils.LogError(err, "couldn't declare exchange")
				received <- err
				return
			}

			err = r.ch.QueueBind(r.queue.Name, r.key, exchangeDeclarationConfig.Name, false, nil) // ToDo check this
			if err != nil {
				utils.LogError(err, "could not bind queue")
				received <- err
				return
			}

			consumptionConfig := rabbitconfigfactory.NewConsumptionConfig(r.queue.Name, "")
			msgs, err := r.ch.Consume(
				consumptionConfig.QueueName,
				consumptionConfig.Consumer,
				consumptionConfig.AutoACK,
				consumptionConfig.Exclusive,
				consumptionConfig.NoLocal,
				consumptionConfig.NoWait,
				consumptionConfig.Arguments,
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

	queueDeclarationConfig := rabbitconfigfactory.NewQueueDeclarationConfig(fmt.Sprintf("%s.%s", queueName, key))
	q, err := ch.QueueDeclare(
		queueDeclarationConfig.Name,
		queueDeclarationConfig.Durable,
		queueDeclarationConfig.DeleteWhenUnused,
		queueDeclarationConfig.Exclusive,
		queueDeclarationConfig.NoWait,
		queueDeclarationConfig.Arguments,
	)
	if err != nil {
		_ = r.Close()
		utils.FailOnError(err, "Failed to declare a queue")
		return nil, err
	}
	errChannel := make(chan *amqp.Error, 1)
	ch.NotifyClose(errChannel)
	r.queue = &q
	r.key = key
	r.closingChannel = errChannel
	r.topicName = topicName
	if topicName == "" {
		r.topicName = queueName
	}
	return &r, nil
}
