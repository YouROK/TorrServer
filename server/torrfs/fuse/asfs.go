package fuse

import (
	"io/fs"
	"path"
	"server/torrfs"
	"strings"
)

type ioFSAdapter struct {
	root *torrfs.RootDir
}

func AsFS(root *torrfs.RootDir) fs.FS {
	return &ioFSAdapter{root: root}
}

func (a *ioFSAdapter) Open(name string) (fs.File, error) {
	name = path.Clean(name)
	if name == "/" || name == "" {
		return a.root.Open(".")
	}
	name = strings.TrimPrefix(name, "/")
	return a.root.Open(name)
}
