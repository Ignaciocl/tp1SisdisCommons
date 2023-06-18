package rabbitconfigfactory

import "github.com/Ignaciocl/tp1SisdisCommons/rabbitconfigfactory/internal/rabbitconfig"

const connectionString = "amqp://guest:guest@%s:5672/"

// NewQueueDeclarationConfig returns the config to declare a queue with the given name
func NewQueueDeclarationConfig(queueName string) rabbitconfig.QueueDeclarationConfig {
	return rabbitconfig.QueueDeclarationConfig{
		Name:             queueName,
		Durable:          true,
		DeleteWhenUnused: false,
		Exclusive:        false,
		NoWait:           false,
		Arguments:        nil,
	}
}

// NewExchangeDeclarationConfig returns the config to declare an exchange with the given name and type
func NewExchangeDeclarationConfig(exchangeName string, exchangeType string) rabbitconfig.ExchangeDeclarationConfig {
	return rabbitconfig.ExchangeDeclarationConfig{
		Name:        exchangeName,
		Type:        exchangeType,
		Durable:     true,
		AutoDeleted: false,
		Internal:    false,
		NoWait:      false,
		Arguments:   nil,
	}
}

// NewPublishingConfig returns the config needed to publish a message in an exchange
func NewPublishingConfig(exchangeName string, routingKey string) rabbitconfig.PublishingConfig {
	return rabbitconfig.PublishingConfig{
		Exchange:    exchangeName,
		RoutingKey:  routingKey,
		Mandatory:   false,
		Immediate:   false,
		ContentType: "application/json",
	}
}

// NewConsumptionConfig returns the config needed to consume from an exchange/queue
func NewConsumptionConfig(queue string, consumer string) rabbitconfig.ConsumptionConfig {
	return rabbitconfig.ConsumptionConfig{
		QueueName: queue,
		Consumer:  consumer,
		AutoACK:   false,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
		Arguments: nil,
	}
}

// GetConnectionString returns the connection string used to establish a connection with RabbitMQ
func GetConnectionString() string {
	return connectionString
}
