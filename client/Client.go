package client

type Client interface {
	OpenConnection() error
	Close() error
	StartListener() error
	AcceptNewConnections() (Receiver, error)
	Send(dataAsBytes []byte) error
	Listen() ([]byte, error)
}

type Receiver interface {
	Listen() ([]byte, error)
	Close() error
}
