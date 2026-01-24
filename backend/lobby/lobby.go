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

type Player struct {
	username string
	conn     *websocket.Conn
	send     chan []byte
}

type Lobby struct {
	players map[*websocket.Conn]*Player
	mutex   sync.Mutex
}

func (c *Player) writePump() {
	defer c.conn.Close()
	for {
		message, ok := <-c.send
		if !ok {
			return
		}
		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}
}

func New() *Lobby {
	return &Lobby{
		players: make(map[*websocket.Conn]*Player),
	}
}

func (l *Lobby) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		username = "Player_" + r.RemoteAddr
	}
	l.mutex.Lock()
	l.players[conn] = &Player{
		username: username,
		conn:     conn,
		send:     make(chan []byte, 256),
	}
	l.mutex.Unlock()

	go l.players[conn].writePump()

	log.Printf("%s joined the lobby", username)

	defer func() {
		l.mutex.Lock()
		player, exists := l.players[conn]
		if exists {
			close(player.send)
			delete(l.players, conn)
		}
		l.mutex.Unlock()
		conn.Close()
		l.broadcastPlayerList()
	}()

	l.broadcastPlayerList()

	for {
		// Keep the connection alive
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (l *Lobby) broadcastPlayerList() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	usernames := []string{}
	for _, player := range l.players {
		usernames = append(usernames, player.username)
	}

	message, err := json.Marshal(map[string]interface{}{
		"type":    "PLAYER_LIST",
		"players": usernames,
	})
	if err != nil {
		return
	}

	for _, player := range l.players {
		select {
		case player.send <- message:
			// Message queued successfully
		default:
			// Mailbox is full
			close(player.send)
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
