package memcache

import (
	"sync"

	"server/torr/storage"
	"server/torr/storage/state"

	"github.com/anacrolix/torrent/metainfo"
	storage2 "github.com/anacrolix/torrent/storage"
)

type Storage struct {
	storage.Storage

	caches   map[metainfo.Hash]*Cache
	capacity int64
	mu       sync.Mutex
}

func NewStorage(capacity int64) storage.Storage {
	stor := new(Storage)
	stor.capacity = capacity
	stor.caches = make(map[metainfo.Hash]*Cache)
	return stor
}

func (s *Storage) OpenTorrent(info *metainfo.Info, infoHash metainfo.Hash) (storage2.TorrentImpl, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch := NewCache(s.capacity, s)
	ch.Init(info, infoHash)
	s.caches[infoHash] = ch
	return ch, nil
}

func (s *Storage) GetStats(hash metainfo.Hash) *state.CacheState {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c, ok := s.caches[hash]; ok {
		st := c.GetState()
		return &st
	}
	return nil
}

func (s *Storage) CloseHash(hash metainfo.Hash) {
	if s.caches == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if ch, ok := s.caches[hash]; ok {
		ch.Close()
		delete(s.caches, hash)
	}
}

func (s *Storage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ch := range s.caches {
		ch.Close()
	}
	return nil
}
