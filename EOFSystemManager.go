package commons

import (
	"context"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
)

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
	nextToNotify    []string
	queueInfo       EOFSender
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
		a.current[key] = 0
		a.sendEOFCorrect(key)
		if actionable != nil {
			actionable.DoActionIfEOF()
		}
	}
}

func (a *answerEofOk) Close() {

}

func (a *answerEofOk) sendEOFCorrect(key string) {
	if a.nextToNotify == nil {
		return
	}
	for _, v := range a.nextToNotify {
		ctx := context.Background()
		body, _ := json.Marshal(EofData{
			EOF:            true,
			IdempotencyKey: key,
		})
		a.queueInfo.GetChannel().PublishWithContext(ctx,
			v, // exchange
			"",
			false, // mandatory
			false, // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        body,
			},
		)
	}
}

type EOFSender interface {
	GetChannel() *amqp.Channel
	GetQueue() *amqp.Queue
}

func CreateConsumerEOF(nextInLine []string, queueType string, queue EOFSender, necessaryAmount int) (WaitForEof, error) {
	if err := queue.GetChannel().ExchangeDeclare(
		queueType, // name
		"fanout",  // type
		true,      // durable
		false,     // auto-deleted
		false,     // internal
		false,     // no-wait
		nil,       // arguments
	); err != nil {
		return nil, err
	}

	err := queue.GetChannel().QueueBind(
		queue.GetQueue().Name, // queue name
		"",                    // routing key
		queueType,             // exchange
		false,
		nil,
	)
	FailOnError(err, "couldn't bind to target")
	kv := make(map[string]int, 0)
	return &answerEofOk{queueInfo: queue, nextToNotify: nextInLine, necessaryAmount: necessaryAmount, current: kv}, nil
}
