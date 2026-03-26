package game

import (
	"encoding/json"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Mech654/chess-server/backend/chess"
	"github.com/Mech654/chess-server/backend/ws"
)

var (
	inviteCounter uint64
	matchCounter  uint64
	matchInvites  = make(map[uint64]*MatchInvite)
	invitesMutex  sync.Mutex
)

type Lobby struct {
	clients       map[*ws.Client]struct{}
	mutex         sync.Mutex
	matchHandlers *sync.Map
}

type LobbyHandler struct {
	lobby *Lobby
}

type MatchInviteDTO struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type MatchAcceptDTO struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func (mi *MatchInvite) timer() {
	time.Sleep(30 * time.Second)
	invitesMutex.Lock()
	defer invitesMutex.Unlock()

	delete(matchInvites, mi.ID)
}

func NewLobby(matchHandler *sync.Map) *Lobby {
	return &Lobby{
		clients:       make(map[*ws.Client]struct{}),
		matchHandlers: matchHandler,
	}
}

func NewLobbyHandler(lobby *Lobby) *LobbyHandler {
	return &LobbyHandler{lobby: lobby}
}

func (lh *LobbyHandler) HandleConnect(client *ws.Client) {
	lh.lobby.mutex.Lock()
	lh.lobby.clients[client] = struct{}{}
	lh.lobby.mutex.Unlock()

	log.Printf("%s joined the lobby", client.Username)
	lh.lobby.broadcastPlayerList()
}

func (lh *LobbyHandler) HandleDisconnect(client *ws.Client) {
	lh.lobby.mutex.Lock()
	delete(lh.lobby.clients, client)
	lh.lobby.mutex.Unlock()

	lh.lobby.broadcastPlayerList()
}

func (l *Lobby) broadcastPlayerList() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	usernames := []string{}
	for client := range l.clients {
		usernames = append(usernames, client.Username)
	}

	message := ws.EnvelopeMarshal("PLAYER_LIST", map[string]any{"players": usernames})

	for client := range l.clients {
		client.Send(message)
	}
}

func (lh *LobbyHandler) HandleMessage(client *ws.Client, data []byte) {
	var envelope ws.Envelope
	err := json.Unmarshal(data, &envelope)
	if err != nil {
		log.Println("Error unmarshalling message:", err)
		return
	}

	switch envelope.Type {
	case "MATCH_INVITE":
		toClient, inviteDTO, err := lh.NewMatchInvite(client, envelope.Data)
		if err != nil {
			log.Println("Error creating MatchInvite:", err)
			return
		}
		toClient.Send(ws.EnvelopeMarshal("MATCH_INVITE", inviteDTO))
	case "MATCH_ACCEPT":
		log.Println("MATCH_ACCEPT received")
		err = lh.NewMatchAccept(client, envelope.Data)
	}
}

func (lh *LobbyHandler) NewMatchInvite(client *ws.Client, data json.RawMessage) (*ws.Client, *MatchInviteDTO, error) {
	newID := atomic.AddUint64(&inviteCounter, 1)

	var inviteDTO MatchInviteDTO
	if err := ws.Unmarshal(data, &inviteDTO); err != nil {
		log.Println("Invalid format:", err)
	}
	inviteDTO.From = client.Username

	toClient := findClientByUsername(lh.lobby, inviteDTO.To)
	if toClient == nil {
		return nil, nil, nil
	}

	invite := MatchInvite{
		ID:         newID,
		FromClient: client,
		ToClient:   toClient,
		CreatedAt:  time.Now(),
	}

	invitesMutex.Lock()
	matchInvites[newID] = &invite
	invitesMutex.Unlock()

	go invite.timer()

	return toClient, &inviteDTO, nil
}

func (lh *LobbyHandler) NewMatchAccept(client *ws.Client, data json.RawMessage) error {
	var acceptDTO MatchAcceptDTO
	if err := ws.Unmarshal(data, &acceptDTO); err != nil {
		return err
	}

	invitesMutex.Lock()
	defer invitesMutex.Unlock()
	var foundInvite *MatchInvite
	for _, invite := range matchInvites {
		if invite.FromClient.Username == acceptDTO.From && invite.ToClient.Username == client.Username {
			foundInvite = invite
			break
		}
	}

	if foundInvite == nil {
		log.Println("No matching invite found")
		return nil
	}

	delete(matchInvites, foundInvite.ID)

	match := &Match{
		Player1:    foundInvite.FromClient,
		Player2:    foundInvite.ToClient,
		MatchState: &MatchState{GameState: chess.NewGameState()},
		CreatedAt:  time.Now(),
	}

	matchHandler := &MatchHandler{match: match, id: atomic.AddUint64(&matchCounter, 1)}
	lh.lobby.matchHandlers.Store(matchHandler.id, matchHandler)

	foundInvite.FromClient.SetHandler(matchHandler)
	foundInvite.ToClient.SetHandler(matchHandler)

	go match.Start(matchHandler.id)

	return nil
}

func findClientByUsername(l *Lobby, username string) *ws.Client {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for client := range l.clients {
		if client.Username == username {
			return client
		}
	}
	return nil
}
