package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Ignaciocl/tp1SisdisCommons/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	"strconv"
)

type sender[S any] struct {
	connectionEnable
	currentAmount int
	maxAmount     int
	consumerName  string
}

func (s *sender[S]) SendMessage(message S) error {
	data, err := json.Marshal(message)
	if err != nil {
		utils.LogError(err, "error while marshalling")
		return errors.New(fmt.Sprintf("data object: %v is not marshable", message))
	}
	ctx := context.Background()
	key := "#"
	if s.maxAmount > 0 {
		key = strconv.Itoa(s.currentAmount%s.maxAmount + 1)
		s.currentAmount = s.currentAmount + 1%s.maxAmount
	}
	return s.ch.PublishWithContext(ctx,
		s.consumerName, // exchange
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

func InitializeSender[S any](consumerName string, maxAmountReceivers int, cr ConnectionRetrievable, connectionName string) (Sender[S], error) {
	c, err := initializeConnectionRabbit(cr, connectionName)
	if err != nil {
		utils.LogError(err, "could not establish connection")
		return nil, err
	}
	conn, ch := c.conn, c.ch
	q, err := ch.QueueDeclare(
		fmt.Sprintf("sender-%s", consumerName), // name
		true,                                   // durable
		false,                                  // delete when unused
		false,                                  // exclusive
		false,                                  // no-wait
		nil,                                    // arguments
	)
	if err != nil {
		c.Close()
		utils.FailOnError(err, "Failed to declare a queue")
		return nil, err
	}
	s := sender[S]{
		maxAmount:    maxAmountReceivers,
		consumerName: connectionName,
	}
	s.ch, s.queue, s.conn = ch, &q, conn
	return &s, nil
}
