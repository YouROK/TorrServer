package torrfs

import (
	"io/fs"
	"path"
	"reflect"
	"server/settings"
	"server/torr"
	"strings"
	"syscall"
	"time"
)

type TorrDir struct {
	INode
	isRoot bool
}

func NewTorrDir(parent INode, name string, torrent *torr.Torrent) *TorrDir {
	d := &TorrDir{
		INode: &Node{
			parent: parent,
			torr:   torrent,
			info: info{
				name:  name,
				size:  4096,
				mode:  0777,
				mtime: time.Unix(torrent.Timestamp, 0),
				isDir: true,
			},
		},
	}
	d.BuildChildren()
	return d
}

func (d *TorrDir) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.Torrent() == nil {
		d.SetChildren(nil)
		return nil, syscall.ENOENT
	}
	return d.INode.ReadDir(n)
}

func (d *TorrDir) BuildChildren() {
	if d.Torrent() == nil {
		return
	}

	currTorrPath := ""
	if !d.isRoot {
		currTorrPath = strings.TrimPrefix(d.getTorrPath(), "/")
	}

	files := d.Torrent().Files()
	nodes := map[string]INode{}

	for _, file := range files {
		filePath := file.DisplayPath()

		if currTorrPath != "" {
			if !strings.HasPrefix(filePath, currTorrPath+"/") && filePath != currTorrPath {
				continue
			}
		}

		inCurrDir := strings.TrimPrefix(filePath, currTorrPath)
		inCurrDir = strings.TrimPrefix(inCurrDir, "/")
		if inCurrDir == "" {
			continue
		}

		arr := strings.SplitN(inCurrDir, "/", 2)
		name := arr[0]

		if len(arr) == 1 {
			nodes[name] = NewTorrFile(d, name, file)
		} else {
			if _, ok := nodes[name]; !ok {
				nodes[name] = NewTorrDir(d, name, d.Torrent())
			}
		}
	}

	d.SetChildren(nodes)
}

var (
	rootType = reflect.TypeOf((*RootDir)(nil))
	catType  = reflect.TypeOf((*CategoryDir)(nil))
)

func (d *TorrDir) getTorrPath() string {
	p := d.Name()
	n := d.Parent()

	for n != nil {
		nt := reflect.TypeOf(n)
		td, ok := n.(*TorrDir)

		stopByType := nt == rootType || nt == catType
		stopByRootTorrent := ok && td.isRoot

		if stopByType || stopByRootTorrent {
			break
		}

		p = n.Name() + "/" + p
		n = n.Parent()
	}

	p = "/" + p
	p = path.Clean(p)

	return p
}

func (d *TorrDir) Open(name string) (fs.File, error) {
	if !d.Torrent().GotInfo() {
		hash := d.Torrent().Hash().String()
		d.SetChildren(nil)
		for i := 0; i < settings.BTsets.TorrentDisconnectTimeout*2; i++ {
			tor := torr.GetTorrent(hash)
			if tor.GotInfo() {
				d.SetTorrent(tor)
				break
			}

			time.Sleep(time.Millisecond * 500)
		}
		if d.Torrent() == nil {
			d.SetChildren(nil)
			return nil, fs.ErrNotExist
		}
		d.BuildChildren()
	}
	return d.INode.Open(name)
}
