package state

import (
	"server/torr/state"
	"server/torr/storage/reader"
)

type CacheState struct {
	Hash         string
	Capacity     int64
	Filled       int64
	PiecesLength int64
	PiecesCount  int
	Torrent      *state.TorrentStatus
	Pieces       map[int]ItemState
	Readers      []*reader.ReaderState
}

type ItemState struct {
	Id        int
	Length    int64
	Size      int64
	Completed bool
}
