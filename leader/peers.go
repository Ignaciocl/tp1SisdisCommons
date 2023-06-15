package leader

import (
	"encoding/json"
	"fmt"
	"github.com/Ignaciocl/tp1SisdisCommons/client"
	log "github.com/sirupsen/logrus"
	"sync"
)

// Peers is an `interface` exposing methods to handle communication with other
// `bully.bully`s.
//
// NOTE: This project offers a default implementation of the `Peers` interface
// that provides basic functions. This will work for the most simple of use
// cases fo exemples, although I strongly recommend you provide your own, safer
// implementation while doing real work.
type Peers interface {
	Add(ID, addr string, c client.Client)
	Delete(ID string)
	Find(ID string) bool
	Write(ID string, msg Message) error
	PeerData() []struct {
		ID   string
		Addr string
	}
	Close()
}

// PeerMap is a `struct` implementing the `Peers` interface and representing
// a container of `bully.Peer`s.
type PeerMap struct {
	mu    *sync.RWMutex
	peers map[string]*Peer
}

// NewPeerMap returns a new `bully.PeerMap`.
func NewPeerMap() *PeerMap {
	return &PeerMap{mu: &sync.RWMutex{}, peers: make(map[string]*Peer)}
}

// Add creates a new `bully.Peer` and adds it to `pm.peers` using `ID` as a key.
func (pm *PeerMap) Add(ID, addr string, c client.Client) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.peers[ID] = NewPeer(ID, addr, c)
}

// Delete erases the `bully.Peer` corresponding to `ID` from `pm.peers` and closes everything associated with it.
func (pm *PeerMap) Delete(ID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.delete(ID)
}

// delete is not thread safe, will remove peer from peerMap
func (pm *PeerMap) delete(ID string) {
	if p, ok := pm.peers[ID]; ok {
		p.sock.Close()
	}

	delete(pm.peers, ID)
}

// Find returns `true` if `pm.peers[ID]` exists, `false` otherwise.
func (pm *PeerMap) Find(ID string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	_, ok := pm.peers[ID]
	return ok
}

// Write writes `msg` of type Message to `pm.peers[ID]`. It returns `nil` or an `error` if
// something occurs.
func (pm *PeerMap) Write(ID string, msg Message) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	d, _ := json.Marshal(msg)

	log.Infof("sending: %s to: %s", string(d), ID)

	if p, ok := pm.peers[ID]; !ok {
		return fmt.Errorf("write: peer %s not found in PeerMap", ID)

	} else if err := p.sock.Send(d); err != nil {
		return fmt.Errorf("send: %v", err)
	}
	return nil
}

// PeerData returns a slice of anonymous structures representing a tupple
// composed of a `Peer.ID` and `Peer.addr`.
func (pm *PeerMap) PeerData() []struct {
	ID   string
	Addr string
} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var IDSlice []struct {
		ID   string
		Addr string
	}
	for _, peer := range pm.peers {
		IDSlice = append(IDSlice, struct {
			ID   string
			Addr string
		}{
			peer.ID,
			peer.addr,
		})
	}
	return IDSlice
}

func (pm *PeerMap) Close() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for _, p := range pm.PeerData() {
		pm.delete(p.ID)
	}
}
