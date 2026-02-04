package chess

type Piece struct {
	Type  PieceType
	Owner int // 1,  2 or 3 for empty
}

type Board [8][8]*Piece

type piece interface {
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
	PosFrom Coordinates
	PosTo   Coordinates
	Piece   *Piece
}

func (p *Piece) IsLegalPieceMove(move Move, board *Board) bool {
	switch p.Type {
	case PawnType:
		pawn := &Pawn{}
		return pawn.IsLegalPieceMove(move, board)
	// Add other piece types here
	default:
		return false
	}
}

type Coordinates struct {
	X int
	Y int
}
