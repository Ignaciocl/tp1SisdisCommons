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
	AnswerEofOk(key string) // blocking method
	Close()
}

type answerEofOk struct {
	nextToNotify    []string
	queueInfo       Queue[EOFSender, EOFSender]
	necessaryAmount int
	current         map[string]int
}

func (a *answerEofOk) AnswerEofOk(key string) {
	if d, ok := a.current[key]; ok {
		d += 1
		if d >= a.necessaryAmount {
			d = 0
			a.sendEOFCorrect(key)
		}
		a.current[key] = d
	} else {
		a.current[key] = 1
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
		a.queueInfo.getChannel().PublishWithContext(ctx,
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
}

func CreateConsumerEOF(nextInLine []string, queueType string, queue Queue[EOFSender, EOFSender], necessaryAmount int) (WaitForEof, error) {
	if err := queue.getChannel().ExchangeDeclare(
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

	err := queue.getChannel().QueueBind(
		queue.getQueue().Name, // queue name
		"",                    // routing key
		queueType,             // exchange
		false,
		nil,
	)
	FailOnError(err, "couldn't bind to target")
	kv := make(map[string]int, 0)
	return &answerEofOk{queueInfo: queue, nextToNotify: nextInLine, necessaryAmount: necessaryAmount, current: kv}, nil
}
