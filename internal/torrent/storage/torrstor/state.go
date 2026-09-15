package torrstor

// CacheState отражает текущее состояние RAM/дискового кэша раздачи
type CacheState struct {
	Capacity     int64             `json:"capacity"`
	Filled       int64             `json:"filled"`
	PiecesLength int64             `json:"piece_length"`
	PiecesCount  int               `json:"piece_count"`
	Hash         string            `json:"hash"`
	Pieces       map[int]ItemState `json:"pieces"`
	Readers      []*ReaderState    `json:"readers"`
}

type ItemState struct {
	Id        int   `json:"id"`
	Size      int64 `json:"size"`
	Length    int64 `json:"length"`
	Completed bool  `json:"completed"`
	Priority  int   `json:"priority"`
}

type ReaderState struct {
	Start  int `json:"start"`
	End    int `json:"end"`
	Reader int `json:"reader"`
}
