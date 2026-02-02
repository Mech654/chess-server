package chess

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
func (p *Pawn) IsLegalPieceMove(move Move, board [8][8]*Piece) bool {
	valid := checkBasicMobility(p.GetMobility(), move, &board)
	if valid {
		return true
	}

	valid = checkCapture(move, &board)
	return valid
}

func checkBasicMobility(mobility Mobility, move Move, board *[8][8]*Piece) bool {
	dest_piece := (*board)[move.PosTo[0]][move.PosTo[1]]
	if dest_piece.Owner != 3 {
		return false
	}

	dx := move.PosTo[0] - move.PosFrom[0]
	dy := move.PosTo[1] - move.PosFrom[1]

	if dx != mobility.X {
		return false
	}
	if dy != mobility.Y {
		return false
	}
	return true
}

func checkCapture(move Move, board *[8][8]*Piece) bool {
	return true
}
