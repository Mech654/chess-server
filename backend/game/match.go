package game

import (
	"math/rand"
	"time"

	"github.com/Mech654/chess-server/backend/chess"
)

type Match struct {
	Player1    *Player
	Player2    *Player
	MatchState *MatchState
	CreatedAt  time.Time
}

type MatchState struct {
	FirstMove string
	Turn      string
	Board     chess.Board
}

type MatchInvite struct {
	ID         uint64
	FromPlayer *Player
	ToPlayer   *Player
	CreatedAt  time.Time
}

type MatchHandler struct {
	match *Match
}

type MoveDTO struct {
	PosFrom [2]int `json:"pos_from"`
	PosTo   [2]int `json:"pos_to"`
}

func (m *Match) Start() {
	matchHandler := &MatchHandler{
		match: m,
	}

	m.Player1.handler = matchHandler
	m.Player2.handler = matchHandler

	if rand.Intn(2) == 1 {
		m.MatchState.FirstMove = m.Player1.username
	} else {
		m.MatchState.FirstMove = m.Player2.username
	}
	m.MatchState.Turn = m.MatchState.FirstMove

	m.Player1.send <- HelperEnvelopeMarshal("MATCH_START", map[string]string{
		"opponent":   m.Player2.username,
		"first_move": m.MatchState.FirstMove,
	})
	m.Player2.send <- HelperEnvelopeMarshal("MATCH_START", map[string]string{
		"opponent":   m.Player1.username,
		"first_move": m.MatchState.FirstMove,
	})

	for {
		time.Sleep(30 * time.Minute)
	}
}

// Entry Point here
func (m *MatchHandler) HandleMessage(p *Player, data []byte) {
	//Check If Turn
	if m.match.MatchState.Turn != p.username {
		p.send <- HelperEnvelopeMarshal("ERROR", "Not your turn")
		return
	}

	var moveDTO MoveDTO
	HelperUnmarshal(data, &moveDTO)

	move := &chess.Move{
		PosFrom: chess.Coordinates{X: moveDTO.PosFrom[0], Y: moveDTO.PosFrom[1]},
		PosTo:   chess.Coordinates{X: moveDTO.PosTo[0], Y: moveDTO.PosTo[1]},
		Piece:   m.match.MatchState.Board[moveDTO.PosFrom[0]][moveDTO.PosFrom[1]],
	}

	//Basic Match Move Validation

	// Check Game Rules + Match Situation here
	valid := move.Piece.IsLegalPieceMove(*move, &m.match.MatchState.Board)
	if !valid {
		p.send <- HelperEnvelopeMarshal("ERROR", "Illegal Move")
		return
	}

}

// I changed mind, No need to reverse the board when
// the only case it would even matter is pawn movement
// We will just translate the (x,y) coordinates with this function
// TODO: Remove comment
func reversal() {

}
