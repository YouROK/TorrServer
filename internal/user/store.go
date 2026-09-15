package user

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"

	"silo/internal/database"

	bolt "go.etcd.io/bbolt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrTokenAlreadyExists = errors.New("token already exists")
	ErrTokenNotFound      = errors.New("token not found")
	ErrTorrentNotFound    = errors.New("user torrent not found")
)

type Store struct {
	db *database.DB
}

func NewStore(db *database.DB) *Store {
	return &Store{db: db}
}

// ============================================================================
// ПОЛЬЗОВАТЕЛИ: CRUD и Индексы
// ============================================================================

// SaveUser создает или обновляет пользователя, обновляя индексы токена и логина
func (s *Store) SaveUser(u *User) error {
	return s.db.GetRawConn().Update(func(tx *bolt.Tx) error {
		bUsers := tx.Bucket(database.BucketUsers)
		bNames := tx.Bucket(database.BucketUsernames)
		bTokens := tx.Bucket(database.BucketUserTokens)

		// 1. Автоинкремент ID через NextSequence bbolt
		if u.ID == "" {
			seq, err := bUsers.NextSequence()
			if err != nil {
				return err
			}
			u.ID = fmt.Sprintf("u_%d", seq) // "u_1", "u_2" и т.д.
		}

		// 2. Проверка уникальности токена
		if u.APIToken != "" {
			if existingID := bTokens.Get([]byte(u.APIToken)); existingID != nil && string(existingID) != u.ID {
				return ErrTokenAlreadyExists
			}
		}

		// 3. Проверка уникальности логина
		if existingID := bNames.Get([]byte(u.Username)); existingID != nil && string(existingID) != u.ID {
			return ErrUserAlreadyExists
		}

		// 4. Зачищаем старые индексы при обновлении пользователя
		if oldBytes := bUsers.Get([]byte(u.ID)); oldBytes != nil {
			var oldUser User
			if err := json.Unmarshal(oldBytes, &oldUser); err == nil {
				if oldUser.Username != u.Username {
					_ = bNames.Delete([]byte(oldUser.Username))
				}
				if oldUser.APIToken != u.APIToken && oldUser.APIToken != "" {
					_ = bTokens.Delete([]byte(oldUser.APIToken))
				}
			}
		}

		// 5. Сериализуем и сохраняем
		data, err := json.Marshal(u)
		if err != nil {
			return err
		}

		if err := bUsers.Put([]byte(u.ID), data); err != nil {
			return err
		}

		// 6. Записываем обновленные индексы
		if err := bNames.Put([]byte(u.Username), []byte(u.ID)); err != nil {
			return err
		}
		if u.APIToken != "" {
			if err := bTokens.Put([]byte(u.APIToken), []byte(u.ID)); err != nil {
				return err
			}
		}

		return nil
	})
}

