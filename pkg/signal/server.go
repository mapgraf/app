package signal

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func StartSignalingServerSimplePeer(serverOptions ServerOptions, r *mux.Router) *signalingServer { //, r *mux.Router
	log.DefaultLogger.Info("Starting Signaling Server111111111111111111111")
	server := &signalingServer{
		peersByID:   make(map[string]*serverPeer),
		peersByRoom: make(map[string]map[string]struct{}),
		serverOpts:  serverOptions,
		serverDone:  make(chan struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				log.DefaultLogger.Info(fmt.Sprintf("Upgrader check origin: %v", r.Header))
				return true
			}, ReadBufferSize: 100 * 1024,
			WriteBufferSize: 100 * 1024,
		},
	}

	http.HandleFunc("/", server.handleWebSocket)

	go server.startCleanupRoutine()

	go func() {

		addr := fmt.Sprintf(":%s", serverOptions.Port)
		token := serverOptions.ApiToken //fmt.Sprintf(":%s", serverOptions.ApiToken)

		log.DefaultLogger.Info("WebSocket server listening on ws://localhost " + addr)
		log.DefaultLogger.Info("SERVER API_TOKEN: " + token)

		err := http.ListenAndServe(addr, nil)

		if err != nil {
			log.DefaultLogger.Error("ListenAndServe: ", err)
		}
	}()

	return server
}

func (s *signalingServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conHeader := strings.ToLower(r.Header.Get("Connection"))
	upg := strings.ToLower(r.Header.Get("Upgrade"))
	log.DefaultLogger.Info(fmt.Sprintf("con i web %v %v", conHeader, upg))

	if upg  != "websocket" {
        return
    }

    if conHeader != "upgrade" && conHeader != "keep-alive, upgrade" {
        return
    }

    if conHeader == "keep-alive, upgrade" {
        r.Header.Del("Connection")
        r.Header.Set("Connection", "Upgrade")
    }

	// Log all headers
	log.DefaultLogger.Info(fmt.Sprintf("w's type is %T\n", w))

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.DefaultLogger.Error("WebSocket Upgrade Error: " + err.Error())
		return
	}

	peerID := randomCouchString(peerIDLength)
	log.DefaultLogger.Info(fmt.Sprintf("right after upgrade. new randomCouchStr pearid %v", peerID))

	peer := &serverPeer{
		id:       peerID,
		socket:   conn,
		rooms:    make(map[string]struct{}),
		lastPing: time.Now().Unix(),
	}

	s.peersMutex.Lock()
	log.DefaultLogger.Info(fmt.Sprintf("in lock"))

	s.peersByID[peerID] = peer
	log.DefaultLogger.Info(fmt.Sprintf("before unlock"))
	s.peersMutex.Unlock()
	log.DefaultLogger.Info(fmt.Sprintf("after unlock"))

	s.sendMessage(peer.socket, message{Type: "init", YourPeerID: peerID})

	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.DefaultLogger.Error(fmt.Sprintf("Error reading message:", err))
				s.disconnectSocket(peerID, "socket disconnected")
				return
			}

			peer.lastPing = time.Now().Unix()
			var receivedMessage message
			err = json.Unmarshal(msg, &receivedMessage)
			if err != nil {
				log.DefaultLogger.Error("JSON Unmarshal Error: " + err.Error())
				s.disconnectSocket(peerID, "invalid message format")
				return
			}

			//log.DefaultLogger.Info(fmt.Sprintf("receivedMessage: %s ", receivedMessage))

			switch receivedMessage.Type {
			case "join":
				roomId := receivedMessage.Room
				if !validateIdString(roomId) || !validateIdString(peerID) {
					s.disconnectSocket(peerID, "invalid ids")
					return
				}

				if _, exists := peer.rooms[peerID]; exists {
					return
				}
				peer.rooms[roomId] = struct{}{}

				room := getFromMapOrCreate(
					s.peersByRoom,
					roomId,
					func() map[string]struct{} { return make(map[string]struct{}) },
					nil, // ifWasThere function is not provided in this case
				)
				room[peerID] = struct{}{}
				//log.DefaultLogger.Info(fmt.Sprintf("peersByRoom: %v", s.peersByRoom))

				// Tell everyone about the new room state
				var otherPeerIDs []string
				for otherPeerID := range room {
					otherPeerIDs = append(otherPeerIDs, otherPeerID)
				}

				for _, otherPeerID := range otherPeerIDs {
					otherPeer := s.peersByID[otherPeerID]
					//s.disconnectSocket(otherPeerID, "no license")
					s.sendMessage(
						otherPeer.socket,
						message{
							Type:         "joined",
							OtherPeerIDs: otherPeerIDs,
						},
					)
				}
				log.DefaultLogger.Info(fmt.Sprintf("joined %v. otherpeers: %v", roomId, otherPeerIDs))

			case "signal":
				if receivedMessage.SenderPeerID != peerID {
					s.disconnectSocket(peerID, "spoofed sender")
					return
				}
				//	log.DefaultLogger.Info(fmt.Sprintf("signal: sender %v, receiver %v", receivedMessage.SenderPeerID, receivedMessage.ReceiverPeerID))

				receiver := s.peersByID[receivedMessage.ReceiverPeerID]
				if receiver != nil {
					s.nSendMessage(receiver.socket, msg)
				}

			case "ping":
				// Handle ping message (if needed)
				// ...

			default:
				s.disconnectSocket(peerID, "unknown message type "+receivedMessage.Type)
				return
			}

		}
	}()

	// Handle WebSocket close event
	conn.SetCloseHandler(func(code int, text string) error {
		s.disconnectSocket(peerID, "socket disconnected")
		return nil
	})
}

