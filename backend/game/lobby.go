package game

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Mech654/chess-server/backend/auth"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Allow all origins for dev
}

var (
	invite_counter uint64
	match_invites  = make(map[uint64]*MatchInvite)
	invites_mutex  sync.Mutex
)

type Player struct {
	username string
	conn     *websocket.Conn
	send     chan []byte
	handler  MessageHandler
}

type Lobby struct {
	players map[*websocket.Conn]*Player
	mutex   sync.Mutex
}

type MessageHandler interface {
	HandleMessage(p *Player, data []byte)
}

type LobbyHandler struct {
	parentLobby *Lobby
}

// DTOs

type MatchInviteDTO struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type MatchAcceptDTO struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func HelperMarshal(data interface{}) []byte {
	b, err := json.Marshal(data)
	if err != nil {
		log.Printf("CRITICAL: Marshal failed: %v", err)
		return []byte("{}")
	}
	return b
}

func HelperEnvelopeMarshal(msgType string, data interface{}) []byte {
	envelope := map[string]interface{}{
		"type": msgType,
		"data": data,
	}
	return HelperMarshal(envelope)
}

func HelperUnmarshal(data []byte, v any) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		log.Printf("CRITICAL: Unmarshal failed: %v", err)
		return err
	}
	return nil
}

func (mi *MatchInvite) Timer() {
	time.Sleep(30 * time.Second)
	invites_mutex.Lock()
	defer invites_mutex.Unlock()

	delete(match_invites, mi.ID)
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
	username, err := auth.GetUsernameFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	l.mutex.Lock()
	l.players[conn] = &Player{
		username: username,
		conn:     conn,
		send:     make(chan []byte, 256),
		handler:  &LobbyHandler{parentLobby: l},
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
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		l.mutex.Lock()
		p := l.players[conn]
		handler := p.handler
		l.mutex.Unlock()

		if handler != nil {
			handler.HandleMessage(p, message)
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
			// Mailbox is full, kill player. Dont do this irl.
			player.conn.Close()
		}
	}
}

func (lh *LobbyHandler) HandleMessage(p *Player, data []byte) {
	var envelope Envelope
	err := json.Unmarshal(data, &envelope)
	if err != nil {
		log.Println("Error unmarshaling message:", err)
		return
	}

	switch envelope.Type {
	case "MATCH_INVITE":
		err, toPlayer, inviteDTO := lh.NewMatchInvite(p, envelope.Data)
		if err != nil {
			log.Println("Error creating MatchInvite:", err)
			return
		}
		toPlayer.send <- HelperEnvelopeMarshal("MATCH_INVITE", inviteDTO)
	case "MATCH_ACCEPT":
		log.Println("MATCH_ACCEPT received")
		err = lh.NewMatchAccept(p, envelope.Data)
	}
}

func (lh *LobbyHandler) NewMatchInvite(p *Player, data json.RawMessage) (error, *Player, *MatchInviteDTO) {
	newID := atomic.AddUint64(&invite_counter, 1)

	var inviteDTO MatchInviteDTO
	HelperUnmarshal(data, &inviteDTO)
	inviteDTO.From = p.username

	toPlayer := findPlayerByUsername(lh.parentLobby, inviteDTO.To)
	if toPlayer == nil {
		return nil, nil, nil
	}

	invite := MatchInvite{
		ID:          newID,
		from_player: p,
		to_player:   toPlayer,
		created_at:  time.Now(),
	}

	invites_mutex.Lock()
	match_invites[newID] = &invite
	invites_mutex.Unlock()

	go invite.Timer()

	return nil, toPlayer, &inviteDTO
}

func (lh *LobbyHandler) NewMatchAccept(p *Player, data json.RawMessage) error {
	// Check for existing invite
	var acceptDTO MatchAcceptDTO
	HelperUnmarshal(data, &acceptDTO)

	invites_mutex.Lock()
	defer invites_mutex.Unlock()
	var foundInvite *MatchInvite
	for _, invite := range match_invites {
		if invite.from_player.username == acceptDTO.From && invite.to_player.username == acceptDTO.To {
			foundInvite = invite
			break
		}
	}

	if foundInvite == nil {
		log.Println("No matching invite found")
		return nil
	}

	delete(match_invites, foundInvite.ID)

	match := &Match{
		Player1:    foundInvite.from_player,
		Player2:    foundInvite.to_player,
		match_info: &MatchInfo{},
		created_at: time.Now(),
	}
	go match.Start()

	return nil
	// Move

}

func findPlayerByUsername(l *Lobby, username string) *Player {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for _, player := range l.players {
		if player.username == username {
			return player
		}
	}
	return nil
}

type Envelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}
