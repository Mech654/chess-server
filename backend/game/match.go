package game

import (
	"math/rand"
	"time"

	"github.com/Mech654/chess-server/backend/chess"
	"github.com/Mech654/chess-server/backend/ws"
)

type Match struct {
	Player1    *ws.Client
	Player2    *ws.Client
	MatchState *MatchState
	CreatedAt  time.Time
}

type MatchState struct {
	WhitePlayer string
	Turn        string
	GameState   *chess.GameState
}

type MatchInvite struct {
	ID         uint64
	FromClient *ws.Client
	ToClient   *ws.Client
	CreatedAt  time.Time
}

type MatchHandler struct {
	match *Match
}

type MoveDTO struct {
	PosFrom [2]int `json:"pos_from"`
	PosTo   [2]int `json:"pos_to"`
}

func FindPlayerMatch(matchHandlers map[string]*MatchHandler, username string) *MatchHandler {
	return nil
}

func (mh *MatchHandler) HandleConnect(client *ws.Client) {
	//TODO: As of now, I will likely use this to implement re-connect if I
	// can figure that out. And another plan is match join link, so people
	// dont have to join my "intuitive" lobby.
}

func (mh *MatchHandler) HandleDisconnect(client *ws.Client) {
	//TODO: handle disconnect - forfeit or pause?
}

func (m *Match) Start() {
	if rand.Intn(2) == 1 {
		m.MatchState.WhitePlayer = m.Player1.Username
	} else {
		m.MatchState.WhitePlayer = m.Player2.Username
	}
	m.MatchState.Turn = m.MatchState.WhitePlayer

	m.Player1.Send(ws.EnvelopeMarshal("MATCH_START", map[string]string{
		"opponent":   m.Player2.Username,
		"first_move": m.MatchState.WhitePlayer,
	}))
	m.Player2.Send(ws.EnvelopeMarshal("MATCH_START", map[string]string{
		"opponent":   m.Player1.Username,
		"first_move": m.MatchState.WhitePlayer,
	}))

	for {
		time.Sleep(30 * time.Minute)
	}
}

func (m *Match) changeTurn() {
	if m.MatchState.Turn == m.Player1.Username {
		m.MatchState.Turn = m.Player2.Username
	} else {
		m.MatchState.Turn = m.Player1.Username
	}
}

func (m *MatchHandler) HandleMessage(client *ws.Client, data []byte) {
	var envelope ws.Envelope
	if err := ws.Unmarshal(data, &envelope); err != nil {
		return
	}

	if m.match.MatchState.Turn != client.Username {
		client.Send(ws.EnvelopeMarshal("ERROR", "Not your turn"))
		return
	}

	var moveDTO MoveDTO
	ws.Unmarshal(envelope.Data, &moveDTO)
	move := makeMove(client, m.match.MatchState.WhitePlayer, &moveDTO)

	if err := GenericMatchMoveValidation(move, &m.match.MatchState.GameState.Board); err != "" {
		client.Send(ws.EnvelopeMarshal("ERROR", err))
		return
	}

	valid := chess.IsLegalMove(*move, m.match.MatchState.GameState)
	if !valid {
		client.Send(ws.EnvelopeMarshal("ERROR", "Illegal Move"))
		return
	}

	if IsKingInCheck(&m.match.MatchState.GameState.Board, move.Player) {
		client.Send(ws.EnvelopeMarshal("ERROR", "Move would put king in check"))
		return
	}

	// Apply the move
	m.match.MatchState.GameState.ApplyMove(*move)
	m.match.changeTurn()
	clientSendUpdate(m.match, moveDTO)
}

func clientSendUpdate(match *Match, moveDTO MoveDTO) {
	if match.Player1.Username == match.MatchState.WhitePlayer {
		match.Player1.Send(ws.EnvelopeMarshal("MOVE_MADE", UpdateDTO{
			LastMove: moveDTO,
			Turnnow:  match.MatchState.Turn,
		}))
	} else {
		match.Player1.Send(ws.EnvelopeMarshal("MOVE_MADE", UpdateDTO{
			LastMove: *reverseMoveDTO(&moveDTO),
			Turnnow:  match.MatchState.Turn,
		}))
	}

	if match.Player2.Username == match.MatchState.WhitePlayer {
		match.Player2.Send(ws.EnvelopeMarshal("MOVE_MADE", UpdateDTO{
			LastMove: moveDTO,
			Turnnow:  match.MatchState.Turn,
		}))
	} else {
		match.Player2.Send(ws.EnvelopeMarshal("MOVE_MADE", UpdateDTO{
			LastMove: *reverseMoveDTO(&moveDTO),
			Turnnow:  match.MatchState.Turn,
		}))
	}
}

type UpdateDTO struct {
	LastMove MoveDTO `json:"last_move"`
	Turnnow  string  `json:"turn_now"`
}

func IsKingInCheck(board *chess.Board, player int) bool {
	kingPos := findKingPosition(board, player)
	opponent := 3 - player
	tempState := &chess.GameState{Board: *board}

	// Check all opponent pieces (max 16)
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			square := board[x][y]
			if square.Owner == opponent {
				attackMove := chess.Move{
					PosFrom: chess.Coordinates{X: x, Y: y},
					PosTo:   kingPos,
					Player:  opponent,
				}

				var attacker chess.Piece
				switch square.Type {
				case chess.PawnType:
					attacker = &chess.Pawn{}
				case chess.RookType:
					attacker = &chess.Rook{}
				case chess.KnightType:
					attacker = &chess.Knight{}
				case chess.BishopType:
					attacker = &chess.Bishop{}
				case chess.QueenType:
					attacker = &chess.Queen{}
				case chess.KingType:
					attacker = &chess.King{}
				}

				if attacker.IsLegalPieceMove(attackMove, tempState) {
					return true
				}
			}
		}
	}
	return false
}

func findKingPosition(board *chess.Board, player int) chess.Coordinates {
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			square := board[x][y]
			if square.Owner == player && square.Type == chess.KingType {
				return chess.Coordinates{X: x, Y: y}
			}
		}
	}

	return chess.Coordinates{X: -1, Y: -1}
}

func GenericMatchMoveValidation(move *chess.Move, board *chess.Board) string {
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

func makeMove(client *ws.Client, whitePlayer string, moveDTO *MoveDTO) *chess.Move {
	playerNum := 2
	if client.Username == whitePlayer {
		playerNum = 1
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
