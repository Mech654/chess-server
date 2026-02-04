package chess

import (
	"math"
)

type Pawn struct {
	Mobility Mobility
}

type Mobility struct {
	X int
	Y int
}

// Only the unconditional mobility of the piece, not considering other pieces on the board
func (p *Pawn) GetMobility() Mobility {
	return Mobility{
		X: 0, Y: 1,
	}
}

// For pawn we need a slightly different approach as its capture method conflicts with its normal movement, so we need one to be true not all.
func (p *Pawn) IsLegalPieceMove(move Move, board *Board) bool {
	valid := checkBasicMobility(p.GetMobility(), move, board)
	if valid {
		return true
	}

	valid = checkCapture(move, board)
	return valid
}

func checkBasicMobility(mobility Mobility, move Move, board *Board) bool {
	destSquare := board.GetSquareAt(move.PosTo)
	if destSquare.Owner != 3 {
		return false
	}

	dx := move.PosTo.X - move.PosFrom.X
	dy := move.PosTo.Y - move.PosFrom.Y

	if dx != mobility.X {
		return false
	}
	if dy != mobility.Y {
		return false
	}
	return true
}

func checkCapture(move Move, board *Board) bool {
	destPiece := board[move.PosTo.X][move.PosTo.Y]
	if destPiece.Owner == 3 {
		return false
	}

	dx := move.PosTo.X - move.PosFrom.X
	dy := move.PosTo.Y - move.PosFrom.Y

	if int(math.Abs(float64(dx))) != 1 {
		return false
	}
	if dy != 1 {
		return false
	}
	return true
}
