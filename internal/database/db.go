package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	BucketTorrents     = []byte("torrents")      // Глобальные карточки раздач
	BucketSettings     = []byte("settings")      // Динамические настройки
	BucketUsers        = []byte("users")         // Учетные записи (ID -> User)
	BucketUsernames    = []byte("user_names")    // Индекс: Username -> UserID (для входа)
	BucketUserTokens   = []byte("user_tokens")   // Индекс: APIToken -> UserID (для плееров)
	BucketUserTorrents = []byte("user_torrents") // Связи: UserID:TorrentHash -> UserTorrent
	BucketTorrentRefs  = []byte("torrent_refs")  // Счетчики ссылок: TorrentHash -> int
	BucketPlugins      = []byte("plugins")
	BucketPluginData   = []byte("plugin_data")
)

type DB struct {
	conn *bolt.DB
}

func Open(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("не удалось создать папку для БД: %w", err)
	}

	conn, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть БД (%s): %w", dbPath, err)
	}

	db := &DB{conn: conn}
	if err := db.initBuckets(); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return db, nil
}

func (d *DB) initBuckets() error {
	return d.conn.Update(func(tx *bolt.Tx) error {
		buckets := [][]byte{
			BucketTorrents,
			BucketSettings,
			BucketUsers,
			BucketUsernames,
			BucketUserTokens,
			BucketUserTorrents,
			BucketTorrentRefs,
			BucketPlugins,
			BucketPluginData,
		}
		for _, b := range buckets {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return fmt.Errorf("ошибка создания бакета %s: %w", string(b), err)
			}
		}
		return nil
	})
}

func (d *DB) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

func (d *DB) GetRawConn() *bolt.DB {
	return d.conn
}
