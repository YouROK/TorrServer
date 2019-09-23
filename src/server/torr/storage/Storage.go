package storage

import (
	"server/torr/storage/state"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

type Storage interface {
	storage.ClientImpl

	GetStats(hash metainfo.Hash) *state.CacheState
	CloseHash(hash metainfo.Hash)
}
