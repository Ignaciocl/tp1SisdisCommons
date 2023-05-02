package commons

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net"
)

type ValidatorStillUsingData interface {
	IsStillUsingNecessaryDataForFile(file string, city string) bool
}

type WaitForEof interface {
	AnswerEofOk(value ValidatorStillUsingData) // blocking method
	Close()
}

type triggerEOF struct {
	File    string `json:"file"`
	City    string `json:"city"`
	Sigterm bool   `json:"sigterm"`
}

type answerEofOk struct {
	addr            string
	conn            *amqp.Connection
	ch              *amqp.Channel
	msgs            <-chan amqp.Delivery
	finish          chan struct{}
	finishedHearing chan struct{}
}

func (a *answerEofOk) AnswerEofOk(value ValidatorStillUsingData) {
	go func() {
		var trigger triggerEOF
		for d := range a.msgs {
			if err := json.Unmarshal(d.Body, &trigger); err != nil {
				FailOnError(err, "could not understand message")
			}
			if value.IsStillUsingNecessaryDataForFile(trigger.File, trigger.City) {
				a.sendEOFCorrect()
			}
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	data := <-a.finish
	a.finishedHearing <- data
}

func (a *answerEofOk) Close() {
	a.Close()
	a.ch.Close()
}

func (a *answerEofOk) sendEOFCorrect() {
	udpAddr, err := net.ResolveUDPAddr("udp", a.addr)
	if err != nil {
		FailOnError(err, "could not parse address")
	}
	conn, _ := net.DialUDP("udp", nil, udpAddr)
	<-a.finishedHearing
	conn.Write([]byte("ok"))
	conn.Close()
}

func CreateConsumerEOF(connection string, queueType string) (WaitForEof, error) {
	url := fmt.Sprintf("amqp://guest:guest@%s:5672/", connection)
	conn, err := amqp.Dial(url)
	if err != nil {
		FailOnError(err, "Failed to connect to RabbitMQ")
		conn.Close()
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		FailOnError(err, "Failed to connect to RabbitMQ")
		ch.Close()
		conn.Close()
		return nil, err
	}
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	FailOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,    // queue name
		"",        // routing key
		queueType, // exchange
		false,
		nil,
	)
	FailOnError(err, "Failed to bind a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	FailOnError(err, "Failed to register a consumer")
	return &answerEofOk{ch: ch, conn: conn, msgs: msgs, finish: make(chan struct{}, 1), finishedHearing: make(chan struct{}, 1)}, nil
}
