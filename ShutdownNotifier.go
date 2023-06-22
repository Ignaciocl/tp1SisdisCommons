package commons

import (
	"context"
	"fmt"
	"github.com/Ignaciocl/tp1SisdisCommons/rabbitconfigfactory"
	"github.com/Ignaciocl/tp1SisdisCommons/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	sigtermStr = "sigterm"
	fanoutStr  = "fanout"
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
	consumptionConfig := rabbitconfigfactory.NewConsumptionConfig(g.queue.Name, "")
	msgs, err := g.ch.Consume(
		consumptionConfig.QueueName,
		consumptionConfig.Consumer,
		true,
		consumptionConfig.Exclusive,
		consumptionConfig.NoLocal,
		consumptionConfig.NoWait,
		consumptionConfig.Arguments,
	)
	if err != nil {
		utils.FailOnError(err, "error initializing sigterm waiter")
	}

	for _ = range msgs {
		return
	}
}

func (g graceful) SignalSigterm() {
	if g.ch.IsClosed() || true { // We won't do anything here for now, lets discuss the functionality of this method
		return
	}
	ctx := context.Background()
	publishConfig := rabbitconfigfactory.NewPublishingConfig(sigtermStr, "")
	g.ch.PublishWithContext(ctx,
		publishConfig.Exchange,
		publishConfig.RoutingKey,
		publishConfig.Mandatory,
		publishConfig.Immediate,
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
	url := fmt.Sprintf(rabbitconfigfactory.GetConnectionString(), connection)
	conn, err := amqp.Dial(url)
	if err != nil {
		utils.LogError(err, "Failed to connect to RabbitMQ")
		conn.Close()
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		utils.LogError(err, "Failed to connect to RabbitMQ")
		conn.Close()
		return nil, err
	}

	exchangeDeclarationConfig := rabbitconfigfactory.NewExchangeDeclarationConfig(sigtermStr, fanoutStr)
	err = ch.ExchangeDeclare(
		exchangeDeclarationConfig.Name,
		exchangeDeclarationConfig.Type,
		exchangeDeclarationConfig.Durable,
		exchangeDeclarationConfig.AutoDeleted,
		exchangeDeclarationConfig.Internal,
		exchangeDeclarationConfig.NoWait,
		exchangeDeclarationConfig.Arguments,
	)

	if err != nil {
		utils.LogError(err, "Failed to declare 'sigterm' exchange")
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	queueDeclarationConfig := rabbitconfigfactory.NewQueueDeclarationConfig("")
	anonymousQueue, err := ch.QueueDeclare(
		queueDeclarationConfig.Name,
		queueDeclarationConfig.Durable,
		queueDeclarationConfig.DeleteWhenUnused,
		queueDeclarationConfig.Exclusive,
		queueDeclarationConfig.NoWait,
		queueDeclarationConfig.Arguments,
	)

	if err != nil {
		utils.LogError(err, "Failed to declare anonymous queue")
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	err = ch.QueueBind(anonymousQueue.Name, "", sigtermStr, false, nil)
	if err != nil {
		utils.LogError(err, "Failed to declare an exchange")
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	return graceful{
		ch:         ch,
		connection: conn,
		queue:      anonymousQueue,
	}, nil
}
