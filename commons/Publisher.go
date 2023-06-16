package commons

import (
	"context"
	"github.com/Ignaciocl/tp1SisdisCommons/queue"
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

func (p *publisher) Publish(targetPublic string, body []byte, key string, kindOfPublish string) error {
	err := p.data.GetChannel().ExchangeDeclare(
		targetPublic,  // name
		kindOfPublish, // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	utils.FailOnError(err, "Failed to declare an exchange")
	ctx := context.Background()
	return p.data.GetChannel().PublishWithContext(ctx,
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
	p.data.Close()
}

func CreatePublisher(connection string, sender queue.ConnectionRetrievable) (Publisher, error) {
	c, err := queue.InitializeConnectionRabbit(sender, connection)
	return &publisher{c}, err
}
