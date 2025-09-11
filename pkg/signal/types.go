package signal

import (
	"github.com/gorilla/websocket"
	"sync"
)

const (
	peerIDLength           = 16
	simplePeerPingInterval = 1000 * 60 * 2 //60000 //2 * time.Minute
)

type message struct {
	Type           string   `json:"type"`
	YourPeerID     string   `json:"yourPeerId"`
	SenderPeerID   string   `json:"senderPeerId"`
	ReceiverPeerID string   `json:"receiverPeerId"`
	Room           string   `json:"room"`
	OtherPeerIDs   []string `json:"otherPeerIds"`
}

type serverPeer struct {
	id       string
	socket   *websocket.Conn
	rooms    map[string]struct{}
	lastPing int64
}

type signalingServer struct {
	peersMutex  sync.Mutex
	peersByID   map[string]*serverPeer
	peersByRoom map[string]map[string]struct{}
	upgrader    websocket.Upgrader
	serverOpts  ServerOptions
	serverDone  chan struct{}
}

type ServerOptions struct {
	ApiToken string
	Port     string
}

// SimplePeerInitMessage represents the Go equivalent of the TypeScript type
type SimplePeerInitMessage struct {
	Type       string `json:"type"`
	YourPeerID string `json:"yourPeerId"`
}

// SimplePeerJoinMessage represents the Go equivalent of the TypeScript type
type SimplePeerJoinMessage struct {
	Type string `json:"type"`
	Room string `json:"room"`
}

// SimplePeerJoinedMessage represents the Go equivalent of the TypeScript type
type SimplePeerJoinedMessage struct {
	Type         string   `json:"type"`
	OtherPeerIDs []string `json:"otherPeerIds"`
}

// SimplePeerSignalMessage represents the Go equivalent of the TypeScript type
type SimplePeerSignalMessage struct {
	Type           string `json:"type"`
	Room           string `json:"room"`
	SenderPeerID   string `json:"senderPeerId"`
	ReceiverPeerID string `json:"receiverPeerId"`
	Data           string `json:"data"`
}

// SimplePeerPingMessage represents the Go equivalent of the TypeScript type
type SimplePeerPingMessage struct {
	Type string `json:"type"`
}

// PeerMessage represents the Go equivalent of the TypeScript union type
type PeerMessage interface {
	isPeerMessage()
}

// Implementations of the union type
func (SimplePeerInitMessage) isPeerMessage()   {}
func (SimplePeerJoinMessage) isPeerMessage()   {}
func (SimplePeerJoinedMessage) isPeerMessage() {}
func (SimplePeerSignalMessage) isPeerMessage() {}
func (SimplePeerPingMessage) isPeerMessage()   {}
