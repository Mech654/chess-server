package chess

import "math"

type Bishop struct {
}

func (b *Bishop) IsLegalPieceMove(move Move, gameState *GameState) bool {
	if math.Abs(float64(move.PosFrom.X-move.PosTo.X)) != math.Abs(float64(move.PosFrom.Y-move.PosTo.Y)) {
		return false
	}

	if !isPathClear(move.PosFrom, move.PosTo, &gameState.Board) {
		return false
	}

	return true
}
