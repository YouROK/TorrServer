package tgbot

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	tele "gopkg.in/telebot.v4"
	"server/rutor/models"
)

type searchCacheEntry struct {
	results []*models.TorrentDetails
	expires time.Time
}

var (
	searchCache   = make(map[int64]*searchCacheEntry)
	searchCacheMu sync.RWMutex
	cacheTTL      = 10 * time.Minute
	cacheMaxSize  = 1000
)

func init() {
	go searchCacheCleanup()
}

func searchCacheCleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		searchCacheMu.Lock()
		now := time.Now()
		for id, entry := range searchCache {
			if entry == nil || now.After(entry.expires) {
				delete(searchCache, id)
			}
		}
		if len(searchCache) > cacheMaxSize {
			evict := len(searchCache) - cacheMaxSize
			if evict < len(searchCache)/10 {
				evict = len(searchCache) / 10
			}
			if evict < 1 {
				evict = 1
			}
			type kv struct {
				id  int64
				exp time.Time
			}
			var entries []kv
			for id, entry := range searchCache {
				if entry != nil {
					entries = append(entries, kv{id, entry.expires})
				}
			}
			for evict > 0 && len(entries) > 0 {
				oldest := 0
				for j := 1; j < len(entries); j++ {
					if entries[j].exp.Before(entries[oldest].exp) {
						oldest = j
					}
				}
				delete(searchCache, entries[oldest].id)
				entries[oldest] = entries[len(entries)-1]
				entries = entries[:len(entries)-1]
				evict--
			}
		}
		searchCacheMu.Unlock()
	}
}

func storeSearchResults(userID int64, results []*models.TorrentDetails) {
	searchCacheMu.Lock()
	defer searchCacheMu.Unlock()
	searchCache[userID] = &searchCacheEntry{
		results: results,
		expires: time.Now().Add(cacheTTL),
	}
}

func getSearchResult(userID int64, index int) *models.TorrentDetails {
	searchCacheMu.Lock()
	defer searchCacheMu.Unlock()
	entry, ok := searchCache[userID]
	if !ok || entry == nil || time.Now().After(entry.expires) {
		return nil
	}
	if index < 0 || index >= len(entry.results) {
		return nil
	}
	return entry.results[index]
}

func getSearchResultsSlice(userID int64, offset, limit int) ([]*models.TorrentDetails, int) {
	searchCacheMu.Lock()
	defer searchCacheMu.Unlock()
	entry, ok := searchCache[userID]
	if !ok || entry == nil || time.Now().After(entry.expires) {
		return nil, 0
	}
	total := len(entry.results)
	if offset >= total {
		return nil, total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	slice := make([]*models.TorrentDetails, end-offset)
	copy(slice, entry.results[offset:end])
	return slice, total
}

func callbackSearchAdd(c tele.Context, indexStr string) error {
	uid := c.Sender().ID
	index, parseErr := strconv.Atoi(indexStr)
	if parseErr != nil {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "error")})
	}
	item := getSearchResult(uid, index)
	if item == nil {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "search_expired")})
	}
	link := item.Magnet
	if link == "" {
		link = item.Link
	}
	if link == "" {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "search_no_link")})
	}
	_ = c.Respond(&tele.CallbackResponse{Text: tr(uid, "search_adding")})
	if err := addTorrent(c, link); err != nil {
		return c.Send(fmt.Sprintf(tr(uid, "add_error"), err.Error()))
	}
	return list(c)
}
