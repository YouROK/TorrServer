package filecache

import (
	"server/torr/reader"
	"server/torr/storage"
	"server/torr/storage/state"
	"sync"

	"github.com/anacrolix/missinggo/filecache"
)

type Cache struct {
	*filecache.Cache
	cache2   storage.Cache
	muReader sync.Mutex
	readers  map[*reader.Reader]struct{}
}

func (cache *Cache) GetState() state.CacheState {
	cState := state.CacheState{}
	return cState
}

func NewCache(root string) (ret *Cache, err error) {
	c, e := filecache.NewCache(root)
	return &Cache{Cache: c,
		readers: make(map[*reader.Reader]struct{})}, e
}

func (c *Cache) AddReader(r *reader.Reader) {
	c.muReader.Lock()
	defer c.muReader.Unlock()
	c.readers[r] = struct{}{}
}

func (c *Cache) RemReader(r *reader.Reader) {
	c.muReader.Lock()
	defer c.muReader.Unlock()
	delete(c.readers, r)
}

func (c *Cache) ReadersLen() int {
	if c == nil || c.readers == nil {
		return 0
	}
	return len(c.readers)
}

func (c *Cache) AdjustRA(readahead int64) {
	for r, _ := range c.readers {
		r.SetReadahead(readahead)
	}
}
