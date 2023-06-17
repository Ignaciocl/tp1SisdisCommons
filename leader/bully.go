package leader

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Ignaciocl/tp1SisdisCommons/client"
	log "github.com/sirupsen/logrus"
	"io"
	"sync"
	"time"
)

const (
	protocol    = "tcp"
	packetLimit = 8 * 1024
)

// bully is a `struct` representing a single node used by the `bully Algorithm`.
//
// NOTE: More details about the `bully algorithm` can be found here
// https://en.wikipedia.org/wiki/Bully_algorithm .
type bully struct {
	client       client.Client
	ID           string
	addr         string
	coordinator  string
	peers        Peers // ReImplement this
	mu           *sync.RWMutex
	receiveChan  chan Message
	electionChan chan Message
	leaderChan   chan struct{}
	ack          string
}

// NewBully returns a new `bully` or an `error`.
//
// NOTE: All connections to `Peer`s are established during this function.
//
// NOTE: The `proto` value can be one of this list: `tcp`, `tcp4`, `tcp6`.
func NewBully(ID, addr, proto string, peers map[string]string) (Leader, error) {
	b := &bully{
		ID:           ID,
		addr:         addr,
		coordinator:  "0",
		peers:        NewPeerMap(),
		mu:           &sync.RWMutex{},
		electionChan: make(chan Message, 1),
		receiveChan:  make(chan Message, 10),
		leaderChan:   make(chan struct{}, 100),
	}

	if err := b.Listen(proto, addr); err != nil {
		return nil, fmt.Errorf("NewBully: %v", err)
	}

	b.Connect(peers)
	return b, nil
}

func (b *bully) receive(conn client.Receiver) {
	var msg Message
	var peerId string

	for {
		data, err := conn.Listen()
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Printf("error while receiving info: %v\n", err)
			continue
		}
		if err == nil {
			if err := json.Unmarshal(data, &msg); err != nil {
				fmt.Printf("error while unmarshalling info: %v data is: %s\n", err, string(data))
				continue
			}
			peerId = msg.PeerID
		}
		if (err != nil && errors.Is(err, io.EOF)) || msg.Type == CLOSE {
			_ = conn.Close()
			if peerId == b.Coordinator() {
				b.peers.Delete(msg.PeerID)
				b.Elect()
			}
			break
		} else if msg.Type == OK {
			select {
			case b.electionChan <- msg:
				continue
			case <-time.After(200 * time.Millisecond):
				continue
			}
		} else {
			b.receiveChan <- msg
		}
	}
}

// listen is a helper function that spawns goroutines handling new `Peers`
// connections to `b`'s socket.
//
// NOTE: this function is an infinite loop.
func (b *bully) listen() {
	for {
		conn, err := b.client.AcceptNewConnections()
		if err != nil {
			log.Printf("listen: %v", err)
			continue
		}
		go b.receive(conn)
	}
}

// Listen makes `b` listens on the address `addr` provided using the protocol
// `proto` and returns an `error` if something occurs.
func (b *bully) Listen(proto, addr string) error {
	socketConfig := client.NewSocketConfig(protocol, addr, packetLimit)
	b.client = client.NewSocket(socketConfig)
	if err := b.client.StartListener(); err != nil {
		return fmt.Errorf("error while creating listener: %v", err)
	}
	go b.listen()
	return nil
}

// connect is a helper function that resolves the tcp address `addr` and try
// to establish a tcp connection using the protocol `proto`. The established
// connection is set to `b.peers[ID]` or the function returns an `error`
// if something occurs.
//
// NOTE: In the case `ID` already exists in `b.peers`, the new connection
// replaces the old one.
func (b *bully) connect(addr, ID string) error {
	socketConfig := client.NewSocketConfig(protocol, addr, packetLimit)
	c := client.NewSocket(socketConfig)
	if err := c.OpenConnection(); err != nil {
		_ = c.Close()
		return fmt.Errorf("could not open connection to ID: %s, err: %v\n", ID, err)
	}
	b.peers.Add(ID, addr, c)
	return nil
}

// Connect performs a connection to the remote `Peer`s.
func (b *bully) Connect(peers map[string]string) {
	for ID, addr := range peers {
		if b.ID == ID {
			continue
		}
		if err := b.connect(addr, ID); err != nil {
			log.Printf("Connect: %v", err)
			b.peers.Delete(ID)
		}
	}
}

// Send sends a `bully.Message` of type `what` to `b.peer[to]` at the address
// `addr`. If no connection is reachable at `addr` or if `b.peer[to]` does not
// exist, the function retries five times and returns an `error` if it does not
// succeed.
func (b *bully) Send(to, addr string, what int) error {
	maxRetries := 5

	if !b.peers.Find(to) {
		_ = b.connect(addr, to)
	}

	for attempts := 0; ; attempts++ {
		err := b.peers.Write(to, Message{PeerID: b.ID, Addr: b.addr, Type: what})
		if err == nil {
			break
		}
		if attempts > maxRetries && err != nil {
			return fmt.Errorf("send: %v", err)
		}
		_ = b.connect(addr, to)
		time.Sleep(time.Millisecond)
	}
	return nil
}

// SetCoordinator sets `ID` as the new `b.coordinator` if `ID` is greater than
// `b.coordinator` or equal to `b.ID`.
//
// NOTE: This function is thread-safe.
func (b *bully) SetCoordinator(ID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if ID > b.coordinator || ID == b.ID {
		log.Infof("setting coordintator to id: %s", ID)
		b.coordinator = ID
		b.leaderChan <- struct{}{}
	}
}

// Coordinator returns `b.coordinator`.
//
// NOTE: This function is thread-safe.
func (b *bully) Coordinator() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.coordinator
}

// Elect handles the leader election mechanism of the `bully algorithm`.
func (b *bully) Elect() {
	for _, rBully := range b.peers.PeerData() {
		if rBully.ID > b.ID {
			_ = b.Send(rBully.ID, rBully.Addr, ELECTION)
		}
	}

	for {
		select {
		case msg := <-b.electionChan:
			if msg.PeerID < b.ID {
				continue
			}
			return
		case <-time.After(5 * time.Second):
			b.SetCoordinator(b.ID)
			for _, rBully := range b.peers.PeerData() {
				_ = b.Send(rBully.ID, rBully.Addr, COORDINATOR)
			}
			return
		}
	}
}

// Run is an infinite loop.
func (b *bully) Run() {
	b.Elect()
	for msg := range b.receiveChan {
		log.Infof("message received: on run %v\n", msg)
		if msg.Type == ELECTION && msg.PeerID < b.ID {
			_ = b.Send(msg.PeerID, msg.Addr, OK)
			b.Elect()
		} else if msg.Type == COORDINATOR {
			b.SetCoordinator(msg.PeerID)
		}
	}
}

// WakeMeUpWhenSeptemberEnds is blocking until id is leader.
func (b *bully) WakeMeUpWhenSeptemberEnds() {
	for {
		if b.ID == b.Coordinator() {
			break
		}
		<-b.leaderChan
	}
}

func (b *bully) Close() {
	b.peers.Close()
}
