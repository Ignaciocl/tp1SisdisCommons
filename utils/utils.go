package utils

import (
	"errors"
	"github.com/Ignaciocl/tp1SisdisCommons"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func WaitForSigterm(manager commons.GracielaManager) {
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

func RecoverFromPanic(manager commons.GracielaManager, connection string) {
	if manager == nil {
		manager, _ = commons.CreateGracefulManager(connection)
	}
	if r := recover(); r != nil {
		log.Infof("recovered from panic at %v", r)
		manager.SignalSigterm()
	}
}

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
