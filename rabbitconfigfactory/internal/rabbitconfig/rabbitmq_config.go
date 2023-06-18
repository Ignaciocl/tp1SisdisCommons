package rabbitconfig

import "github.com/rabbitmq/amqp091-go"

// QueueDeclarationConfig contains the parameters to declare a RabbitMQ queue
type QueueDeclarationConfig struct {
	Name             string
	Durable          bool
	DeleteWhenUnused bool
	Exclusive        bool
	NoWait           bool
	Arguments        amqp091.Table
}

// ExchangeDeclarationConfig contains the parameters to declare a RabbitMQ exchange
type ExchangeDeclarationConfig struct {
	Name        string
	Type        string
	Durable     bool
	AutoDeleted bool
	Internal    bool
	NoWait      bool
	Arguments   amqp091.Table
}

// PublishingConfig config use it for publishing messages in a RabbitMQ queue
type PublishingConfig struct {
	Exchange    string
	RoutingKey  string
	Mandatory   bool
	Immediate   bool
	ContentType string
}

// ConsumptionConfig config use it for consuming a RabbitMQ queue
type ConsumptionConfig struct {
	QueueName string
	Consumer  string
	AutoACK   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Arguments amqp091.Table
}
