package chess

import "math"

type King struct {
}

func (k *King) IsLegalPieceMove(move *Move, gameState *GameState) bool {
	dx := math.Abs(float64(move.PosTo.X - move.PosFrom.X))
	dy := math.Abs(float64(move.PosTo.Y - move.PosFrom.Y))

	if dx <= 1 && dy <= 1 && (dx != 0 || dy != 0) {
		return true
	}

	if isCastlingMove(move, gameState) {
		return true
	}

	return false
}

func isCastlingMove(move *Move, gameState *GameState) bool {
	if move.PosFrom.X != 4 {
		return false
	}

	var homeRank int
	if move.Player == 1 {
		homeRank = 0
	} else if move.Player == 2 {
		homeRank = 7
	} else {
		return false
	}

	if move.PosFrom.Y != homeRank || move.PosTo.Y != homeRank {
		return false
	}

	if move.PosTo.X != 6 && move.PosTo.X != 2 {
		return false
	}

	if gameState.MoveRecords.HasPieceMovedFrom(move.Player, move.PosFrom) {
		return false
	}

	var rookFrom Coordinates
	var rookTo Coordinates
	if move.PosTo.X == 6 {
		rookFrom = Coordinates{X: 7, Y: homeRank}
		rookTo = Coordinates{X: 5, Y: homeRank}
	} else {
		rookFrom = Coordinates{X: 0, Y: homeRank}
		rookTo = Coordinates{X: 3, Y: homeRank}
	}

	rookSquare := gameState.Board.GetSquareAt(rookFrom)
	if rookSquare.Owner != move.Player || rookSquare.Type != RookType {
		return false
	}

	if gameState.MoveRecords.HasPieceMovedFrom(move.Player, rookFrom) {
		return false
	}

	if !isPathClear(move.PosFrom, rookFrom, &gameState.Board) {
		return false
	}

	destSquare := gameState.Board.GetSquareAt(move.PosTo)
	if destSquare.Owner != 3 {
		return false
	}

	move.SpecialMoveEffect = &SpecialMoveEffect{
		PosFrom:         rookFrom,
		PosTo:           rookTo,
		SpecialMoveType: "castling",
		PieceType:       RookType,
	}

	return true
}

