package commons

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

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
