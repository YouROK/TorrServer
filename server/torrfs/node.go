package torrfs

import (
	"io/fs"
	"path"
	"server/torr"
	"strings"
)

type INode interface {
	fs.ReadDirFile
	fs.DirEntry

	Open(name string) (fs.File, error)

	Parent() INode
	Children() map[string]INode
	SetChildren(map[string]INode)
	//FindChild(name string) INode

	Path() string
	Torrent() *torr.Torrent
	SetTorrent(torr *torr.Torrent)
	BuildChildren()
}

type Node struct {
	INode
	parent   INode
	children map[string]INode

	path string
	torr *torr.Torrent

	info fs.FileInfo
}

func (d *Node) Open(name string) (fs.File, error) {
	trimPath := strings.TrimPrefix(name, d.Name())
	trimPath = strings.TrimSuffix(trimPath, "/")
	trimPath = strings.TrimPrefix(trimPath, "/")
	if trimPath == "" {
		return d, nil
	}
	arr := strings.Split(trimPath, "/")
	if len(arr) == 0 {
		return nil, fs.ErrNotExist
	}
	chds := d.Children()
	if c, ok := chds[arr[0]]; ok {
		return c.Open(trimPath)
	}
	return nil, fs.ErrNotExist
}

func (d *Node) Stat() (fs.FileInfo, error) { return d.info, nil }
func (d *Node) Read(p []byte) (int, error) { return 0, fs.ErrInvalid }
func (d *Node) Close() error               { return nil }
func (d *Node) Info() (fs.FileInfo, error) { return d.Stat() }

func (d *Node) Parent() INode              { return d.parent }
func (d *Node) Children() map[string]INode { return d.children }

//func (d *Node) FindChild(name string) INode {
//	arr := strings.Split(name, "/")
//	var node INode = d
//	for _, n := range arr {
//		if n == "" {
//			continue
//		}
//		if c, ok := d.children[n]; ok {
//			node = c
//		} else {
//			return nil
//		}
//	}
//	return node
//}

func (d *Node) SetChildren(children map[string]INode) { d.children = children }
func (d *Node) Torrent() *torr.Torrent                { return d.torr }
func (d *Node) SetTorrent(torr *torr.Torrent)         { d.torr = torr }

func (d *Node) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.info.IsDir() {
		list := d.Children()
		var entries []fs.DirEntry
		for _, c := range list {
			entries = append(entries, c)
		}
		if n > 0 && len(entries) > n {
			entries = entries[:n]
		}
		return entries, nil
	}

	return nil, fs.ErrInvalid
}

func (d *Node) Name() string {
	s, _ := d.Stat()
	return s.Name()
}

func (d *Node) IsDir() bool {
	s, _ := d.Stat()
	return s.IsDir()
}

func (d *Node) Type() fs.FileMode {
	s, _ := d.Stat()
	return s.Mode()
}

func (d *Node) Path() string {
	if d.path == "" {
		p := d.Name()
		n := d.Parent()
		for n != nil {
			p = n.Name() + "/" + p
			n = n.Parent()
		}
		d.path = path.Clean(p)
	}
	return d.path
}

func (d *Node) BuildChildren() {}