// GetUserByID находит пользователя по ID
func (s *Store) GetUserByID(id string) (*User, error) {
	var u User
	err := s.db.GetRawConn().View(func(tx *bolt.Tx) error {
		bUsers := tx.Bucket(database.BucketUsers)
		data := bUsers.Get([]byte(id))
		if data == nil {
			return ErrUserNotFound
		}
		return json.Unmarshal(data, &u)
	})
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByUsername находит пользователя по логину через индекс
func (s *Store) GetUserByUsername(username string) (*User, error) {
	var userID string
	err := s.db.GetRawConn().View(func(tx *bolt.Tx) error {
		bNames := tx.Bucket(database.BucketUsernames)
		id := bNames.Get([]byte(username))
		if id == nil {
			return ErrUserNotFound
		}
		userID = string(id)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetUserByID(userID)
}

// GetUserByToken находит пользователя по API-токену за O(1)
func (s *Store) GetUserByToken(token string) (*User, error) {
	var userID string
	err := s.db.GetRawConn().View(func(tx *bolt.Tx) error {
		bTokens := tx.Bucket(database.BucketUserTokens)
		id := bTokens.Get([]byte(token))
		if id == nil {
			return ErrTokenNotFound
		}
		userID = string(id)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetUserByID(userID)
}

// ListUsers возвращает всех пользователей
func (s *Store) ListUsers() ([]*User, error) {
	var users []*User
	err := s.db.GetRawConn().View(func(tx *bolt.Tx) error {
		b := tx.Bucket(database.BucketUsers)
		return b.ForEach(func(k, v []byte) error {
			var u User
			if err := json.Unmarshal(v, &u); err == nil {
				users = append(users, &u)
			}
			return nil
		})
	})
	return users, err
}

// DeleteUser удаляет пользователя, все его индексы и привязки к торрентам.
// Возвращает список orphanedHashes — торренты, которые больше никому не принадлежат.
func (s *Store) DeleteUser(userID string) ([]string, error) {
	var orphanedHashes []string

	err := s.db.GetRawConn().Update(func(tx *bolt.Tx) error {
		bUsers := tx.Bucket(database.BucketUsers)
		bNames := tx.Bucket(database.BucketUsernames)
		bTokens := tx.Bucket(database.BucketUserTokens)
		bUT := tx.Bucket(database.BucketUserTorrents)
		bRefs := tx.Bucket(database.BucketTorrentRefs)
		bTorrents := tx.Bucket(database.BucketTorrents)

		// 1. Получаем пользователя для очистки индексов
		userBytes := bUsers.Get([]byte(userID))
		if userBytes == nil {
			return ErrUserNotFound
		}
		var u User
		_ = json.Unmarshal(userBytes, &u)

		// 2. Сначала считываем торренты и ключи для безопасного удаления
		prefix := []byte(userID + ":")
		c := bUT.Cursor()

		var keysToDelete [][]byte
		var userTorrents []UserTorrent

		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var ut UserTorrent
			if err := json.Unmarshal(v, &ut); err == nil {
				userTorrents = append(userTorrents, ut)
			}
			// Клонируем байты ключа, так как память mmap изменится при удалении
			keysToDelete = append(keysToDelete, bytes.Clone(k))
		}

		// 3. Безопасно удаляем привязки и уменьшаем счетчики ссылок
		for _, k := range keysToDelete {
			if err := bUT.Delete(k); err != nil {
				return err
			}
		}

		for _, ut := range userTorrents {
			newRefCount := decrementRef(bRefs, ut.TorrentHash)
			if newRefCount <= 0 {
				_ = bTorrents.Delete([]byte(ut.TorrentHash))
				orphanedHashes = append(orphanedHashes, ut.TorrentHash)
			}
		}

		// 4. Удаляем индексы и самого пользователя
		_ = bNames.Delete([]byte(u.Username))
		if u.APIToken != "" {
			_ = bTokens.Delete([]byte(u.APIToken))
		}
		return bUsers.Delete([]byte(userID))
	})

	return orphanedHashes, err
}

// ============================================================================
// ПОЛЬЗОВАТЕЛЬСКИЕ ТОРРЕНТЫ (Связи и просмотры)
// ============================================================================

// AddUserTorrent связывает пользователя с торрентом и увеличивает счетчик ссылок
func (s *Store) AddUserTorrent(ut *UserTorrent) error {
	return s.db.GetRawConn().Update(func(tx *bolt.Tx) error {
		bUT := tx.Bucket(database.BucketUserTorrents)
		bRefs := tx.Bucket(database.BucketTorrentRefs)

		key := []byte(fmt.Sprintf("%s:%s", ut.UserID, ut.TorrentHash))
		isNew := bUT.Get(key) == nil

		data, err := json.Marshal(ut)
		if err != nil {
			return err
		}

		if err := bUT.Put(key, data); err != nil {
			return err
		}

		// Увеличиваем счетчик ссылок только при первом добавлении
		if isNew {
			incrementRef(bRefs, ut.TorrentHash)
		}

		return nil
	})
}

// GetUserTorrent получает привязку торрента пользователя
func (s *Store) GetUserTorrent(userID, hash string) (*UserTorrent, error) {
	var ut UserTorrent
	key := []byte(fmt.Sprintf("%s:%s", userID, hash))

	err := s.db.GetRawConn().View(func(tx *bolt.Tx) error {
		b := tx.Bucket(database.BucketUserTorrents)
		data := b.Get(key)
		if data == nil {
			return ErrTorrentNotFound
		}
		return json.Unmarshal(data, &ut)
	})
	if err != nil {
		return nil, err
	}
	return &ut, nil
}

// ListUserTorrents возвращает все торренты конкретного пользователя (префиксный поиск)
func (s *Store) ListUserTorrents(userID string) ([]*UserTorrent, error) {
	var list []*UserTorrent
	prefix := []byte(userID + ":")

	err := s.db.GetRawConn().View(func(tx *bolt.Tx) error {
		b := tx.Bucket(database.BucketUserTorrents)
		c := b.Cursor()

		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var ut UserTorrent
			if err := json.Unmarshal(v, &ut); err == nil {
				list = append(list, &ut)
			}
		}
		return nil
	})
	return list, err
}

// CountUserTorrents возвращает текущее число торрентов у пользователя (для проверки лимитов)
func (s *Store) CountUserTorrents(userID string) (int, error) {
	count := 0
	prefix := []byte(userID + ":")

	err := s.db.GetRawConn().View(func(tx *bolt.Tx) error {
		b := tx.Bucket(database.BucketUserTorrents)
		c := b.Cursor()
		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			count++
		}
		return nil
	})
	return count, err
}

// RemoveUserTorrent удаляет торрент у пользователя и уменьшает счетчик ссылок.
// Возвращает isLastOwner = true, если торрент больше никем не используется.
func (s *Store) RemoveUserTorrent(userID, hash string) (isLastOwner bool, err error) {
	key := []byte(fmt.Sprintf("%s:%s", userID, hash))

	err = s.db.GetRawConn().Update(func(tx *bolt.Tx) error {
		bUT := tx.Bucket(database.BucketUserTorrents)
		bRefs := tx.Bucket(database.BucketTorrentRefs)
		bTorrents := tx.Bucket(database.BucketTorrents)

		if bUT.Get(key) == nil {
			return ErrTorrentNotFound
		}

		if err := bUT.Delete(key); err != nil {
			return err
		}

		// Уменьшаем счетчик
		if decrementRef(bRefs, hash) <= 0 {
			isLastOwner = true
			_ = bTorrents.Delete([]byte(hash))
		}

		return nil
	})
	return isLastOwner, err
}

// SetFileViewedStatus отмечает или снимает отметку просмотра с конкретного файла в торренте
func (s *Store) SetFileViewedStatus(userID, hash string, fileIdx int, viewed bool) error {
	ut, err := s.GetUserTorrent(userID, hash)
	if err != nil {
		return err
	}

	if viewed {
		ut.MarkFileViewed(fileIdx)
	} else {
		ut.UnmarkFileViewed(fileIdx)
	}

	return s.AddUserTorrent(ut)
}

// ============================================================================
// Вспомогательные функции счетчиков ссылок (Reference Counting)
// ============================================================================

func incrementRef(b *bolt.Bucket, hash string) int {
	k := []byte(hash)
	val := b.Get(k)
	count := 0
	if val != nil {
		count = int(binary.BigEndian.Uint32(val))
	}
	count++
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(count))
	_ = b.Put(k, buf)
	return count
}

func decrementRef(b *bolt.Bucket, hash string) int {
	k := []byte(hash)
	val := b.Get(k)
	if val == nil {
		return 0
	}
	count := int(binary.BigEndian.Uint32(val))
	count--
	if count <= 0 {
		_ = b.Delete(k)
		return 0
	}
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(count))
	_ = b.Put(k, buf)
	return count
}
