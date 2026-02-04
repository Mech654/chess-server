package chess

type BoardSquare struct {
	Type  PieceType
	Owner int // 1,  2 or 3 for empty
}

type Board [8][8]*BoardSquare

func (b *Board) GetSquareAt(pos Coordinates) *BoardSquare {
	return b[pos.X][pos.Y]
}

type Piece interface {
	IsLegalPieceMove(move Move, board *Board) bool
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

func IsLegalMove(move Move, board *Board) bool {
	square := board.GetSquareAt(move.PosFrom)

	var validator Piece
	switch square.Type {
	case PawnType:
		validator = &Pawn{}
	// Add other piece types here
	default:
		return false
	}
	return validator.IsLegalPieceMove(move, board)
}

type Coordinates struct {
	X int
	Y int
}
