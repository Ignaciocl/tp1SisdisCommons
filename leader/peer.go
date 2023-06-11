package leader

import (
	"github.com/Ignaciocl/tp1SisdisCommons/client"
)

// Peer is a `struct` representing a remote `bully.Bully`.
type Peer struct {
	ID   string
	addr string
	sock client.Client
}

// NewPeer returns a new `*bully.Peer`.
func NewPeer(ID, addr string, c client.Client) *Peer {
	return &Peer{ID: ID, addr: addr, sock: c}
}
