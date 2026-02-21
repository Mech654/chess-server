package chess

type BoardSquare struct {
	Type  PieceType
	Owner int // 1,  2 or 3 for empty
}

type Board [8][8]*BoardSquare

func (b *Board) GetSquareAt(pos Coordinates) *BoardSquare {
	return b[pos.X][pos.Y]
}

type GameState struct {
	Board       Board
	MoveRecords MoveRecords
}

type Piece interface {
	IsLegalPieceMove(move Move, gameState *GameState) bool
}

// Wanabe enum class
type PieceType string

const (
	PawnType   PieceType = "Pawn"
	RookType   PieceType = "Rook"
	KnightType PieceType = "Knight"
	BishopType PieceType = "Bishop"
	QueenType  PieceType = "Queen"
	KingType   PieceType = "King"
	EmptyType  PieceType = "Empty" // For empty squares, a bit sketcy idk
)

type Move struct {
	PosFrom      Coordinates
	PosTo        Coordinates
	Player       int // 1 for white, 2 for black
	MoveReversed bool
}

func IsLegalMove(move Move, gameState *GameState) bool {
	square := gameState.Board.GetSquareAt(move.PosFrom)

	var validator Piece
	switch square.Type {
	case PawnType:
		validator = &Pawn{}
	// Add other piece types here
	default:
		return false
	}
	return validator.IsLegalPieceMove(move, gameState)
}

type Coordinates struct {
	X int
	Y int
}

func NewBoard() Board {
	var board Board

	// Init empty squares
	for i := range board {
		for j := range board[i] {
			board[i][j] = &BoardSquare{Type: EmptyType, Owner: 3}
		}
	}

	// White pieces (player 1)
	board[0][0] = &BoardSquare{Type: RookType, Owner: 1}
	board[1][0] = &BoardSquare{Type: KnightType, Owner: 1}
	board[2][0] = &BoardSquare{Type: BishopType, Owner: 1}
	board[3][0] = &BoardSquare{Type: QueenType, Owner: 1}
	board[4][0] = &BoardSquare{Type: KingType, Owner: 1}
	board[5][0] = &BoardSquare{Type: BishopType, Owner: 1}
	board[6][0] = &BoardSquare{Type: KnightType, Owner: 1}
	board[7][0] = &BoardSquare{Type: RookType, Owner: 1}

	for i := range 8 {
		board[i][1] = &BoardSquare{Type: PawnType, Owner: 1}
	}

	// Black pieces (player 2)
	board[0][7] = &BoardSquare{Type: RookType, Owner: 2}
	board[1][7] = &BoardSquare{Type: KnightType, Owner: 2}
	board[2][7] = &BoardSquare{Type: BishopType, Owner: 2}
	board[3][7] = &BoardSquare{Type: QueenType, Owner: 2}
	board[4][7] = &BoardSquare{Type: KingType, Owner: 2}
	board[5][7] = &BoardSquare{Type: BishopType, Owner: 2}
	board[6][7] = &BoardSquare{Type: KnightType, Owner: 2}
	board[7][7] = &BoardSquare{Type: RookType, Owner: 2}

	for i := range 8 {
		board[i][6] = &BoardSquare{Type: PawnType, Owner: 2}
	}

	return board
}

func NewGameState() *GameState {
	return &GameState{
		Board:       NewBoard(),
		MoveRecords: MoveRecords{},
	}
}
