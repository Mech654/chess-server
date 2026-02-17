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

// I got an intuition like a search related
// receiver is crucial, but I dont know why yet
// So additional will come as needed
