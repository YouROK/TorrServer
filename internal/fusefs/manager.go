package fusefs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"silo/internal/log"
)

// mountServer - смонтированная файловая система.
type mountServer interface {
	Unmount() error
	// done закрывается, когда сервер прекращает обслуживать запросы.
	done() <-chan struct{}
}

// Manager владеет одним монтированием источника. Повторное монтирование
// отклоняется: размонтировать нужно явно.
type Manager struct {
	mu   sync.Mutex
	src  FS
	srv  mountServer
	info Status
}

func NewManager() *Manager {
	return &Manager{}
}

// SetSource задаёт источник данных для последующих монтирований.
func (m *Manager) SetSource(src FS) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.src = src
}

// Supported сообщает, доступно ли монтирование на текущей платформе.
func Supported() bool { return platformSupported() }

// Mount монтирует источник в каталог path.
func (m *Manager) Mount(path string, opts Options) (Status, error) {
	status, srv, err := m.mount(path, opts)
	if err != nil {
		return Status{}, err
	}

	// Наблюдение запускается после снятия блокировки: обратный вызов
	// берёт её заново.
	done := srv.done()
	if done != nil {
		go func() {
			<-done
			m.clearExternal(srv)
		}()
	}
	return status, nil
}

func (m *Manager) mount(path string, opts Options) (Status, mountServer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !platformSupported() {
		return Status{}, nil, ErrUnsupported
	}

	// Путь проверяется раньше источника: ошибка в пути - это ошибка
	// пользователя, и она понятнее, чем отсутствие источника.
	clean, err := preparePath(path)
	if err != nil {
		return Status{}, nil, err
	}
	if m.srv != nil {
		return Status{}, nil, ErrAlreadyMounted
	}
	if m.src == nil {
		return Status{}, nil, ErrNoSource
	}

	srv, err := mountAt(clean, m.src, opts)
	if err != nil {
		return Status{}, nil, fmt.Errorf("failed to mount at %q: %w", clean, wrapMountError(err))
	}

	m.srv = srv
	m.info = Status{
		Path:      clean,
		FsName:    mountFsName,
		MountTime: time.Now().Unix(),
	}
	return m.info, srv, nil
}

// clearExternal сбрасывает состояние, когда файловая система снята извне:
// иначе страница плагина показывала бы монтирование, которого уже нет.
func (m *Manager) clearExternal(srv mountServer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.srv != srv {
		return
	}
	log.Warnf("[Fuse] Mount at %q ended unexpectedly, the mount point may need a manual cleanup", m.info.Path)
	m.srv = nil
	m.info = Status{}
}

// Unmount снимает монтирование. При ошибке состояние сохраняется:
// занятую точку можно снять повторно или вручную через fusermount.
func (m *Manager) Unmount() (Status, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.unmountLocked()
}

// Status возвращает состояние монтирования; false - файловая система не смонтирована.
func (m *Manager) Status() (Status, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.srv == nil {
		return Status{}, false
	}
	return m.info, true
}

// Mounted сообщает, смонтирован ли источник.
func (m *Manager) Mounted() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.srv != nil
}

// Close размонтирует файловую систему, если она смонтирована.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.srv == nil {
		return nil
	}
	_, err := m.unmountLocked()
	return err
}

func (m *Manager) unmountLocked() (Status, error) {
	if m.srv == nil {
		return Status{}, ErrNotMounted
	}

	info := m.info
	if err := m.srv.Unmount(); err != nil {
		log.Warnf("[Fuse] Unmount of %q failed: %v", info.Path, err)
		return Status{}, fmt.Errorf("failed to unmount %q: %w", info.Path, err)
	}

	m.srv = nil
	m.info = Status{}
	return info, nil
}

// wrapMountError уточняет отказ монтирования. Текст от fusermount
// содержит причину, но go-fuse отдаёт только код возврата, поэтому
// распознаём по тому, что успело попасть в ошибку.
func wrapMountError(err error) error {
	msg := err.Error()

	switch {
	case strings.Contains(msg, "user_allow_other"):
		return &Error{Kind: ErrAllowOther, Detail: msg}
	case strings.Contains(msg, "/dev/fuse"), strings.Contains(msg, "fuse device"):
		return &Error{Kind: ErrNoDevice, Detail: msg}
	default:
		return &Error{Kind: ErrMountFailed, Detail: msg}
	}
}

// preparePath проверяет каталог монтирования. Путь должен быть
// абсолютным: относительный зависит от рабочего каталога сервера.
func preparePath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", &Error{Kind: ErrEmptyPath}
	}
	if !filepath.IsAbs(p) {
		return "", &Error{Kind: ErrNotAbsolute, Detail: p}
	}

	clean := filepath.Clean(p)
	info, err := os.Stat(clean)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist):
			return "", &Error{Kind: ErrNotExist, Detail: clean}
		case errors.Is(err, os.ErrPermission):
			return "", &Error{Kind: ErrPermission, Detail: clean}
		default:
			return "", &Error{Kind: ErrPermission, Detail: err.Error()}
		}
	}
	if !info.IsDir() {
		return "", &Error{Kind: ErrNotDir, Detail: clean}
	}
	return clean, nil
}
