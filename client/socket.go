package client

import (
	"github.com/Ignaciocl/tp1SisdisCommons/utils"
	log "github.com/sirupsen/logrus"
	"net"
)

type SocketConfig struct {
	Protocol      string
	ServerAddress string
	ServerACK     string
	PacketLimit   int
}

type socket struct {
	config     SocketConfig
	connection net.Conn
	listener   net.Listener
}

// NewSocket returns a socket with the corresponding configuration set.
// OBS: none connection opens here.
func NewSocket(socketConfig SocketConfig) Client {
	return &socket{
		config: socketConfig,
	}
}

func (s *socket) OpenConnection() error {
	connection, err := net.Dial(s.config.Protocol, s.config.ServerAddress)
	if err != nil {
		log.Errorf(
			"action: connect | result: fail | error: %v",
			err,
		)
		return err
	}
	s.connection = connection
	return nil
}

func (s *socket) Close() error {
	if s.connection != nil {
		if err := s.connection.Close(); err != nil {
			return err
		}
	}
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *socket) StartListener() error {
	listener, err := net.Listen(s.config.Protocol, s.config.ServerAddress)
	if err != nil {
		log.Errorf(
			"action: get listener | result: fail | error: %v",
			err,
		)
		return err
	}
	s.listener = listener
	return nil
}

func (s *socket) AcceptNewConnections() (net.Conn, error) {
	connection, err := s.listener.Accept()
	if err != nil {
		log.Errorf(
			"action: accept connection| result: fail | error: %v",
			err,
		)
		return nil, err
	}
	return connection, nil
}

func (s *socket) Send(dataAsBytes []byte) error {
	messageLength := len(dataAsBytes)

	shortWriteAvoidance := 0
	amountOfBytesSent := 0

	for amountOfBytesSent < messageLength {
		lowerLimit := amountOfBytesSent - shortWriteAvoidance
		upperLimit := lowerLimit + s.config.PacketLimit

		if upperLimit > messageLength {
			upperLimit = messageLength
		}

		bytesToSend := dataAsBytes[lowerLimit:upperLimit]
		bytesSent, err := s.connection.Write(bytesToSend)
		if err != nil {
			return err
		}
		amountOfBytesSent += bytesSent
		shortWriteAvoidance = len(bytesToSend) - bytesSent
	}

	return nil
}

func (s *socket) Listen(targetEndMessage string, finMessages []string) ([]byte, error) {
	message := make([]byte, 0) // Will contain the message

	for {
		buffer := make([]byte, s.config.PacketLimit)
		bytesRead, err := s.connection.Read(buffer)
		log.Debugf("debug message: %s", string(buffer))
		if err != nil {
			log.Errorf("unexpected error while trying to get message: %s", err.Error())
			return nil, err
		}

		message = append(message, buffer[:bytesRead]...)
		size := len(message)

		if size >= 4 && string(message[size-4:size]) == targetEndMessage {
			log.Debugf("Got message correctly!")
			break
		}

		if size >= 4 && utils.Contains(string(message), finMessages) {
			log.Debugf("Got FIN message correctly!")
			break
		}
	}

	return message, nil
}
