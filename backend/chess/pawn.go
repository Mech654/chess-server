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

// GetMobility returns the unconditional mobility of the piece, not considering other pieces on the board
func (p *Pawn) GetMobility() Mobility {
	return Mobility{
		X: 0, Y: 1,
	}
}

func (p *Pawn) IsLegalPieceMove(move *Move, gameState *GameState) bool {
	// For pawn, we need a slightly different approach as its capture method
	//conflicts with its normal movement, so we need one to be true not all.
	valid := checkBasicMobility(p.GetMobility(), move, &gameState.Board)
	if valid {
		applyPromotionIfNeeded(move)
		return true
	}

	valid = checkBasicCapture(move, &gameState.Board)
	if valid {
		applyPromotionIfNeeded(move)
		return true
	}

	valid = checkEnPassant(move, gameState)
	return valid
}

func applyPromotionIfNeeded(move *Move) {
	if move.Player == 1 && move.PosTo.Y == 7 {
		move.SpecialMoveEffect = &SpecialMoveEffect{
			PosFrom:         move.PosTo,
			PosTo:           move.PosTo,
			SpecialMoveType: "promotion",
			PieceType:       QueenType,
		}
		return
	}

	if move.Player == 2 && move.PosTo.Y == 0 {
		move.SpecialMoveEffect = &SpecialMoveEffect{
			PosFrom:         move.PosTo,
			PosTo:           move.PosTo,
			SpecialMoveType: "promotion",
			PieceType:       QueenType,
		}
	}
}

func checkBasicMobility(mobility Mobility, move *Move, board *Board) bool {
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

		if dy == 2 && move.PosFrom.Y != 1 && isPathClear(move.PosFrom, move.PosTo, board) {
			return false
		}
	} else if move.Player == 2 {
		if dy != -mobility.Y && dy != -2 {
			return false
		}

		if dy == -2 && move.PosFrom.Y != 6 && isPathClear(move.PosFrom, move.PosTo, board) {
			return false
		}
	}

	return true
}

func checkBasicCapture(move *Move, board *Board) bool {
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

func checkEnPassant(move *Move, gameState *GameState) bool {
	destSquare := gameState.Board.GetSquareAt(move.PosTo)
	if destSquare.Owner != 3 {
		return false
	}

	dx := move.PosTo.X - move.PosFrom.X
	dy := move.PosTo.Y - move.PosFrom.Y

	isDiagonal := dx == 1 || dx == -1
	isCorrectDirection := (move.Player == 1 && dy == 1) || (move.Player == 2 && dy == -1)

	if !isDiagonal || !isCorrectDirection {
		return false
	}

	adjacentPos := Coordinates{X: move.PosTo.X, Y: move.PosFrom.Y}
	adjacentSquare := gameState.Board.GetSquareAt(adjacentPos)

	isEnemyPawn := adjacentSquare.Type == PawnType && adjacentSquare.Owner != move.Player
	if !isEnemyPawn {
		return false
	}

	latestMove := gameState.MoveRecords.GetLastMove()
	if latestMove == nil {
		return false
	}

	movedByCorrectPlayer := latestMove.Player == adjacentSquare.Owner
	movedToCorrectColumn := latestMove.PosFrom.X == move.PosTo.X
	if !movedByCorrectPlayer || !movedToCorrectColumn {
		return false
	}

	if move.Player == 1 {
		twoSquareAdvance := latestMove.PosFrom.Y == 6 && latestMove.PosTo.Y == 4
		correctCaptureRank := move.PosTo.Y == 5
		if !twoSquareAdvance || !correctCaptureRank {
			return false
		}
	} else if move.Player == 2 {
		twoSquareAdvance := latestMove.PosFrom.Y == 1 && latestMove.PosTo.Y == 3
		correctCaptureRank := move.PosTo.Y == 2
		if !twoSquareAdvance || !correctCaptureRank {
			return false
		}
	}

	move.SpecialMoveEffect = &SpecialMoveEffect{
		PosFrom:         adjacentPos,
		PosTo:           Coordinates{X: 99, Y: 99},
		SpecialMoveType: "enpassant",
		PieceType:       PawnType,
	}

	return true
}
