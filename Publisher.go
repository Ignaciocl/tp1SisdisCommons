package commons

import (
	"context"
	"fmt"
	"github.com/Ignaciocl/tp1SisdisCommons/queue"
	"github.com/Ignaciocl/tp1SisdisCommons/rabbitconfigfactory"
	"github.com/Ignaciocl/tp1SisdisCommons/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Publish(targetPublic string, body []byte, key string, kop string) error
	Close()
}

type publisher struct {
	data queue.ConnectionRetrievable
}

func (p *publisher) Publish(targetPublic string, body []byte, routingKey string, kindOfPublish string) error {
	exchangeDeclarationConfig := rabbitconfigfactory.NewExchangeDeclarationConfig(targetPublic, kindOfPublish)
	err := p.data.GetChannel().ExchangeDeclare(
		exchangeDeclarationConfig.Name,
		exchangeDeclarationConfig.Type,
		exchangeDeclarationConfig.Durable,
		exchangeDeclarationConfig.AutoDeleted,
		exchangeDeclarationConfig.Internal,
		exchangeDeclarationConfig.NoWait,
		exchangeDeclarationConfig.Arguments,
	)
	if err != nil {
		utils.FailOnError(err, "Failed to declare an exchange")
		return err
	}

	ctx := context.Background()
	publishConfig := rabbitconfigfactory.NewPublishingConfig(targetPublic, routingKey)
	err = p.data.GetChannel().PublishWithContext(ctx,
		publishConfig.Exchange,
		publishConfig.RoutingKey,
		publishConfig.Mandatory,
		publishConfig.Immediate,
		amqp.Publishing{
			ContentType: publishConfig.ContentType,
			Body:        body,
		},
	)

	if err != nil {
		utils.FailOnError(err, fmt.Sprintf("Failed publishing message in exchange '%s' with routing key '%s'", publishConfig.Exchange, routingKey))
		return err
	}
	return nil
}

func (p *publisher) Close() {
	p.data.Close()
}

func CreatePublisher(connection string, sender queue.ConnectionRetrievable) (Publisher, error) {
	c, err := queue.InitializeConnectionRabbit(sender, connection)
	return &publisher{c}, err
}
