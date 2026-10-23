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

		// Автоинкремент ID через NextSequence bbolt
		if u.ID == "" {
			seq, err := bUsers.NextSequence()
			if err != nil {
				return err
			}
			u.ID = fmt.Sprintf("u_%d", seq) // "u_1", "u_2" и т.д.
		}

		// Проверка уникальности токена
		if u.APIToken != "" {
			if existingID := bTokens.Get([]byte(u.APIToken)); existingID != nil && string(existingID) != u.ID {
				return ErrTokenAlreadyExists
			}
		}

		// Проверка уникальности логина
		if existingID := bNames.Get([]byte(u.Username)); existingID != nil && string(existingID) != u.ID {
			return ErrUserAlreadyExists
		}

		// Зачистка старых индексов при обновлении пользователя
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

		// Сериализация и сохранение
		data, err := json.Marshal(u)
		if err != nil {
			return err
		}

		if err := bUsers.Put([]byte(u.ID), data); err != nil {
			return err
		}

		// Запись обновленных индексов
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
// Возвращает список orphanedHashes - торренты, которые больше никому не принадлежат.
func (s *Store) DeleteUser(userID string) ([]string, error) {
	var orphanedHashes []string

	err := s.db.GetRawConn().Update(func(tx *bolt.Tx) error {
		bUsers := tx.Bucket(database.BucketUsers)
		bNames := tx.Bucket(database.BucketUsernames)
		bTokens := tx.Bucket(database.BucketUserTokens)
		bUT := tx.Bucket(database.BucketUserTorrents)
		bRefs := tx.Bucket(database.BucketTorrentRefs)
		bTorrents := tx.Bucket(database.BucketTorrents)

		// Чтение пользователя для очистки индексов
		userBytes := bUsers.Get([]byte(userID))
		if userBytes == nil {
			return ErrUserNotFound
		}
		var u User
		_ = json.Unmarshal(userBytes, &u)

		// Сначала считываем торренты и ключи для безопасного удаления
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

		// Безопасное удаление привязок и уменьшение счетчиков ссылок
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

		// Удаление индексов и самого пользователя
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

// putUserTorrent записывает карточку и обновляет счетчик ссылок в текущей транзакции
func putUserTorrent(tx *bolt.Tx, ut *UserTorrent) error {
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
}

// countUserTorrents считает карточки пользователя в готовом бакете
func countUserTorrents(b *bolt.Bucket, userID string) int {
	count := 0
	prefix := []byte(userID + ":")

	c := b.Cursor()
	for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
		count++
	}
	return count
}

// AddUserTorrentLimited добавляет карточку с проверкой лимита.
// Проверка идёт в той же транзакции, что и запись: иначе параллельные
// добавления видят одно и то же число и обходят MaxTorrents.
func (s *Store) AddUserTorrentLimited(ut *UserTorrent, maxTorrents int) error {
	return s.db.GetRawConn().Update(func(tx *bolt.Tx) error {
		if maxTorrents > 0 {
			bUT := tx.Bucket(database.BucketUserTorrents)
			key := []byte(fmt.Sprintf("%s:%s", ut.UserID, ut.TorrentHash))

			// Обновление существующей карточки лимит не расходует
			if bUT.Get(key) == nil && countUserTorrents(bUT, ut.UserID) >= maxTorrents {
				return fmt.Errorf("torrent limit reached for your account (max: %d)", maxTorrents)
			}
		}

		return putUserTorrent(tx, ut)
	})
}

// updateUserTorrent меняет карточку внутри одной транзакции.
// Чтение и запись отдельными транзакциями теряют параллельные правки:
// обе стороны читают одно состояние и последняя запись затирает первую.
func (s *Store) updateUserTorrent(userID, hash string, mutate func(ut *UserTorrent) error) error {
	return s.db.GetRawConn().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(database.BucketUserTorrents)
		key := []byte(fmt.Sprintf("%s:%s", userID, hash))

		data := b.Get(key)
		if data == nil {
			return ErrTorrentNotFound
		}

		var ut UserTorrent
		if err := json.Unmarshal(data, &ut); err != nil {
			return err
		}

		if err := mutate(&ut); err != nil {
			return err
		}

		out, err := json.Marshal(&ut)
		if err != nil {
			return err
		}

		return b.Put(key, out)
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
	return s.updateUserTorrent(userID, hash, func(ut *UserTorrent) error {
		if viewed {
			ut.MarkFileViewed(fileIdx)
		} else {
			ut.UnmarkFileViewed(fileIdx)
		}
		return nil
	})
}

// SetAllFilesViewedStatus отмечает или снимает отметку у всех файлов раздачи
func (s *Store) SetAllFilesViewedStatus(userID, hash string, fileIdx []int, viewed bool) error {
	return s.updateUserTorrent(userID, hash, func(ut *UserTorrent) error {
		ut.SetAllFilesViewed(fileIdx, viewed)
		return nil
	})
}

// UpdateUserTorrentMeta правит личные данные карточки, не трогая отметки просмотра
func (s *Store) UpdateUserTorrentMeta(userID, hash, title, poster, category string) error {
	return s.updateUserTorrent(userID, hash, func(ut *UserTorrent) error {
		ut.Title = title
		ut.Poster = poster
		ut.Category = category
		return nil
	})
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
