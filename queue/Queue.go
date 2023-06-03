package queue

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type ConnectionRetrievable interface {
	GetChannel() *amqp.Channel
	GetQueue() *amqp.Queue
	Close() error
}

type Sender[Sendable any] interface {
	SendMessage(message Sendable) error
	ConnectionRetrievable
}

type Receiver[Receivable any] interface {
	ReceiveMessage() (Receivable, uint64, error) // blocking until message is received
	AckMessage(id uint64) error
	RejectMessage(id uint64) error
	ConnectionRetrievable
}
