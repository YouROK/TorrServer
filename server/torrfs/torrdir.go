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
}

func NewTorrDir(parent INode, name string, torrent *torr.Torrent) *TorrDir {
	d := &TorrDir{
		INode: &Node{
			parent: parent,
			torr:   torrent,
			info: info{
				name:  name,
				size:  4096,
				mode:  0555,
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

	torrPath := d.getTorrPath()
	torrPath = strings.TrimPrefix(torrPath, "/"+d.Name())
	files := d.Torrent().Files()

	nodes := map[string]INode{}
	for _, file := range files {
		if strings.HasPrefix(file.Path(), torrPath) {
			right := strings.TrimLeft(file.Path(), torrPath)
			arr := strings.Split(right, "/")
			if len(arr) == 1 {
				nodes[arr[0]] = NewTorrFile(d, arr[0], file)
			} else {
				nodes[arr[0]] = NewTorrDir(d, arr[0], d.Torrent())
			}
		}
	}

	d.SetChildren(nodes)
}

func (d *TorrDir) getTorrPath() string {
	rootType := reflect.TypeOf((*RootDir)(nil))
	catType := reflect.TypeOf((*CategoryDir)(nil))

	p := d.Name()
	n := d.Parent()
	for n != nil && reflect.TypeOf(n) != rootType && reflect.TypeOf(n) != catType {
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
