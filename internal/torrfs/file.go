package torrfs

import (
	"io"
	"time"

	"silo/internal/user"
)

type FileNode struct {
	fs      *TorrFS
	user    *user.User
	hash    string
	fileIdx int
	name    string
	path    string
	size    int64
	mtime   time.Time
	mime    string
}

func (f *FileNode) Name() string       { return f.name }
func (f *FileNode) Path() string       { return f.path }
func (f *FileNode) IsDir() bool        { return false }
func (f *FileNode) Size() int64        { return f.size }
func (f *FileNode) ModTime() time.Time { return f.mtime }
func (f *FileNode) MimeType() string   { return f.mime }

func (f *FileNode) Children() ([]Node, error) {
	return nil, ErrNotDir
}

// Open - загружает файл в движок торрента
func (f *FileNode) Open() (io.ReadSeekCloser, error) {
	r, _, err := f.fs.mgr.GetStreamReader(f.user, f.hash, f.fileIdx)
	if err != nil {
		return nil, err
	}
	return r, nil
}
