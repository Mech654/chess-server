package game

import (
	"math/rand"
	"time"
)

type Match struct {
	Player1    *Player
	Player2    *Player
	match_info *MatchInfo
	created_at time.Time
}

type MatchInfo struct {
	FirstMove string
	Turn      string
}

type MatchInvite struct {
	ID          uint64
	from_player *Player
	to_player   *Player
	created_at  time.Time
}

type MatchHandler struct {
	parentMatch *Match
}

func (m *Match) Start() {
	matchHandler := &MatchHandler{
		parentMatch: m,
	}

	m.Player1.handler = matchHandler
	m.Player2.handler = matchHandler

	var FirstMove string
	if rand.Intn(2) == 1 {
		FirstMove = m.Player1.username
	} else {
		FirstMove = m.Player2.username
	}
	m.match_info.FirstMove = FirstMove
	m.match_info.Turn = FirstMove

	m.Player1.send <- HelperEnvelopeMarshal("MATCH_START", map[string]string{
		"opponent":   m.Player2.username,
		"first_move": m.match_info.FirstMove,
	})
	m.Player2.send <- HelperEnvelopeMarshal("MATCH_START", map[string]string{
		"opponent":   m.Player1.username,
		"first_move": m.match_info.FirstMove,
	})

	for {
		time.Sleep(30 * time.Minute)
	}
}

func (m *MatchHandler) HandleMessage(p *Player, data []byte) {
	//Entry point here
}
