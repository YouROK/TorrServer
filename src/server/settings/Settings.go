package settings

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

var (
	sets      *Settings
	StartTime time.Time
)

func init() {
	sets = new(Settings)
	sets.CacheSize = 200 * 1024 * 1024
	sets.PreloadBufferSize = 20 * 1024 * 1024
	sets.ConnectionsLimit = 100
	sets.RetrackersMode = 1
	sets.DisableDHT = true
	StartTime = time.Now()
}

type Settings struct {
	CacheSize         int64 // in byte, def 200 mb
	PreloadBufferSize int64 // in byte, buffer for preload

	RetrackersMode int //0 - don`t add, 1 - add retrackers, 2 - remove retrackers

	//BT Config
	DisableTCP        bool
	DisableUTP        bool
	DisableUPNP       bool
	DisableDHT        bool
	DisableUpload     bool
	Encryption        int // 0 - Enable, 1 - disable, 2 - force
	DownloadRateLimit int // in kb, 0 - inf
	UploadRateLimit   int // in kb, 0 - inf
	ConnectionsLimit  int
}

func Get() *Settings {
	return sets
}

func (s *Settings) String() string {
	buf, _ := json.MarshalIndent(sets, "", " ")
	return string(buf)
}

func ReadSettings() error {
	err := openDB()
	if err != nil {
		return err
	}
	buf := make([]byte, 0)
	err = db.View(func(tx *bolt.Tx) error {
		sdb := tx.Bucket(dbSettingsName)
		if sdb == nil {
			return fmt.Errorf("error load settings")
		}

		buf = sdb.Get([]byte("json"))
		if buf == nil {
			return fmt.Errorf("error load settings")
		}
		return nil
	})
	err = json.Unmarshal(buf, sets)
	if err != nil {
		return err
	}
	if sets.ConnectionsLimit <= 0 {
		sets.ConnectionsLimit = 50
	}
	if sets.CacheSize <= 0 {
		sets.CacheSize = 200 * 1024 * 1024
	}
	return nil
}

func SaveSettings() error {
	err := openDB()
	if err != nil {
		return err
	}

	buf, err := json.Marshal(sets)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		setsDB, err := tx.CreateBucketIfNotExists(dbSettingsName)
		if err != nil {
			return err
		}
		return setsDB.Put([]byte("json"), []byte(buf))
	})
}
