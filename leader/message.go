package leader

// Message Types.
const (
	ELECTION = iota
	OK
	COORDINATOR
	CLOSE
)

// Message is a `struct` used for communication between `bully.bully`s.
type Message struct {
	PeerID string
	Addr   string
	Type   int
}
