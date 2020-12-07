package state

type CacheState struct {
	Hash          string
	Capacity      int64
	Filled        int64
	PiecesLength  int64
	PiecesCount   int
	DownloadSpeed float64
	Pieces        map[int]ItemState
}

type ItemState struct {
	Id        int
	Length    int64
	Size      int64
	Completed bool
}
