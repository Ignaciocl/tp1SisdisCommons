package utils

import (
	"errors"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
)

func FailOnError(err error, msg string) {
	if err != nil {
		if errors.Is(err, amqp.ErrClosed) {
			log.Debugf("Channel closed when try to throw error: %v", err)
		} else {
			log.Panicf("%s: %s", msg, err)
		}
	}
}

func LogError(err error, msg string) {
	if err != nil {
		log.Errorf("%s: %s", msg, err)
	}
}

func Contains[T comparable](targetString T, sliceOfStrings []T) bool {
	for i := range sliceOfStrings {
		if sliceOfStrings[i] == targetString {
			return true
		}
	}
	return false
}
