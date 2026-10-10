package profile

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"silo/internal/database"

	bolt "go.etcd.io/bbolt"
)

var (
	ErrNotFound   = errors.New("profile not found")
	ErrNameTaken  = errors.New("profile name is already used")
	ErrPermission = errors.New("permission denied")
	ErrBuiltin    = errors.New("builtin profile cannot be changed")
)

// Store хранит профили транскодирования в базе данных.
type Store struct {
	db *database.DB
}

// NewStore создаёт хранилище профилей.
func NewStore(db *database.DB) *Store {
	return &Store{db: db}
}

// Save сохраняет профиль, проверяя уникальность имени владельца.
func (s *Store) Save(p *Profile) error {
	if p == nil {
		return errors.New("profile is required")
	}
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("profile name is required")
	}

	p.Normalize()
	now := time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now

	return s.db.GetRawConn().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(database.BucketProfiles)

		if err := checkNameUnique(b, p); err != nil {
			return err
		}

		if p.ID == "" {
			seq, err := b.NextSequence()
			if err != nil {
				return err
			}
			p.ID = fmt.Sprintf("tp_%d", seq)
		}

		// Профиль по умолчанию может быть только один
		if p.IsDefault {
			if err := clearDefaults(b, p.ID); err != nil {
				return err
			}
		}

		data, err := json.Marshal(p)
		if err != nil {
			return err
		}
		return b.Put([]byte(p.ID), data)
	})
}

// Get находит профиль по идентификатору.
func (s *Store) Get(id string) (*Profile, error) {
	var p Profile

	err := s.db.GetRawConn().View(func(tx *bolt.Tx) error {
		data := tx.Bucket(database.BucketProfiles).Get([]byte(id))
		if data == nil {
			return ErrNotFound
		}
		return json.Unmarshal(data, &p)
	})
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// List возвращает все профили, отсортированные по имени.
func (s *Store) List() ([]*Profile, error) {
	var list []*Profile

	err := s.db.GetRawConn().View(func(tx *bolt.Tx) error {
		return tx.Bucket(database.BucketProfiles).ForEach(func(_, v []byte) error {
			var p Profile
			if err := json.Unmarshal(v, &p); err != nil {
				return nil
			}
			list = append(list, &p)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].IsDefault != list[j].IsDefault {
			return list[i].IsDefault
		}
		return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
	})
	return list, nil
}

// Available возвращает профили, доступные пользователю с указанным рангом.
func (s *Store) Available(userID string, rank int) ([]*Profile, error) {
	all, err := s.List()
	if err != nil {
		return nil, err
	}

	out := make([]*Profile, 0, len(all))
	for _, p := range all {
		if p.Available(p.OwnerID == userID, rank) {
			out = append(out, p)
		}
	}
	return out, nil
}

// Default возвращает профиль по умолчанию, а если его нет - встроенный.
func (s *Store) Default() *Profile {
	all, err := s.List()
	if err == nil {
		for _, p := range all {
			if p.IsDefault {
				return p
			}
		}
	}

	builtin := Default()
	return &builtin
}

// FindDefault возвращает профиль по умолчанию из базы.
func (s *Store) FindDefault() (*Profile, bool) {
	all, err := s.List()
	if err != nil {
		return nil, false
	}
	for _, p := range all {
		if p.IsDefault {
			return p, true
		}
	}
	return nil, false
}

// Delete удаляет профиль по идентификатору.
func (s *Store) Delete(id string) error {
	if id == DefaultID {
		return ErrBuiltin
	}

	return s.db.GetRawConn().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(database.BucketProfiles)
		if b.Get([]byte(id)) == nil {
			return ErrNotFound
		}
		return b.Delete([]byte(id))
	})
}

// checkNameUnique проверяет, что имя не занято другим профилем того же владельца.
func checkNameUnique(b *bolt.Bucket, p *Profile) error {
	name := strings.ToLower(strings.TrimSpace(p.Name))

	return b.ForEach(func(k, v []byte) error {
		if string(k) == p.ID {
			return nil
		}
		var other Profile
		if err := json.Unmarshal(v, &other); err != nil {
			return nil
		}
		if other.OwnerID == p.OwnerID && strings.ToLower(other.Name) == name {
			return ErrNameTaken
		}
		return nil
	})
}

// clearDefaults снимает признак по умолчанию со всех профилей, кроме указанного.
func clearDefaults(b *bolt.Bucket, keepID string) error {
	var updates [][]byte

	err := b.ForEach(func(k, v []byte) error {
		if string(k) == keepID {
			return nil
		}
		var p Profile
		if err := json.Unmarshal(v, &p); err != nil || !p.IsDefault {
			return nil
		}
		p.IsDefault = false

		data, err := json.Marshal(&p)
		if err != nil {
			return nil
		}
		updates = append(updates, append([]byte(nil), k...), data)
		return nil
	})
	if err != nil {
		return err
	}

	for i := 0; i < len(updates); i += 2 {
		if err := b.Put(updates[i], updates[i+1]); err != nil {
			return err
		}
	}
	return nil
}
