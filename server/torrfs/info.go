package torrfs

import (
	"io/fs"
	"time"
)

type info struct {
	name  string
	size  int64
	mode  fs.FileMode
	mtime time.Time
	isDir bool
}

func (i info) Name() string       { return i.name }
func (i info) Size() int64        { return i.size }
func (i info) Mode() fs.FileMode  { return i.mode }
func (i info) ModTime() time.Time { return i.mtime }
func (i info) IsDir() bool        { return i.isDir }
func (i info) Sys() any           { return nil }
