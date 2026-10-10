package profile

import (
	"errors"
	"strings"
	"time"

	"silo/internal/bus"
	"silo/internal/log"
)

// Service управляет профилями транскодирования с учётом прав.
type Service struct {
	store *Store
	bus   *bus.Client
}

// NewService создаёт сервис профилей.
func NewService(store *Store) *Service {
	return &Service{
		store: store,
		bus:   bus.Get("transcode_profiles"),
	}
}

// List возвращает профили, доступные пользователю.
func (s *Service) List(actor Actor) ([]*Profile, error) {
	return s.store.Available(actor.ID, actor.Rank)
}

// ListAll возвращает все профили без фильтра по правам.
func (s *Service) ListAll(actor Actor) ([]*Profile, error) {
	if !actor.Owner {
		return nil, ErrPermission
	}
	return s.store.List()
}

// Get возвращает профиль, если он доступен пользователю.
func (s *Service) Get(actor Actor, id string) (*Profile, error) {
	p, err := s.store.Get(id)
	if err != nil {
		return nil, err
	}
	if !p.Available(p.OwnerID == actor.ID, actor.Rank) {
		return nil, ErrPermission
	}
	return p, nil
}

// Create создаёт профиль от имени пользователя.
func (s *Service) Create(actor Actor, p *Profile) (*Profile, error) {
	if !actor.CanCreate() {
		return nil, ErrPermission
	}
	if p == nil {
		return nil, errors.New("profile is required")
	}

	// Профиль по умолчанию меняет только владелец сервера
	if p.IsDefault && !actor.Owner {
		return nil, ErrPermission
	}

	created := *p
	created.ID = ""
	created.OwnerID = actor.ID
	created.CreatedAt = time.Time{}
	created.UpdatedAt = time.Time{}

	if err := s.store.Save(&created); err != nil {
		return nil, err
	}

	log.Infof("[FFmpeg] Profile '%s' created by %s", created.Name, actor.ID)
	s.bus.Emit("transcode:profile:created", &created)
	return &created, nil
}

// Update изменяет профиль.
func (s *Service) Update(actor Actor, id string, patch *Profile) (*Profile, error) {
	if patch == nil {
		return nil, errors.New("profile is required")
	}

	current, err := s.store.Get(id)
	if err != nil {
		return nil, err
	}
	if !actor.CanManage(current) {
		return nil, ErrPermission
	}

	// Профиль по умолчанию переключает только владелец сервера
	if patch.IsDefault != current.IsDefault && !actor.Owner {
		return nil, ErrPermission
	}

	updated := *patch
	updated.ID = current.ID
	updated.OwnerID = current.OwnerID
	updated.CreatedAt = current.CreatedAt

	if err := s.store.Save(&updated); err != nil {
		return nil, err
	}

	log.Infof("[FFmpeg] Profile '%s' updated by %s", updated.Name, actor.ID)
	s.bus.Emit("transcode:profile:updated", &updated)
	return &updated, nil
}

// Delete удаляет профиль.
func (s *Service) Delete(actor Actor, id string) error {
	if id == DefaultID {
		return ErrBuiltin
	}

	current, err := s.store.Get(id)
	if err != nil {
		return err
	}
	if !actor.CanManage(current) {
		return ErrPermission
	}

	if err := s.store.Delete(id); err != nil {
		return err
	}

	log.Infof("[FFmpeg] Profile '%s' deleted by %s", current.Name, actor.ID)
	s.bus.Emit("transcode:profile:deleted", map[string]string{
		"id":      id,
		"name":    current.Name,
		"user_id": actor.ID,
	})
	return nil
}

// SetDefault делает профиль используемым по умолчанию.
func (s *Service) SetDefault(actor Actor, id string) (*Profile, error) {
	if !actor.Owner {
		return nil, ErrPermission
	}

	p, err := s.store.Get(id)
	if err != nil {
		return nil, err
	}

	p.IsDefault = true
	if err := s.store.Save(p); err != nil {
		return nil, err
	}

	log.Infof("[FFmpeg] Profile '%s' is now default", p.Name)
	s.bus.Emit("transcode:profile:updated", p)
	return p, nil
}

// Default возвращает профиль по умолчанию.
func (s *Service) Default() *Profile {
	return s.store.Default()
}

// ResolveProfile выбирает профиль для запроса транскодирования.
func (s *Service) ResolveProfile(actor Actor, id string) (*Profile, error) {
	// Пустой идентификатор означает профиль по умолчанию
	if strings.TrimSpace(id) == "" || id == DefaultID {
		if p, ok := s.store.FindDefault(); ok {
			if p.Available(p.OwnerID == actor.ID, actor.Rank) {
				return p, nil
			}
		}
		builtin := Default()
		return &builtin, nil
	}

	return s.Get(actor, id)
}
