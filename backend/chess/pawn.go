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
func (p *Pawn) IsLegalPieceMove(move Move, gameState *GameState) bool {
	valid := checkBasicMobility(p.GetMobility(), move, &gameState.Board)
	if valid {
		return true
	}

	valid = checkBasicCapture(move, &gameState.Board)
	if valid {
		return true
	}

	valid = checkEnpassant(move, gameState)
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

	if move.Player == 1 {
		if dy != mobility.Y && dy != 2 {
			return false
		}

		if dy == 2 && move.PosFrom.Y != 1 {
			return false
		}
	} else if move.Player == 2 {
		if dy != -mobility.Y && dy != -2 {
			return false
		}

		if dy == -2 && move.PosFrom.Y != 6 {
			return false
		}
	}

	return true
}

func checkBasicCapture(move Move, board *Board) bool {
	destSquare := board.GetSquareAt(move.PosTo)
	if destSquare.Owner == 3 {
		return false
	}

	dx := move.PosTo.X - move.PosFrom.X
	dy := move.PosTo.Y - move.PosFrom.Y

	if int(math.Abs(float64(dx))) != 1 {
		return false
	}

	if move.Player == 1 {
		if dy != 1 {
			return false
		}
	} else if move.Player == 2 {
		if dy != -1 {
			return false
		}
	}

	return true
}

func checkEnpassant(move Move, gameState *GameState) bool {
	destSquare := gameState.Board.GetSquareAt(move.PosTo)
	if destSquare.Owner != 3 {
		return false
	}

	dx := move.PosTo.X - move.PosFrom.X
	dy := move.PosTo.Y - move.PosFrom.Y

	if dx != 1 && dx != -1 {
		return false
	}

	if (dy != 1 && move.Player != 1) {
		return false
	}

	if (dy != -1 && move.Player != 2){
		return false
	}

	adjacentSquare := gameState.Board.GetSquareAt(Coordinates{X: move.PosTo.X, Y: move.PosFrom.Y})
	if adjacentSquare.Type != PawnType {
		return false
	}

	if adjacentSquare.Owner == move.Player {
		return false
	}

	latestMove := gameState.MoveRecords.GetLastMove()
	if latestMove == nil {
		return false
	}

	if latestMove.Player != adjacentSquare.Owner {
		return false
	}

	if latestMove.PosFrom.X != move.PosTo.X {
		return false
	}

	if move.Player == 1 {
		if latestMove.PosFrom.Y != 6 || latestMove.PosTo.Y != 4 {
			return false
		}

		if move.PosTo.Y != 5 {
			return false
		}

	} else if move.Player == 2 {
		if latestMove.PosFrom.Y != 1 || latestMove.PosTo.Y != 3 {
			return false
		}

		if move.PosTo.Y != 2 {
			return false
		}
	}

	return true
}

