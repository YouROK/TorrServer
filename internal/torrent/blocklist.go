package torrent

import (
	"bufio"
	"io"
	"net"
	"sync"

	"silo/internal/log"

	"github.com/anacrolix/torrent/iplist"
)

// DynamicBlocklist позволяет на лету подменять список блокировки без перезапуска клиента
type DynamicBlocklist struct {
	mu     sync.RWMutex
	ranger iplist.Ranger
}

func NewDynamicBlocklist(initial iplist.Ranger) *DynamicBlocklist {
	return &DynamicBlocklist{ranger: initial}
}

func (d *DynamicBlocklist) Lookup(ip net.IP) (iplist.Range, bool) {
	if d == nil {
		return iplist.Range{}, false
	}
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.ranger == nil {
		return iplist.Range{}, false
	}
	return d.ranger.Lookup(ip)
}

func (d *DynamicBlocklist) NumRanges() int {
	if d == nil {
		return 0
	}
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.ranger == nil {
		return 0
	}
	return d.ranger.NumRanges()
}

func (d *DynamicBlocklist) Set(r iplist.Ranger) {
	if d == nil {
		return
	}
	d.mu.Lock()
	d.ranger = r
	d.mu.Unlock()
	log.Info("[Torrent Engine] IP blocklist successfully updated in memory")
}

// ParseBlocklistP2P парсит P2P-блоклист из любого потока (строка от плагина, ответ по сети и т.д.)
func ParseBlocklistP2P(r io.Reader) (iplist.Ranger, error) {
	scanner := bufio.NewScanner(r)
	var ranges []iplist.Range

	for scanner.Scan() {
		line := scanner.Bytes()
		rng, ok, err := iplist.ParseBlocklistP2PLine(line)
		if err != nil {
			continue
		}
		if ok {
			ranges = append(ranges, rng)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(ranges) == 0 {
		return nil, nil
	}

	log.Infof("[Torrent Engine] Loaded %d IP blocklist ranges", len(ranges))
	return iplist.New(ranges), nil
}
