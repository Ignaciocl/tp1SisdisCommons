package client

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"
)

const amountBytesPrefix = 5

type SocketConfig struct {
	Protocol    string
	NodeAddress string
	PacketLimit int
}

func NewSocketConfig(protocol string, nodeAddress string, packetLimit int) SocketConfig {
	return SocketConfig{
		Protocol:    protocol,
		NodeAddress: nodeAddress,
		PacketLimit: packetLimit,
	}
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
	connection, err := net.Dial(s.config.Protocol, s.config.NodeAddress)
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
	listener, err := net.Listen(s.config.Protocol, s.config.NodeAddress)
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

func (s *socket) AcceptNewConnections() (MessageHandler, error) {
	connection, err := s.listener.Accept()
	if err != nil {
		log.Errorf(
			"action: accept connection| result: fail | error: %v",
			err,
		)
		return nil, err
	}

	return &socket{
		connection: connection,
		config:     s.config,
	}, nil
}

func (s *socket) Send(bytes []byte) error {
	size := len(bytes)
	bytesAmount := []byte(fmt.Sprintf("%05d", size))
	bytesToSend := append(bytesAmount, bytes...)
	size = len(bytesToSend)

	for i := 0; i <= len(bytesToSend); i += s.config.PacketLimit {
		var sending []byte
		if size < i+s.config.PacketLimit {
			sending = bytesToSend[i:size]
		} else {
			sending = bytesToSend[i : i+s.config.PacketLimit]
		}
		amountSent, err := s.connection.Write(sending)
		if err != nil {
			log.Printf("weird error happened, stopping but something should be checked: %v", err)
			return err
		}
		if dif := len(sending) - amountSent; dif > 0 { // Avoiding short write
			i -= dif
		}
	}
	return nil
}

func (s *socket) Listen() ([]byte, error) {
	bytesToRead := make([]byte, amountBytesPrefix)
	total := make([]byte, 0)
	if i, err := s.connection.Read(bytesToRead); i < amountBytesPrefix {
		bytesToRead = bytesToRead[0:i]
		if err != nil {
			log.Errorf("error while reading is %v", err)
			return nil, err
		}
		j := i
		remaining := amountBytesPrefix
		for {
			r := remaining - j
			remaining -= j
			innerBytes := make([]byte, r)
			j, err = s.connection.Read(innerBytes)
			if err != nil {
				log.Errorf("error while receiving amount of bytes of message, ending receiver: %v", err)
				return nil, err
			}
			innerBytes = innerBytes[0:j]
			bytesToRead = append(bytesToRead, innerBytes...)
			if j == remaining {
				break
			}
		}
	}
	n, err := strconv.Atoi(string(bytesToRead))
	if err != nil {
		log.Errorf("error is %v while converting %s", err, string(bytesToRead))
	}
	realN := n
	received := make([]byte, n)
	for {
		if i, err := s.connection.Read(received); err != nil {
			log.Errorf("error while receiving message, ending receiver: %v", err)
			return nil, err
		} else {
			total = append(total, received[0:i]...)
			if i < n {
				n = n - i
				received = make([]byte, n)
			} else {
				break
			}
		}
	}
	finalData := total[0:realN]
	return finalData, nil
}
