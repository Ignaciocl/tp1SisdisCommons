package healthcheck

import (
	"fmt"
	"github.com/Ignaciocl/tp1SisdisCommons/client"
	"github.com/Ignaciocl/tp1SisdisCommons/healthcheck/internal/config"
	log "github.com/sirupsen/logrus"
)

// Run initialize the health check routine. This routine listen for heartbeats in a given port and replies with
// an ACK message
func Run(serviceName string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	address := fmt.Sprintf("%s:%v", serviceName, cfg.Port)
	socket := client.NewSocket(
		client.NewSocketConfig(
			cfg.Protocol,
			address,
			cfg.PacketLimit,
		),
	)

	err = socket.StartListener()
	if err != nil {
		return err
	}

	conn, err := socket.AcceptNewConnections()
	if err != nil {
		return err
	}

	for {
		message, err := conn.Listen()
		if err != nil {
			log.Error(getLogMessage(serviceName, "error listening heartbeat", err))
			conn, _ = socket.AcceptNewConnections() // ToDo: check this
			continue
		}

		if string(message) == cfg.HealthCheckMessage {
			log.Debug(getLogMessage(serviceName, "got heartbeat message correctly! Sending ACK...", nil))
			healthCheckACKBytes := []byte(cfg.HealthCheckACK)
			err := socket.Send(healthCheckACKBytes)
			if err != nil {
				log.Error(getLogMessage(serviceName, "error sending ACK message for heartbeat", err))
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
