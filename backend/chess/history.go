package chess

type MoveRecords struct {
	records []Move
}

func (mr *MoveRecords) AddMove(move Move) {
	mr.records = append(mr.records, move)
}

func (mr *MoveRecords) GetLastMove() *Move {
	if len(mr.records) == 0 {
		return nil
	}
	return &mr.records[len(mr.records)-1]
}

func (mr *MoveRecords) GetAllMoves() []Move {
	return mr.records
}

func (mr *MoveRecords) HasPieceMovedFrom(player int, pos Coordinates) bool {
	for i := range mr.records {
		move := mr.records[i]
		if move.Player == player && move.PosFrom.X == pos.X && move.PosFrom.Y == pos.Y {
			return true
		}
	}
	return false
}
