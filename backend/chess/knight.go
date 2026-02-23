package chess

import "math"

type Knight struct {
}

func (k *Knight) IsLegalPieceMove(move Move, gameState *GameState) bool {
	dx := int(math.Abs(float64(move.PosTo.X - move.PosFrom.X)))
	dy := int(math.Abs(float64(move.PosTo.Y - move.PosFrom.Y)))

	if (dx == 2 && dy == 1) || (dx == 1 && dy == 2) {
		return true
	}

	return false
}
