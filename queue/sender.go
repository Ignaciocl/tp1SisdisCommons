package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Ignaciocl/tp1SisdisCommons/rabbitconfigfactory"
	"github.com/Ignaciocl/tp1SisdisCommons/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	"strconv"
)

const wildcardRoutingKey = "#"

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
	routingKey := wildcardRoutingKey
	if s.maxAmount > 0 {
		routingKey = strconv.Itoa(s.currentAmount%s.maxAmount + 1)
		s.currentAmount = s.currentAmount + 1%s.maxAmount
	}

	publishConfig := rabbitconfigfactory.NewPublishingConfig(s.consumerName, routingKey)
	return s.ch.PublishWithContext(ctx,
		publishConfig.Exchange,
		publishConfig.RoutingKey,
		publishConfig.Mandatory,
		publishConfig.Immediate,
		amqp.Publishing{
			DeliveryMode: amqp.Transient,
			ContentType:  publishConfig.ContentType,
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
	queueDeclarationConfig := rabbitconfigfactory.NewQueueDeclarationConfig(fmt.Sprintf("sender-%s", consumerName))
	q, err := ch.QueueDeclare(
		queueDeclarationConfig.Name,
		queueDeclarationConfig.Durable,
		queueDeclarationConfig.DeleteWhenUnused,
		queueDeclarationConfig.Exclusive,
		queueDeclarationConfig.NoWait,
		queueDeclarationConfig.Arguments,
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
