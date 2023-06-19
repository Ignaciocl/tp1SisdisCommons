package healthcheck

import (
	"fmt"
	"github.com/Ignaciocl/tp1SisdisCommons/client"
	"github.com/Ignaciocl/tp1SisdisCommons/healthcheck/internal/config"
	log "github.com/sirupsen/logrus"
)

type healthChecker struct {
	config      *config.HealthCheckConfig
	serviceName string
	socket      client.Client
}

// InitHealthChecker inits a health checker for the given serviceName
func InitHealthChecker(serviceName string) (HealthChecker, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	address := fmt.Sprintf("%s:%v", serviceName, cfg.Port)
	socket := client.NewSocket(
		client.NewSocketConfig(
			cfg.Protocol,
			address,
			cfg.PacketLimit,
		),
	)

	return &healthChecker{
		config:      cfg,
		serviceName: serviceName,
		socket:      socket,
	}, nil
}

// Run triggers the health check routine. This routine listen for heartbeats in a given port and replies them with an ACK message
func (hc *healthChecker) Run() error {
	err := hc.socket.StartListener()
	if err != nil {
		return err
	}

	messageHandler, err := hc.socket.AcceptNewConnections()
	if err != nil {
		return err
	}

	for {
		message, err := messageHandler.Listen()
		if err != nil {
			log.Error(getLogMessage(hc.serviceName, "error listening heartbeat. Accepting new connections again...", err))
			messageHandler, err = hc.socket.AcceptNewConnections()

			if err != nil {
				log.Error(getLogMessage(hc.serviceName, "error accepting new connections", err))
				return err
			}

			continue
		}

		if string(message) == hc.config.HealthCheckMessage {
			log.Debug(getLogMessage(hc.serviceName, fmt.Sprintf("got heartbeat message '%s' correctly! Sending ACK...", string(message)), nil))
			healthCheckACKBytes := []byte(hc.config.HealthCheckACK)

			err = messageHandler.Send(healthCheckACKBytes)
			if err != nil {
				log.Error(getLogMessage(hc.serviceName, "error sending ACK message for heartbeat", err))
			}
		}
	}
}

func getLogMessage(service string, message string, err error) string {
	if err != nil {
		return fmt.Sprintf("[service: %s][status: ERROR] %s: %v", service, message, err)
	}
	return fmt.Sprintf("[service: %s][status: OK] %s", service, message)
}
