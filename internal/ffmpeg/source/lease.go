// Package source выдаёт временные доступы к файлам раздачи для ffmpeg.
package source

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

const (
	DefaultTTL = 10 * time.Minute
	IDBytes    = 16
)

// Lease описывает выданный доступ к файлу раздачи.
type Lease struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Hash      string    `json:"hash"`
	FileIdx   int       `json:"file_idx"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Expired сообщает, истёк ли доступ.
func (l *Lease) Expired(now time.Time) bool {
	return !now.Before(l.ExpiresAt)
}

// Registry хранит выданные доступы и следит за их сроком.
type Registry struct {
	mu     sync.Mutex
	leases map[string]*Lease
	ttl    time.Duration

	// now подменяется в тестах для детерминированного времени
	now func() time.Time
	// newID подменяется в тестах для предсказуемых идентификаторов
	newID func() string
}

// NewRegistry создаёт реестр доступов с указанным сроком жизни.
func NewRegistry(ttl time.Duration) *Registry {
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	return &Registry{
		leases: make(map[string]*Lease),
		ttl:    ttl,
		now:    time.Now,
		newID:  randomID,
	}
}

// Issue выдаёт новый доступ к файлу раздачи.
func (r *Registry) Issue(userID, hash string, fileIdx int) *Lease {
	now := r.now()
	lease := &Lease{
		ID:        r.newID(),
		UserID:    userID,
		Hash:      hash,
		FileIdx:   fileIdx,
		CreatedAt: now,
		ExpiresAt: now.Add(r.ttl),
	}

	r.mu.Lock()
	r.leases[lease.ID] = lease
	r.mu.Unlock()

	return lease
}

// Get возвращает доступ по идентификатору, если он ещё действителен.
func (r *Registry) Get(id string) (*Lease, bool) {
	if id == "" {
		return nil, false
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	lease, ok := r.leases[id]
	if !ok {
		return nil, false
	}
	if lease.Expired(r.now()) {
		delete(r.leases, id)
		return nil, false
	}
	return lease, true
}

// Touch продлевает срок действия доступа и сообщает, найден ли он.
func (r *Registry) Touch(id string) bool {
	if id == "" {
		return false
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	lease, ok := r.leases[id]
	if !ok {
		return false
	}
	now := r.now()
	if lease.Expired(now) {
		delete(r.leases, id)
		return false
	}
	lease.ExpiresAt = now.Add(r.ttl)
	return true
}

// Release отзывает доступ.
func (r *Registry) Release(id string) {
	r.mu.Lock()
	delete(r.leases, id)
	r.mu.Unlock()
}

// Cleanup убирает просроченные доступы и возвращает их число.
func (r *Registry) Cleanup() int {
	now := r.now()

	r.mu.Lock()
	defer r.mu.Unlock()

	removed := 0
	for id, lease := range r.leases {
		if lease.Expired(now) {
			delete(r.leases, id)
			removed++
		}
	}
	return removed
}

// Len возвращает число активных доступов.
func (r *Registry) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.leases)
}

// randomID создаёт непредсказуемый идентификатор доступа.
func randomID() string {
	buf := make([]byte, IDBytes)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
