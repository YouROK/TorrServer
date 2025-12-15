package torrfs

import (
	"io/fs"
	"path"
	"strings"
)

type ioFSAdapter struct {
	root *RootDir
}

func AsFS(root *RootDir) fs.FS {
	return &ioFSAdapter{root: root}
}

func (a *ioFSAdapter) Open(name string) (fs.File, error) {
	name = path.Clean(name)
	if name == "." || name == "/" || name == "" {
		return a.root.Open(".")
	}
	name = strings.TrimPrefix(name, "/")
	return a.root.Open(name)
}
