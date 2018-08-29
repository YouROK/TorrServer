package settings

import (
	"fmt"
	"strings"

	"github.com/boltdb/bolt"
)

func AddInfo(hash, info string) error {
	err := openDB()
	if err != nil {
		return err
	}

	hash = strings.ToUpper(hash)
	return db.Update(func(tx *bolt.Tx) error {
		dbt, err := tx.CreateBucketIfNotExists([]byte(dbInfosName))
		if err != nil {
			return err
		}

		dbi, err := dbt.CreateBucketIfNotExists([]byte(hash))
		if err != nil {
			return err
		}

		err = dbi.Put([]byte("Info"), []byte(info))
		if err != nil {
			return fmt.Errorf("error save torrent info %v", err)
		}
		return nil
	})
}

func GetInfo(hash string) string {
	err := openDB()
	if err != nil {
		return "{}"
	}

	hash = strings.ToUpper(hash)
	ret := "{}"
	err = db.View(func(tx *bolt.Tx) error {
		hdb := tx.Bucket(dbInfosName)
		if hdb == nil {
			return fmt.Errorf("could not find torrent info")
		}
		hdb = hdb.Bucket([]byte(hash))
		if hdb != nil {
			info := hdb.Get([]byte("Info"))
			if info == nil {
				return fmt.Errorf("error get torrent info")
			}
			ret = string(info)
			return nil
		}
		return nil
	})
	return ret
}
