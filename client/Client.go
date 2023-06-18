package client

type Client interface {
	OpenConnection() error
	Close() error
	StartListener() error
	AcceptNewConnections() (MessageHandler, error)
	Send(dataAsBytes []byte) error
	Listen() ([]byte, error)
}

type MessageHandler interface {
	Listen() ([]byte, error)
	Send(bytes []byte) error
	Close() error
}
