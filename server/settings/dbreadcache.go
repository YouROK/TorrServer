package settings

import (
	"server/log"
	"sync"
)

type DBReadCache struct {
	db             TorrServerDB
	listCache      map[string][]string
	dataCache      map[[2]string][]byte
	dataCacheMutex sync.RWMutex
}

func NewDBReadCache(db TorrServerDB) TorrServerDB {
	cdb := &DBReadCache{
		db:        db,
		listCache: map[string][]string{},
		dataCache: map[[2]string][]byte{},
	}
	return cdb
}

func (v *DBReadCache) CloseDB() {
	v.db.CloseDB()
	v.db = nil
	v.listCache = nil
	v.dataCache = nil
}

func (v *DBReadCache) Get(xPath, name string) []byte {
	cacheKey := v.makeDataCacheKey(xPath, name)
	v.dataCacheMutex.RLock()
	if data, ok := v.dataCache[cacheKey]; ok {
		defer v.dataCacheMutex.RUnlock()
		return data
	}
	v.dataCacheMutex.RUnlock()
	data := v.db.Get(xPath, name)
	v.dataCacheMutex.Lock()
	v.dataCache[cacheKey] = data
	v.dataCacheMutex.Unlock()
	return data
}

func (v *DBReadCache) Set(xPath, name string, value []byte) {
	if ReadOnly {
		log.TLogln("DB.Set: Read-only DB mode!", name)
		return
	}
	cacheKey := v.makeDataCacheKey(xPath, name)
	v.dataCacheMutex.Lock()
	v.dataCache[cacheKey] = value
	v.dataCacheMutex.Unlock()
	delete(v.listCache, xPath)
	v.db.Set(xPath, name, value)
}

func (v *DBReadCache) List(xPath string) []string {
	if names, ok := v.listCache[xPath]; ok {
		return names
	}
	names := v.db.List(xPath)
	v.listCache[xPath] = names
	return names
}

func (v *DBReadCache) Rem(xPath, name string) {
	if ReadOnly {
		log.TLogln("DB.Rem: Read-only DB mode!", name)
		return
	}
	cacheKey := v.makeDataCacheKey(xPath, name)
	delete(v.dataCache, cacheKey)
	delete(v.listCache, xPath)
	v.db.Rem(xPath, name)
}

func (v *DBReadCache) makeDataCacheKey(xPath, name string) [2]string {
	return [2]string{xPath, name}
}
