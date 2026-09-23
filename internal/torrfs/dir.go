package torrfs

import (
	"io"
	"time"

	"silo/internal/user"
)

type DirNode struct {
	fs    *TorrFS
	user  *user.User
	hash  string
	name  string
	path  string
	mtime time.Time
	sub   *treeBuilder
}

func (d *DirNode) Name() string                     { return d.name }
func (d *DirNode) Path() string                     { return d.path }
func (d *DirNode) IsDir() bool                      { return true }
func (d *DirNode) Size() int64                      { return 0 }
func (d *DirNode) ModTime() time.Time               { return d.mtime }
func (d *DirNode) MimeType() string                 { return "" }
func (d *DirNode) Open() (io.ReadSeekCloser, error) { return nil, ErrIsDir }

func (d *DirNode) Children() ([]Node, error) {
	return buildChildrenFromTree(d.fs, d.user, d.hash, d.mtime, d.sub, d.path), nil
}
