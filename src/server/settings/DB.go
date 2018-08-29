package settings

import (
	"fmt"
	"path/filepath"

	"github.com/boltdb/bolt"
)

var (
	db             *bolt.DB
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
	db, err = bolt.Open(filepath.Join(Path, "torrserver.db"), 0666, nil)
	if err != nil {
		fmt.Print(err)
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
