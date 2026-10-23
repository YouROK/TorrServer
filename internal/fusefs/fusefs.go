// Package fusefs предоставляет библиотеку торрентов как локальную файловую
// систему. Монтирование доступно только на платформах с FUSE, на остальных
// собирается заглушка с тем же API.
package fusefs

import (
	"errors"
	"io"
	"time"
)

const (
	// mountFsName - имя файловой системы в выводе mount.
	mountFsName = "silo"
	// mountName - подтип монтирования (fuse.silo).
	mountName = "silo"
)

// Node - узел дерева, отдаваемого в файловую систему.
type Node interface {
	Name() string
	IsDir() bool
	Size() int64
	ModTime() time.Time
}

// File - открытый файл источника.
type File interface {
	io.ReadSeekCloser
	Name() string
	ModTime() time.Time
}

// FS - источник данных для монтирования: дерево уже привязано к владельцу.
// Отсутствующий путь должен возвращать ошибку, оборачивающую os.ErrNotExist.
type FS interface {
	Stat(path string) (Node, error)
	List(path string) ([]Node, error)
	Open(path string) (File, error)
}

// Options - параметры монтирования.
type Options struct {
	AllowOther bool `json:"allow_other"`
}

// Status - состояние смонтированной файловой системы.
type Status struct {
	Path      string `json:"path"`
	FsName    string `json:"fs_name"`
	MountTime int64  `json:"mount_time"`
}

var (
	ErrUnsupported    = errors.New("fuse is not supported on this platform")
	ErrNoSource       = errors.New("torrfs source is not linked")
	ErrEmptyPath      = errors.New("mount path is required")
	ErrNotAbsolute    = errors.New("mount path must be absolute")
	ErrNotExist       = errors.New("mount path does not exist")
	ErrPermission     = errors.New("mount path is not accessible")
	ErrNotDir         = errors.New("mount path is not a directory")
	ErrAlreadyMounted = errors.New("path is already mounted")
	ErrNotMounted     = errors.New("path is not mounted")
	ErrAllowOther     = errors.New("allow_other requires user_allow_other in fuse.conf")
	ErrNoDevice       = errors.New("fuse is not available on this system")
	ErrMountFailed    = errors.New("failed to mount")
)

// Error связывает категорию отказа с подробностью для показа пользователю.
// Категория переводится на стороне плагина, подробность показывается как есть.
type Error struct {
	Kind   error
	Detail string
}

func (e *Error) Error() string {
	if e.Detail == "" {
		return e.Kind.Error()
	}
	return e.Kind.Error() + ": " + e.Detail
}

// Unwrap делает категорию доступной для errors.Is.
func (e *Error) Unwrap() error { return e.Kind }
