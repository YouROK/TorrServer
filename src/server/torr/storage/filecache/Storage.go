package filecache

import (
	"path/filepath"
	"sync"

	"server/settings"
	"server/torr/storage"
	"server/torr/storage/state"

	"github.com/anacrolix/missinggo/filecache"
	"github.com/anacrolix/torrent/metainfo"
	storage2 "github.com/anacrolix/torrent/storage"
)

type Storage struct {
	storage.Storage

	caches   map[metainfo.Hash]*filecache.Cache
	capacity int64
	mu       sync.Mutex
}

func NewStorage(capacity int64) storage.Storage {
	stor := new(Storage)
	stor.capacity = capacity
	stor.caches = make(map[metainfo.Hash]*filecache.Cache)
	return stor
}

func (s *Storage) OpenTorrent(info *metainfo.Info, infoHash metainfo.Hash) (storage2.TorrentImpl, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	path := filepath.Join(settings.Path, "cache", infoHash.String())
	cache, err := filecache.NewCache(path)
	if err != nil {
		return nil, err
	}
	cache.SetCapacity(s.capacity)
	s.caches[infoHash] = cache
	return storage2.NewResourcePieces(cache.AsResourceProvider()).OpenTorrent(info, infoHash)
}

func (s *Storage) GetStats(hash metainfo.Hash) *state.CacheState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return nil
}

func (s *Storage) Clean() {
	s.mu.Lock()
	defer s.mu.Unlock()
}

func (s *Storage) CloseHash(hash metainfo.Hash) {
	if s.caches == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

}

func (s *Storage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return nil
}
