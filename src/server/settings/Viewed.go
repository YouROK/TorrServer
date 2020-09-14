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
			return fmt.Errorf("could not find torrent")
		}
		hdb, err := dbt.CreateBucketIfNotExists([]byte(hash))
		if err != nil {
			return fmt.Errorf("could not find torrent")
		}

		fdb, err := hdb.CreateBucketIfNotExists([]byte("Files"))
		if err != nil {
			return fmt.Errorf("could not find torrent")
		}

		fdb, err = fdb.CreateBucketIfNotExists([]byte(filename))
		if err != nil {
			return fmt.Errorf("could not find torrent")
		}

		err = fdb.Put([]byte("Viewed"), []byte{1})
		if err != nil {
			return fmt.Errorf("error save torrent %v", err)
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
			return fmt.Errorf("could not find torrent")
		}
		hdb = hdb.Bucket([]byte(hash))
		if hdb != nil {
			fdb := hdb.Bucket([]byte("Files"))
			if fdb == nil {
				return fmt.Errorf("error load torrent files")
			}
			cf := fdb.Cursor()
			for fn, _ := cf.First(); fn != nil; fn, _ = cf.Next() {
				if string(fn) != filename {
					continue
				}

				ffdb := fdb.Bucket(fn)
				if ffdb == nil {
					return fmt.Errorf("error load torrent files")
				}

				tmp := ffdb.Get([]byte("Viewed"))
				if tmp == nil {
					return fmt.Errorf("error load torrent file")
				}
				if len(tmp) > 0 && tmp[0] == 1 {
					viewed = true
					break
				}
			}
		}
		return nil
	})
	return viewed
}

func GetViewedList() []struct {
	Hash string
	File string
} {
	err := openDB()
	if err != nil {
		return nil
	}
	viewed := false
	err = db.View(func(tx *bolt.Tx) error {
		hdb := tx.Bucket(dbViewedName)
		if hdb == nil {
			return fmt.Errorf("could not find torrent")
		}
		hdb = hdb.Bucket([]byte(hash))
		if hdb != nil {
			fdb := hdb.Bucket([]byte("Files"))
			if fdb == nil {
				return fmt.Errorf("error load torrent files")
			}
			cf := fdb.Cursor()
			for fn, _ := cf.First(); fn != nil; fn, _ = cf.Next() {
				if string(fn) != filename {
					continue
				}

				ffdb := fdb.Bucket(fn)
				if ffdb == nil {
					return fmt.Errorf("error load torrent files")
				}

				tmp := ffdb.Get([]byte("Viewed"))
				if tmp == nil {
					return fmt.Errorf("error load torrent file")
				}
				if len(tmp) > 0 && tmp[0] == 1 {
					viewed = true
					break
				}
			}
		}
		return nil
	})
	return viewed
}
