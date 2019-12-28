package storage

import (
	"server/torr/reader"
	"server/torr/storage/state"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

type Storage interface {
	storage.ClientImpl

	GetStats(hash metainfo.Hash) *state.CacheState
	CloseHash(hash metainfo.Hash)
	GetCache(hashes metainfo.Hash) Cache
}

type Cache interface {
	GetState() state.CacheState
	AddReader(r *reader.Reader)
	RemReader(r *reader.Reader)
	ReadersLen() int
	AdjustRA(readahead int64)
}
