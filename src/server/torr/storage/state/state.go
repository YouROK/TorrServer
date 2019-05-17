package state

type CacheState struct {
	Hash         string
	Capacity     int64
	Filled       int64
	PiecesLength int64
	PiecesCount  int
	Pieces       map[int]ItemState
}

type ItemState struct {
	Id         int
	Accessed   int64
	BufferSize int64
	Completed  bool
	Hash       string
}
