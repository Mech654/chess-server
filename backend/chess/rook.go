package chess

type Rook struct {
}

func (r *Rook) IsLegalPieceMove(move Move, gameState *GameState) bool {
	if move.PosFrom.X != move.PosTo.X && move.PosFrom.Y != move.PosTo.Y {
		return false
	}

	if !isPathClear(move.PosFrom, move.PosTo, &gameState.Board) {
		return false
	}

	return true
}
