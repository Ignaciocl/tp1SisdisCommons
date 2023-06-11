package client

import "net"

type Client interface {
	OpenConnection() error
	Close() error
	StartListener() error
	AcceptNewConnections() (net.Conn, error)
	Send(dataAsBytes []byte) error
	Listen(targetEndMessage string, finMessages []string) ([]byte, error)
}
