package game

import (
	"math/rand"
	"time"

	"github.com/Mech654/chess-server/backend/chess"
)

type Match struct {
	Player1     *Player
	Player2     *Player
	match_state *MatchState
	created_at  time.Time
}

type MatchState struct {
	FirstMove string
	Turn      string
	Board     [8][8]*chess.Piece
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

type MoveDTO struct {
	PosFrom [2]int `json:"pos_from"`
	PosTo   [2]int `json:"pos_to"`
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
	m.match_state.FirstMove = FirstMove
	m.match_state.Turn = FirstMove

	m.Player1.send <- HelperEnvelopeMarshal("MATCH_START", map[string]string{
		"opponent":   m.Player2.username,
		"first_move": m.match_state.FirstMove,
	})
	m.Player2.send <- HelperEnvelopeMarshal("MATCH_START", map[string]string{
		"opponent":   m.Player1.username,
		"first_move": m.match_state.FirstMove,
	})

	for {
		time.Sleep(30 * time.Minute)
	}
}

// Entry Point here
func (m *MatchHandler) HandleMessage(p *Player, data []byte) {
	//Check If Turn
	if m.parentMatch.match_state.Turn != p.username {
		p.send <- HelperEnvelopeMarshal("ERROR", "Not your turn")
		return
	}

	var moveDTO MoveDTO
	HelperUnmarshal(data, &moveDTO)

	move := &chess.Move{
		PosFrom: moveDTO.PosFrom,
		PosTo:   moveDTO.PosTo,
		Piece:   m.parentMatch.match_state.Board[moveDTO.PosFrom[0]][moveDTO.PosFrom[1]],
	}

	//Basic Match Move Validation

	// Check Game Rules + Match Situation here
	valid := move.Piece.IsLegalPieceMove(*move, m.parentMatch.match_state.Board)
	if !valid {
		p.send <- HelperEnvelopeMarshal("ERROR", "Illegal Move")
		return
	}

}

func Reversal(arr [8][8]*chess.Piece) [8][8]*chess.Piece {
	var newArr [8][8]*chess.Piece
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			newArr[7-i][7-j] = arr[i][j]
		}
	}
	return newArr
}
