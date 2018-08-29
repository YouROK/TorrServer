package settings

import (
	"encoding/binary"
	"fmt"

	"github.com/boltdb/bolt"
)

type Torrent struct {
	Name      string
	Magnet    string
	Hash      string
	Size      int64
	Timestamp int64

	Files []File
}

type File struct {
	Name   string
	Size   int64
	Viewed bool
}

func SetViewed(hash, filename string) error {
	err := openDB()
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		dbt := tx.Bucket(dbTorrentsName)
		if dbt == nil {
			return fmt.Errorf("could not find torrent")
		}
		hdb := dbt.Bucket([]byte(hash))
		if hdb == nil {
			return fmt.Errorf("could not find torrent")
		}

		fdb := hdb.Bucket([]byte("Files"))
		if fdb == nil {
			return fmt.Errorf("could not find torrent")
		}

		fdb = fdb.Bucket([]byte(filename))
		if fdb == nil {
			return fmt.Errorf("could not find torrent")
		}

		err = fdb.Put([]byte("Viewed"), []byte{1})
		if err != nil {
			return fmt.Errorf("error save torrent %v", err)
		}
		return nil
	})
}

func SaveTorrentDB(torrent *Torrent) error {
	err := openDB()
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		dbt, err := tx.CreateBucketIfNotExists(dbTorrentsName)
		if err != nil {
			return fmt.Errorf("could not create Torrents bucket: %v", err)
		}
		fmt.Println("Save torrent:", torrent.Name)
		hdb, err := dbt.CreateBucketIfNotExists([]byte(torrent.Hash))
		if err != nil {
			return fmt.Errorf("could not create Torrent bucket: %v", err)
		}

		err = hdb.Put([]byte("Name"), []byte(torrent.Name))
		if err != nil {
			return fmt.Errorf("error save torrent: %v", err)
		}
		err = hdb.Put([]byte("Link"), []byte(torrent.Magnet))
		if err != nil {
			return fmt.Errorf("error save torrent: %v", err)
		}
		err = hdb.Put([]byte("Size"), i2b(torrent.Size))
		if err != nil {
			return fmt.Errorf("error save torrent: %v", err)
		}
		err = hdb.Put([]byte("Timestamp"), i2b(torrent.Timestamp))
		if err != nil {
			return fmt.Errorf("error save torrent: %v", err)
		}

		fdb, err := hdb.CreateBucketIfNotExists([]byte("Files"))
		if err != nil {
			return fmt.Errorf("error save torrent files: %v", err)
		}

		for _, f := range torrent.Files {
			ffdb, err := fdb.CreateBucketIfNotExists([]byte(f.Name))
			if err != nil {
				return fmt.Errorf("error save torrent files: %v", err)
			}
			err = ffdb.Put([]byte("Size"), i2b(f.Size))
			if err != nil {
				return fmt.Errorf("error save torrent files: %v", err)
			}

			b := 0
			if f.Viewed {
				b = 1
			}

			err = ffdb.Put([]byte("Viewed"), []byte{byte(b)})
			if err != nil {
				return fmt.Errorf("error save torrent files: %v", err)
			}
		}

		return nil
	})
}

func RemoveTorrentDB(hash string) error {
	err := openDB()
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		dbt := tx.Bucket(dbTorrentsName)
		if dbt == nil {
			return fmt.Errorf("could not find torrent")
		}

		return dbt.DeleteBucket([]byte(hash))
	})
}

func LoadTorrentDB(hash string) (*Torrent, error) {
	err := openDB()
	if err != nil {
		return nil, err
	}

	var torr *Torrent
	err = db.View(func(tx *bolt.Tx) error {
		hdb := tx.Bucket(dbTorrentsName)
		if hdb == nil {
			return fmt.Errorf("could not find torrent")
		}
		hdb = hdb.Bucket([]byte(hash))
		if hdb != nil {
			torr = new(Torrent)
			torr.Hash = string(hash)
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

			fdb := hdb.Bucket([]byte("Files"))
			if fdb == nil {
				return fmt.Errorf("error load torrent files")
			}
			cf := fdb.Cursor()
			for fn, _ := cf.First(); fn != nil; fn, _ = cf.Next() {
				file := File{Name: string(fn)}
				ffdb := fdb.Bucket(fn)
				if ffdb == nil {
					return fmt.Errorf("error load torrent files")
				}

				tmp := ffdb.Get([]byte("Size"))
				if tmp == nil {
					return fmt.Errorf("error load torrent file")
				}
				file.Size = b2i(tmp)

				tmp = ffdb.Get([]byte("Viewed"))
				if tmp == nil {
					return fmt.Errorf("error load torrent file")
				}
				file.Viewed = len(tmp) > 0 && tmp[0] == 1
				torr.Files = append(torr.Files, file)
			}
			SortFiles(torr.Files)
		}
		return nil
	})
	return torr, err
}

func LoadTorrentsDB() ([]*Torrent, error) {
	err := openDB()
	if err != nil {
		return nil, err
	}

	torrs := make([]*Torrent, 0)
	err = db.View(func(tx *bolt.Tx) error {
		tdb := tx.Bucket(dbTorrentsName)
		c := tdb.Cursor()
		for h, _ := c.First(); h != nil; h, _ = c.Next() {
			hdb := tdb.Bucket(h)
			if hdb != nil {
				torr := new(Torrent)
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

				fdb := hdb.Bucket([]byte("Files"))
				if fdb == nil {
					return fmt.Errorf("error load torrent files")
				}
				cf := fdb.Cursor()
				for fn, _ := cf.First(); fn != nil; fn, _ = cf.Next() {
					file := File{Name: string(fn)}
					ffdb := fdb.Bucket(fn)
					if ffdb == nil {
						return fmt.Errorf("error load torrent files")
					}
					tmp := ffdb.Get([]byte("Size"))
					if tmp == nil {
						return fmt.Errorf("error load torrent file")
					}
					file.Size = b2i(tmp)

					tmp = ffdb.Get([]byte("Viewed"))
					if tmp == nil {
						return fmt.Errorf("error load torrent file")
					}
					file.Viewed = len(tmp) > 0 && tmp[0] == 1
					torr.Files = append(torr.Files, file)
				}
				SortFiles(torr.Files)
				torrs = append(torrs, torr)
			}
		}
		return nil
	})
	return torrs, err
}

func i2b(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func b2i(v []byte) int64 {
	return int64(binary.BigEndian.Uint64(v))
}