func (s *signalingServer) disconnectSocket(peerID, reason string) {
	s.peersMutex.Lock()
	defer s.peersMutex.Unlock()

	log.DefaultLogger.Info("# disconnect peer %s reason: %s", peerID, reason)
	peer, ok := s.peersByID[peerID]
	if !ok {
		return
	}

	err := peer.socket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.DefaultLogger.Error("Error sending close message:", err)
	}

	for roomID := range peer.rooms {
		room, ok := s.peersByRoom[roomID]
		if ok {
			delete(room, peerID)
			if len(room) == 0 {
				delete(s.peersByRoom, roomID)
			}
		}
	}

	delete(s.peersByID, peerID)
}

func (s *signalingServer) startCleanupRoutine() {
	for {
		select {
		case <-s.serverDone:
			return
		default:
			time.Sleep(5 * time.Second)
			s.cleanupPeers()
		}
	}
}

func (s *signalingServer) cleanupPeers() {
	s.peersMutex.Lock()
	defer s.peersMutex.Unlock()

	minTime := time.Now().Unix() - int64(simplePeerPingInterval) //.Seconds()
	for peerID, peer := range s.peersByID {
		if peer.lastPing < minTime {
			s.disconnectSocket(peerID, "no ping for 2 minutes")
		}
	}
}

func randomCouchString(length int) string {
	const couchNameChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var text strings.Builder

	for i := 0; i < length; i++ {
		text.WriteByte(couchNameChars[rand.Intn(len(couchNameChars))])
	}

	return text.String()
}

func validateIdString(roomID string) bool {
	return len(roomID) > 2 && len(roomID) < 100
}

func promiseWait(duration time.Duration) <-chan time.Time {
	timeout := time.NewTimer(duration)
	defer timeout.Stop()
	return timeout.C
}

func getFromMapOrCreate[K comparable, V any](
	m map[K]V,
	key K,
	creator func() V,
	ifWasThere func(value V),
) V {
	value, exists := m[key]
	if !exists {
		value = creator()
		m[key] = value
	} else if ifWasThere != nil {
		ifWasThere(value)
	}
	return value
}

func (s *signalingServer) sendMessage(conn *websocket.Conn, msg message) {
	s.peersMutex.Lock()
	defer s.peersMutex.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		log.DefaultLogger.Error("Error encoding message:", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.DefaultLogger.Error("Error sending message:", err)
	}
}

func (s *signalingServer) nSendMessage(conn *websocket.Conn, data []byte) {
	s.peersMutex.Lock()
	defer s.peersMutex.Unlock()

	//data, err := json.Marshal(msg)
	//if err != nil {
	//	log.DefaultLogger.Error("Error encoding message:", err)
	//	return
	//}

	err := conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.DefaultLogger.Error("Error sending message:", err)
	}
}
