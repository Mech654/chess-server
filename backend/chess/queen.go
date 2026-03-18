package chess

import "math"

type Queen struct {
}

func (q *Queen) IsLegalPieceMove(move *Move, gameState *GameState) bool {
	dx := math.Abs(float64(move.PosFrom.X - move.PosTo.X))
	dy := math.Abs(float64(move.PosFrom.Y - move.PosTo.Y))

	if dx != 0 && dy != 0 && dx != dy {
		return false
	}

	if !isPathClear(move.PosFrom, move.PosTo, &gameState.Board) {
		return false
	}

	return true
}
