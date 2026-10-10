package probe

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"silo/internal/ffmpeg/media"
)

// decode разбирает JSON-вывод ffprobe.
func decode(data []byte) (*result, error) {
	var res result
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}
	return &res, nil
}

// entry хранит результат разбора и время его создания.
type entry struct {
	info      *media.MediaInfo
	createdAt time.Time
}

// cache хранит результаты разбора в памяти.
type cache struct {
	mu    sync.Mutex
	items map[string]*entry
	order []string
	size  int
	ttl   time.Duration
}

func newCache(size int, ttl time.Duration) *cache {
	return &cache{
		items: make(map[string]*entry),
		size:  size,
		ttl:   ttl,
	}
}

// get возвращает запись кэша, если она ещё актуальна.
func (c *cache) get(key string) (*media.MediaInfo, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.items[key]
	if !ok {
		return nil, false
	}
	if c.ttl > 0 && time.Since(e.createdAt) > c.ttl {
		delete(c.items, key)
		c.removeOrder(key)
		return nil, false
	}
	return e.info, true
}

// put сохраняет результат разбора, вытесняя самые старые записи.
func (c *cache) put(key string, info *media.MediaInfo) {
	if info == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.items[key]; !ok {
		c.order = append(c.order, key)
	}
	c.items[key] = &entry{info: info, createdAt: time.Now()}

	for len(c.order) > c.size {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.items, oldest)
	}
}

// removeOrder убирает ключ из очереди вытеснения.
func (c *cache) removeOrder(key string) {
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			return
		}
	}
}
