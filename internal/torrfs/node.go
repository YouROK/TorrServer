package torrfs

import (
	"errors"
	"io"
	"time"
)

var (
	ErrNotFound = errors.New("path not found")
	ErrNotDir   = errors.New("not a directory")
	ErrIsDir    = errors.New("is a directory")
)

// Node — узел виртуального дерева торрентов.
type Node interface {
	Name() string
	Path() string
	IsDir() bool
	Size() int64
	ModTime() time.Time
	MimeType() string

	Children() ([]Node, error)
	Open() (io.ReadSeekCloser, error)
}
