package torrfs

import (
	"io/fs"
	"path"
	"server/torr"
	"strings"
	"time"
)

type RootDir struct {
	INode
}

func NewRootDir() *RootDir {
	r := &RootDir{
		INode: &Node{
			info: info{
				name:  "/",
				size:  4096,
				mode:  0555,
				mtime: time.Unix(477033600, 0),
				isDir: true,
			},
		},
	}
	r.buildChildren()
	return r
}

func (d *RootDir) Open(name string) (fs.File, error) {
	name = path.Clean(name)
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Path: name, Err: fs.ErrInvalid}
	}

	if name == "." || name == "/" {
		d.BuildChildren()
		return d, nil
	}

	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}

	return d.INode.Open(name)
}

func (d *RootDir) buildChildren() {
	torrs := torr.ListTorrent()
	cats := map[string]struct{}{}
	nodes := map[string]INode{}

	for _, torrent := range torrs {
		cats[torrent.Category] = struct{}{}
	}

	for cat := range cats {
		if cat == "" {
			cat = "other"
		}
		nodes[cat] = NewCategoryDir(d, cat)
	}

	d.SetChildren(nodes)
}
