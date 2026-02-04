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
	WhitePlayer string
	Turn        string
	Board       chess.Board
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
		m.MatchState.WhitePlayer = m.Player1.username
	} else {
		m.MatchState.WhitePlayer = m.Player2.username
	}
	m.MatchState.Turn = m.MatchState.WhitePlayer

	m.Player1.send <- HelperEnvelopeMarshal("MATCH_START", map[string]string{
		"opponent":   m.Player2.username,
		"first_move": m.MatchState.WhitePlayer,
	})
	m.Player2.send <- HelperEnvelopeMarshal("MATCH_START", map[string]string{
		"opponent":   m.Player1.username,
		"first_move": m.MatchState.WhitePlayer,
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
	move := makeMove(p, m.match.MatchState.WhitePlayer, &moveDTO)

	//Basic Match Move Validation
	if err := GenericMatchMoveValidation(move, m.match.MatchState.Board); err != "" {
		p.send <- HelperEnvelopeMarshal("ERROR", err)
		return
	}

	// Check Game Rules + Match Situation here
	valid := chess.IsLegalMove(*move, &m.match.MatchState.Board)
	if !valid {
		p.send <- HelperEnvelopeMarshal("ERROR", "Illegal Move")
		return
	}

}

func GenericMatchMoveValidation(move *chess.Move, board chess.Board) string {
	// Check if piece exists at source
	sourceSquare := board.GetSquareAt(move.PosFrom)
	if sourceSquare.Owner == 3 {
		return "No piece at source position"
	}

	// Check if destination is within bounds
	if move.PosTo.X < 0 || move.PosTo.X >= 8 || move.PosTo.Y < 0 || move.PosTo.Y >= 8 {
		return "Destination out of bounds"
	}

	// Check if moving player's own piece
	if move.Player != sourceSquare.Owner {
		return "Cannot move opponent's piece"
	}

	// Check if destination is occupied by player's own piece
	destSquare := board.GetSquareAt(move.PosTo)
	if destSquare.Owner == move.Player {
		return "Cannot capture your own piece"
	}

	return ""
}

func reverseMoveDTO(moveDTO *MoveDTO) *MoveDTO {
	return &MoveDTO{
		PosFrom: [2]int{7 - moveDTO.PosFrom[0], 7 - moveDTO.PosFrom[1]},
		PosTo:   [2]int{7 - moveDTO.PosTo[0], 7 - moveDTO.PosTo[1]},
	}
}

func makeMove(p *Player, whitePlayer string, moveDTO *MoveDTO) *chess.Move {
	// Determine player number (1=white, 2=black)
	playerNum := 2 // Black player
	if p.username == whitePlayer {
		playerNum = 1 // White player
	}

	reversed := false
	if playerNum == 2 {
		moveDTO = reverseMoveDTO(moveDTO)
		reversed = true
	}

	move := &chess.Move{
		PosFrom:      chess.Coordinates{X: moveDTO.PosFrom[0], Y: moveDTO.PosFrom[1]},
		PosTo:        chess.Coordinates{X: moveDTO.PosTo[0], Y: moveDTO.PosTo[1]},
		Player:       playerNum,
		MoveReversed: reversed,
	}

	return move
}
