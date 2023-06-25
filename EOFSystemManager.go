package commons

import (
	"context"
	"encoding/json"
	"github.com/Ignaciocl/tp1SisdisCommons/queue"
	"github.com/Ignaciocl/tp1SisdisCommons/rabbitconfigfactory"
	"github.com/Ignaciocl/tp1SisdisCommons/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

const eofRoutingKey = "eof"

type EofData struct {
	EOF            bool   `json:"eof"`
	IdempotencyKey string `json:"idempotencyKey"`
}

type WaitForEof interface {
	AnswerEofOk(key string, actionable Actionable)
	Close()
}

type Actionable interface {
	DoActionIfEOF()
}

type answerEofOk struct {
	nextToNotify    []NextToNotify
	queueInfo       queue.ConnectionRetrievable
	necessaryAmount int
	current         map[string]int
}

func (a *answerEofOk) AnswerEofOk(key string, actionable Actionable) {
	if d, ok := a.current[key]; ok {
		d += 1
		a.current[key] = d
	} else {
		a.current[key] = 1
	}
	d := a.current[key]
	if d >= a.necessaryAmount {
		log.Infof("received %d times the key: %s, map is: %v", d, key, a.current)
		a.current[key] = 0
		if actionable != nil {
			actionable.DoActionIfEOF()
		}
		a.sendEOF(key)
	}
}

func (a *answerEofOk) Close() {
	// No Op
}

// sendEOF sends an EOF with the given idempotency key
func (a *answerEofOk) sendEOF(key string) {
	if a.nextToNotify == nil {
		return
	}
	for _, v := range a.nextToNotify {
		ctx := context.Background()

		body, _ := json.Marshal(EofData{
			EOF:            true,
			IdempotencyKey: key,
		})
		var ch *amqp.Channel
		if v.Connection == nil {
			ch = a.queueInfo.GetChannel()
		} else {
			ch = v.Connection.GetChannel()
		}

		publishingConfig := rabbitconfigfactory.NewPublishingConfig(v.Name, eofRoutingKey)

		ch.PublishWithContext(ctx,
			publishingConfig.Exchange,
			publishingConfig.RoutingKey,
			publishingConfig.Mandatory,
			publishingConfig.Immediate,
			amqp.Publishing{
				ContentType: publishingConfig.ContentType,
				Body:        body,
			},
		)
	}
}

type NextToNotify struct {
	Name       string
	Connection queue.ConnectionRetrievable // Notify with a channel previously created, default is the current
}

func CreateConsumerEOF(nextInLine []NextToNotify, queueType string, queue queue.ConnectionRetrievable, necessaryAmount int) (WaitForEof, error) {
	exchangeConfig := rabbitconfigfactory.NewExchangeDeclarationConfig(queueType, "topic")
	if err := queue.GetChannel().ExchangeDeclare(
		exchangeConfig.Name,
		exchangeConfig.Type,
		exchangeConfig.Durable,
		exchangeConfig.AutoDeleted,
		exchangeConfig.Internal,
		exchangeConfig.NoWait,
		exchangeConfig.Arguments,
	); err != nil {
		return nil, err
	}

	err := queue.GetChannel().QueueBind(
		queue.GetQueue().Name, // queue name
		eofRoutingKey,         // routing key
		queueType,             // exchange
		false,
		nil,
	)
	utils.FailOnError(err, "couldn't bind to target")
	kv := make(map[string]int, 0)
	log.Infof("queue for manager eof %s created", queueType)
	return &answerEofOk{queueInfo: queue, nextToNotify: nextInLine, necessaryAmount: necessaryAmount, current: kv}, nil
}

func WaitForSigterm(manager GracielaManager) {
	oniChan := make(chan os.Signal, 1)
	// catch SIGETRM or SIGINTERRUPT
	signal.Notify(oniChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		if manager != nil {
			manager.WaitForSigterm()
			oniChan <- syscall.SIGTERM
		}
	}()
	<-oniChan
}

func RecoverFromPanic(manager GracielaManager, connection string) {
	if manager == nil {
		manager, _ = CreateGracefulManager(connection)
	}
	if r := recover(); r != nil {
		log.Infof("recovered from panic at %v", r)
		manager.SignalSigterm()
	}
}
