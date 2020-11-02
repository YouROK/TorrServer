package settings

import (
	"fmt"
	"path/filepath"

	bolt "go.etcd.io/bbolt"
)

var (
	db             *bolt.DB
	dbViewedName   = []byte("Viewed")
	dbInfosName    = []byte("Infos")
	dbTorrentsName = []byte("Torrents")
	dbSettingsName = []byte("Settings")
	Path           string
)

func openDB() error {
	if db != nil {
		return nil
	}

	var err error
	var ro = Get().ReadOnlyMode
	db, err = bolt.Open(filepath.Join(Path, "torrserver.db"), 0666, &bolt.Options{ReadOnly: ro})
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists(dbSettingsName)
		if err != nil {
			return fmt.Errorf("could not create Settings bucket: %v", err)
		}
		_, err = tx.CreateBucketIfNotExists(dbTorrentsName)
		if err != nil {
			return fmt.Errorf("could not create Torrents bucket: %v", err)
		}
		return nil
	})
	if err != nil {
		CloseDB()
	}
	return err
}

func CloseDB() {
	if db != nil {
		db.Close()
		db = nil
	}
}
