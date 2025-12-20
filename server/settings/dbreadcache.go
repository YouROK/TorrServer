package settings

import (
	"sync"

	"server/log"
)

type DBReadCache struct {
	db             TorrServerDB
	listCache      map[string][]string
	listCacheMutex sync.RWMutex
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
	if v.dataCache == nil {
		return nil // или panic, или возвращаем ошибку
	}
	cacheKey := v.makeDataCacheKey(xPath, name)

	v.dataCacheMutex.RLock()
	if data, ok := v.dataCache[cacheKey]; ok {
		defer v.dataCacheMutex.RUnlock()
		return data
	}
	v.dataCacheMutex.RUnlock()

	// Если база данных закрыта, не пытаемся к ней обращаться
	if v.db == nil {
		return nil
	}
	data := v.db.Get(xPath, name)

	v.dataCacheMutex.Lock()
	if v.dataCache != nil { // Двойная проверка
		v.dataCache[cacheKey] = data
	}
	v.dataCacheMutex.Unlock()

	return data
}

func (v *DBReadCache) Set(xPath, name string, value []byte) {
	if ReadOnly {
		log.TLogln("DB.Set: Read-only DB mode!", name)
		return
	}
	// Проверяем, не закрыта ли база
	if v.dataCache == nil || v.db == nil {
		log.TLogln("DB.Set: no dataCache or DB is closed, cannot set", name)
		return
	}

	cacheKey := v.makeDataCacheKey(xPath, name)

	v.dataCacheMutex.Lock()
	if v.dataCache != nil { // Двойная проверка
		v.dataCache[cacheKey] = value
	}
	v.dataCacheMutex.Unlock()

	if v.listCache != nil {
		delete(v.listCache, xPath)
	}

	v.db.Set(xPath, name, value)
}

func (v *DBReadCache) List(xPath string) []string {
	if v.listCache == nil {
		return nil
	}

	v.listCacheMutex.RLock()
	if names, ok := v.listCache[xPath]; ok {
		defer v.listCacheMutex.RUnlock()
		return names
	}
	v.listCacheMutex.RUnlock()

	// Проверяем, не закрыта ли база
	if v.db == nil {
		return nil
	}

	names := v.db.List(xPath)

	v.listCacheMutex.Lock()
	if v.listCache != nil { // Двойная проверка
		v.listCache[xPath] = names
	}
	v.listCacheMutex.Unlock()

	return names
}

func (v *DBReadCache) Rem(xPath, name string) {
	if ReadOnly {
		log.TLogln("DB.Rem: Read-only DB mode!", name)
		return
	}
	// Проверяем, не закрыта ли база
	if v.dataCache == nil || v.db == nil {
		log.TLogln("DB.Rem: no dataCache or DB is closed, cannot remove", name)
		return
	}

	cacheKey := v.makeDataCacheKey(xPath, name)

	v.dataCacheMutex.Lock()
	if v.dataCache != nil {
		delete(v.dataCache, cacheKey)
	}
	v.dataCacheMutex.Unlock()

	if v.listCache != nil {
		delete(v.listCache, xPath)
	}

	v.db.Rem(xPath, name)
}

func (v *DBReadCache) makeDataCacheKey(xPath, name string) [2]string {
	return [2]string{xPath, name}
}
