package chess

type Piece struct {
	Type  PieceType
	Owner int // 1,  2 or 3 for empty
}

type piece interface {
	IsLegalPieceMove(move Move, board [8][8]*Piece) bool
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
	PosFrom [2]int
	PosTo   [2]int
	Piece   *Piece
}

func (p *Piece) IsLegalPieceMove(move Move, board [8][8]*Piece) bool {
	switch p.Type {
	case PawnType:
		pawn := &Pawn{}
		return pawn.IsLegalPieceMove(move, board)
	// Add other piece types here
	default:
		return false
	}
}
