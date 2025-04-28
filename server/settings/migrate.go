package settings

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"server/log"
	"server/web/api/utils"

	bolt "go.etcd.io/bbolt"
)

var dbTorrentsName = []byte("Torrents")

type torrentOldDB struct {
	Name      string
	Magnet    string
	InfoBytes []byte
	Hash      string
	Size      int64
	Timestamp int64
}

// Migrate from torrserver.db to config.db
func MigrateTorrents() {
	if _, err := os.Lstat(filepath.Join(Path, "torrserver.db")); os.IsNotExist(err) {
		return
	}

	db, err := bolt.Open(filepath.Join(Path, "torrserver.db"), 0o666, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		log.TLogln("MigrateTorrents", err)
		return
	}

	torrs := make([]*torrentOldDB, 0)
	err = db.View(func(tx *bolt.Tx) error {
		tdb := tx.Bucket(dbTorrentsName)
		if tdb == nil {
			return nil
		}
		c := tdb.Cursor()
		for h, _ := c.First(); h != nil; h, _ = c.Next() {
			hdb := tdb.Bucket(h)
			if hdb != nil {
				torr := new(torrentOldDB)
				torr.Hash = string(h)
				tmp := hdb.Get([]byte("Name"))
				if tmp == nil {
					return fmt.Errorf("error load torrent")
				}
				torr.Name = string(tmp)

				tmp = hdb.Get([]byte("Link"))
				if tmp == nil {
					return fmt.Errorf("error load torrent")
				}
				torr.Magnet = string(tmp)

				tmp = hdb.Get([]byte("Size"))
				if tmp == nil {
					return fmt.Errorf("error load torrent")
				}
				torr.Size = b2i(tmp)

				tmp = hdb.Get([]byte("Timestamp"))
				if tmp == nil {
					return fmt.Errorf("error load torrent")
				}
				torr.Timestamp = b2i(tmp)

				torrs = append(torrs, torr)
			}
		}
		return nil
	})
	db.Close()
	if err == nil && len(torrs) > 0 {
		for _, torr := range torrs {
			spec, err := utils.ParseLink(torr.Magnet)
			if err != nil {
				continue
			}

			title := torr.Name
			if len(spec.DisplayName) > len(title) {
				title = spec.DisplayName
			}
			log.TLogln("Migrate torrent", torr.Name, torr.Hash, torr.Size)
			AddTorrent(&TorrentDB{
				TorrentSpec: spec,
				Title:       title,
				Timestamp:   torr.Timestamp,
				Size:        torr.Size,
			})
		}
	}
	os.Remove(filepath.Join(Path, "torrserver.db"))
}

func b2i(v []byte) int64 {
	return int64(binary.BigEndian.Uint64(v))
}

/*
	=== MigrateToJson ===

Migrate 'Settings' and 'Viewed' buckets from BBolt ('config.db')
to separate JSON files ('settings.json' and 'viewed.json')

'Torrents' data continues to remain in the BBolt database ('config.db')
due to the fact that BLOBs are stored there

To make user be able to roll settings back, no data is deleted from 'config.db' file.
*/
func MigrateToJson(bboltDB, jsonDB TorrServerDB) error {
	var err error = nil

	const XPATH_SETTINGS = "Settings"
	const NAME_BITTORR = "BitTorr"
	const XPATH_VIEWED = "Viewed"

	if BTsets != nil {
		msg := "Migrate0002 MUST be called before initializing BTSets"
		log.TLogln(msg)
		os.Exit(1)
	}

	isByteArraysEqualJson := func(a, b []byte) (bool, error) {
		var objectA interface{}
		var objectB interface{}
		var err error = nil
		if err = json.Unmarshal(a, &objectA); err == nil {
			if err = json.Unmarshal(b, &objectB); err == nil {
				return reflect.DeepEqual(objectA, objectB), nil
			} else {
				err = fmt.Errorf("error unmashalling B: %s", err.Error())
			}
		} else {
			err = fmt.Errorf("error unmashalling A: %s", err.Error())
		}
		return false, err
	}

	migrateXPath := func(xPath, name string) error {
		if jsonDB.Get(xPath, name) == nil {
			bboltDBBlob := bboltDB.Get(xPath, name)
			if bboltDBBlob != nil {
				log.TLogln(fmt.Sprintf("Attempting to migrate %s->%s from TDB to JsonDB", xPath, name))
				jsonDB.Set(xPath, name, bboltDBBlob)
				jsonDBBlob := jsonDB.Get(xPath, name)
				if isEqual, err := isByteArraysEqualJson(bboltDBBlob, jsonDBBlob); err == nil {
					if isEqual {
						log.TLogln(fmt.Sprintf("Migrated %s->%s successful", xPath, name))
					} else {
						msg := fmt.Sprintf("Failed to migrate %s->%s TDB to JsonDB: equality check failed", xPath, name)
						log.TLogln(msg)
						return errors.New(msg)
					}
				} else {
					msg := fmt.Sprintf("Failed to migrate %s->%s TDB to JsonDB: %s", xPath, name, err)
					log.TLogln(msg)
					return errors.New(msg)
				}
			}
		}
		return nil
	}

	if err = migrateXPath(XPATH_SETTINGS, NAME_BITTORR); err != nil {
		return err
	}

	jsonDBViewedNames := jsonDB.List(XPATH_VIEWED)
	if len(jsonDBViewedNames) <= 0 {
		bboltDBViewedNames := bboltDB.List(XPATH_VIEWED)
		if len(bboltDBViewedNames) > 0 {
			for _, name := range bboltDBViewedNames {
				err = migrateXPath(XPATH_VIEWED, name)
				if err != nil {
					break
				}
			}
		}
	}
	return err
}
