package torrstor

import (
	"sync"

	"silo/internal/torrent/storage"

	"github.com/anacrolix/torrent/metainfo"
	ts "github.com/anacrolix/torrent/storage"
)

type Storage struct {
	storage.Storage

	cfg    *Config
	caches map[metainfo.Hash]*Cache
	mu     sync.Mutex
}

func NewStorage(cfg *Config) *Storage {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Storage{
		cfg:    cfg,
		caches: make(map[metainfo.Hash]*Cache),
	}
}

func (s *Storage) OpenTorrent(info *metainfo.Info, infoHash metainfo.Hash) (ts.TorrentImpl, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.caches[infoHash]; ok {
		return ch, nil
	}

	ch := NewCache(s.cfg.Capacity, s)
	ch.Init(info, infoHash)
	s.caches[infoHash] = ch

	return ch, nil
}

func (s *Storage) CloseHash(hash metainfo.Hash) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.caches == nil {
		return
	}
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
	s.caches = make(map[metainfo.Hash]*Cache)
	return nil
}

func (s *Storage) GetCache(hash metainfo.Hash) *Cache {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.caches[hash]
}
