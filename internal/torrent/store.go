package torrent

import (
	"encoding/json"
	"errors"
	"fmt"

	"silo/internal/database"

	bolt "go.etcd.io/bbolt"
)

var (
	ErrTorrentNotFound = errors.New("torrent not found in database")
)

type Store struct {
	db *database.DB
}

func NewStore(db *database.DB) *Store {
	return &Store{db: db}
}

// Save сохраняет метаданные раздачи в базу данных
func (s *Store) Save(rec *TorrentRecord) error {
	return s.db.GetRawConn().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(database.BucketTorrents)
		data, err := json.Marshal(rec)
		if err != nil {
			return err
		}
		return b.Put([]byte(rec.Hash), data)
	})
}

// Get находит метаданные раздачи по хэшу
func (s *Store) Get(hash string) (*TorrentRecord, error) {
	var rec TorrentRecord
	err := s.db.GetRawConn().View(func(tx *bolt.Tx) error {
		b := tx.Bucket(database.BucketTorrents)
		data := b.Get([]byte(hash))
		if data == nil {
			return ErrTorrentNotFound
		}
		return json.Unmarshal(data, &rec)
	})
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// Delete удаляет карточку раздачи из базы
func (s *Store) Delete(hash string) error {
	return s.db.GetRawConn().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(database.BucketTorrents)
		return b.Delete([]byte(hash))
	})
}

// List возвращает все сохраненные раздачи
func (s *Store) List() ([]*TorrentRecord, error) {
	var list []*TorrentRecord
	err := s.db.GetRawConn().View(func(tx *bolt.Tx) error {
		b := tx.Bucket(database.BucketTorrents)
		return b.ForEach(func(k, v []byte) error {
			var rec TorrentRecord
			if err := json.Unmarshal(v, &rec); err == nil {
				list = append(list, &rec)
			}
			return nil
		})
	})
	return list, err
}

// GetConfig читает настройки торрент-движка из базы данных
func (s *Store) GetConfig() (*Config, error) {
	var cfg Config
	err := s.db.GetRawConn().View(func(tx *bolt.Tx) error {
		b := tx.Bucket(database.BucketSettings)
		data := b.Get([]byte("torrent_config"))
		if data == nil {
			return errors.New("not found")
		}
		return json.Unmarshal(data, &cfg)
	})

	if err != nil {
		// Если конфига еще нет в базе — возвращаем дефолтный
		return DefaultConfig(), nil
	}
	return &cfg, nil
}

// SaveConfig сохраняет конфиг движка в базу данных
func (s *Store) SaveConfig(cfg *Config) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal torrent config: %w", err)
	}
	return s.db.GetRawConn().Update(func(tx *bolt.Tx) error {
		return tx.Bucket(database.BucketSettings).Put([]byte("torrent_config"), data)
	})
}
