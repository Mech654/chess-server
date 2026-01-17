package lobby

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Allow all origins for dev
}

type Lobby struct {
	// Clients maps a connection to their auto-generated username
	clients map[*websocket.Conn]string
	mutex   sync.Mutex
}

func New() *Lobby {
	return &Lobby{
		clients: make(map[*websocket.Conn]string),
	}
}

func (l *Lobby) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	// 1. Generate a name and register client
	username := r.URL.Query().Get("username")
	if username == "" {
		username = "Player_" + r.RemoteAddr
	}
	l.mutex.Lock()
	l.clients[conn] = username
	l.mutex.Unlock()

	log.Printf("%s joined the lobby", username)

	// 2. Keep connection alive and listen for disconnects
	defer func() {
		l.mutex.Lock()
		delete(l.clients, conn)
		l.mutex.Unlock()
		conn.Close()
		l.broadcastPlayerList()
	}()

	l.broadcastPlayerList()

	for {
		// Read messages (or just wait for disconnect)
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (l *Lobby) broadcastPlayerList() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 1. Create a slice of current usernames
	usernames := []string{}
	for _, name := range l.clients {
		usernames = append(usernames, name)
	}

	// 2. Turn that slice into JSON
	message, err := json.Marshal(map[string]interface{}{
		"type":    "PLAYER_LIST",
		"players": usernames,
	})
	if err != nil {
		return
	}

	// 3. Send to every connected client
	for conn := range l.clients {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			conn.Close()
			delete(l.clients, conn)
		}
	}
}

// for {
// 	// Read messages
// 	_, message, err := conn.ReadMessage()
// 	if err != nil {
// 		break
// 	}
//
// 	var envelope WSmessage
// 	json.Unmarshal(message, &envelope)
// }
