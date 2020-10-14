package settings

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

func SetViewed(hash, filename string) error {
	err := openDB()
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		dbt, err := tx.CreateBucketIfNotExists(dbViewedName)
		if err != nil {
			return fmt.Errorf("error save viewed %v", err)
		}
		hdb, err := dbt.CreateBucketIfNotExists([]byte(hash))
		if err != nil {
			return fmt.Errorf("error save viewed %v", err)
		}

		err = hdb.Put([]byte(filename), []byte{1})
		if err != nil {
			return fmt.Errorf("error save viewed %v", err)
		}
		return nil
	})
}

func RemTorrentViewed(hash string) error {
	err := openDB()
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		dbt := tx.Bucket(dbViewedName)
		if dbt == nil {
			return nil
		}
		err = dbt.DeleteBucket([]byte(hash))
		if err == nil || err == bolt.ErrBucketNotFound {
			return nil
		}

		return err
	})
}

func GetViewed(hash, filename string) bool {
	err := openDB()
	if err != nil {
		return false
	}
	viewed := false
	err = db.View(func(tx *bolt.Tx) error {
		hdb := tx.Bucket(dbViewedName)
		if hdb == nil {
			return fmt.Errorf("error get viewed")
		}
		hdb = hdb.Bucket([]byte(hash))
		if hdb != nil {
			vw := hdb.Get([]byte(filename))
			viewed = vw != nil && vw[0] == 1
		}
		return nil
	})
	return viewed
}

func GetViewedList() map[string][]string {
	err := openDB()
	if err != nil {
		return nil
	}

	var list = make(map[string][]string)

	err = db.View(func(tx *bolt.Tx) error {
		rdb := tx.Bucket(dbViewedName)
		if rdb == nil {
			return fmt.Errorf("could not find torrent")
		}

		rdb.ForEach(func(hash, _ []byte) error {
			hdb := rdb.Bucket(hash)
			fdb := hdb.Bucket([]byte("Files"))
			fdb.ForEach(func(fileName, _ []byte) error {
				vw := fdb.Get(fileName)
				if vw != nil && vw[0] == 1 {
					list[string(hash)] = append(list[string(hash)], string(fileName))
				}
				return nil
			})
			return nil
		})
		return nil
	})
	return list
}
