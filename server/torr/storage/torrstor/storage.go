package torrstor

import (
	"sync"
	"slices"
	"os"
	"path/filepath"
	"server/log"
	"server/torr/storage"
	"server/settings"

	"github.com/anacrolix/torrent/metainfo"
	ts "github.com/anacrolix/torrent/storage"
)

type Storage struct {
	storage.Storage

	caches   map[metainfo.Hash]*Cache
	capacity int64
	mu       sync.Mutex
}

func NewStorage(capacity int64) *Storage {
	stor := new(Storage)
	stor.capacity = capacity
	stor.caches = make(map[metainfo.Hash]*Cache)
	return stor
}

func (s *Storage) OpenTorrent(info *metainfo.Info, infoHash metainfo.Hash) (ts.TorrentImpl, error) {
	// capFunc := func() (int64, bool) { //	NE
	// 	return s.capacity, true //	NE
	// } //	NE
	s.mu.Lock()
	defer s.mu.Unlock()
	ch := NewCache(s.capacity, s)
	ch.Init(info, infoHash)
	s.caches[infoHash] = ch
	return ch, nil //	OE
	// return ts.TorrentImpl{ //	NE
	// 	Piece:    ch.Piece, //	NE
	// 	Close:    ch.Close, //	NE
	// 	Capacity: &capFunc, //	NE
	// }, nil //	NE
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

func (s *Storage) GetCache(hash metainfo.Hash) *Cache {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cache, ok := s.caches[hash]; ok {
		return cache
	}
	return nil
}

func (s *Storage) cleanPieces() {
	s.mu.Lock()
	defer s.mu.Unlock()
	var filled int64 = 0
	for _, ch := range s.caches {
		filled += ch.filled
	}
	overfill := filled - s.capacity
	if overfill < 0 {
		return
	}
	sortCachesByModifiedDate := func(a, b *Cache) int {
		aname := filepath.Join(settings.BTsets.TorrentsSavePath, a.hash.HexString())
		bname := filepath.Join(settings.BTsets.TorrentsSavePath, b.hash.HexString())
		ainfo, err := os.Stat(aname)
		if err != nil {
			return -1
		}
		binfo, err := os.Stat(bname)
		if err != nil {
			return 1
		}
		return ainfo.ModTime().Compare(binfo.ModTime())
	}
	nonempty := slices.Values(slices.Collect(func(yield func(*Cache) bool) {
		for _, c := range s.caches {
			if c.filled > 0 {
				if !yield(c) {
					return
				}
			}
		}
	}))
	// fill sortedcaches with refs to non-empty caches sorted by respective
	//  folder modified date in descending order (older first)
	sortedcaches := slices.SortedFunc(nonempty, sortCachesByModifiedDate)
	for _, c := range sortedcaches {
		_cap := c.capacity
		c.capacity = max(c.filled - overfill, 0)
		overfill -= c.filled
		c.doCleanPieces()
		overfill += c.filled
		c.capacity = _cap
		if overfill <= 0 {
			return
		}
	}
}
