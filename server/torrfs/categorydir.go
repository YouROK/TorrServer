package torrfs

import (
	"server/torr"
	"time"
)

type CategoryDir struct {
	INode
}

func NewCategoryDir(parent INode, category string) *CategoryDir {
	if category == "" {
		category = "other"
	}
	d := &CategoryDir{
		INode: &Node{
			parent: parent,
			info: info{
				name:  category,
				size:  4096,
				mode:  0777,
				mtime: time.Unix(477033666, 0),
				isDir: true,
			},
		},
	}
	d.BuildChildren()
	return d
}

func (d *CategoryDir) BuildChildren() {
	nodes := map[string]INode{}
	torrs := torr.ListTorrent()
	for _, t := range torrs {
		if t.Category == "" {
			t.Category = "other"
		}
		if t.Category == d.Name() {
			td := NewTorrDir(d, t.Title, t)
			td.isRoot = true
			nodes[t.Title] = td
		}
	}
	d.SetChildren(nodes)
}
