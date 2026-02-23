package chess

import "math"

type King struct {
}

func (k *King) IsLegalPieceMove(move Move, gameState *GameState) bool {
	dx := math.Abs(float64(move.PosTo.X - move.PosFrom.X))
	dy := math.Abs(float64(move.PosTo.Y - move.PosFrom.Y))

	if dx <= 1 && dy <= 1 && (dx != 0 || dy != 0) {
		return true
	}

	return false
}
