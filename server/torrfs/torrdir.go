package torrfs

import (
	"io/fs"
	"path"
	"strings"
	"time"

	"server/settings"
	"server/torr"
)

type TorrDir struct {
	parent   INode
	children map[string]INode

	info fs.FileInfo

	torr *torr.Torrent
}

func NewTorrDir(parent INode, name string, torrent *torr.Torrent) *TorrDir {
	d := &TorrDir{
		parent: parent,
		torr:   torrent,
		info: info{
			name:  name,
			size:  4096,
			mode:  0o555,
			mtime: time.Unix(torrent.Timestamp, 0),
			isDir: true,
		},
	}
	return d
}

func (d *TorrDir) ReadDir(n int) ([]fs.DirEntry, error) {
	if d.Torrent() == nil {
		return nil, fs.ErrInvalid
	}
	// соединяемся с торрентом при чтении директории торрента
	if !d.Torrent().GotInfo() {
		hash := d.Torrent().Hash().String()
		for i := 0; i < settings.BTsets.TorrentDisconnectTimeout*2; i++ {
			tor := torr.GetTorrent(hash)
			if tor.GotInfo() {
				d.SetTorrent(tor)
				break
			}

			time.Sleep(time.Millisecond * 500)
		}
		if d.Torrent() == nil {
			return nil, fs.ErrNotExist
		}
	}

	files := d.Torrent().Files()
	nodes := map[string]fs.DirEntry{}

	currTorrPath := d.getTorrPath()

	for _, file := range files {
		dp := file.DisplayPath()

		var rel string
		if currTorrPath == "" {
			rel = dp
		} else {
			prefix := currTorrPath + "/"
			if !strings.HasPrefix(dp, prefix) {
				continue
			}
			rel = strings.TrimPrefix(dp, prefix)
		}

		if rel == "" {
			continue
		}

		arr := strings.SplitN(rel, "/", 2)
		name := arr[0]
		if name == "" {
			continue
		}

		if len(arr) == 1 {
			nodes[name] = NewTorrFile(d, name, file)
		} else if _, ok := nodes[name]; !ok {
			nodes[name] = NewTorrDir(d, name, d.Torrent())
		}
	}

	var entries []fs.DirEntry
	for _, c := range nodes {
		entries = append(entries, c)
	}
	if n > 0 && len(entries) > n {
		entries = entries[:n]
	}
	return entries, nil
}

func (d *TorrDir) getTorrPath() string {
	parts := []string{}

	for n := INode(d); n != nil && n.Torrent() != nil; n = n.Parent() {
		if n.Parent() != nil && n.Parent().Torrent() == nil {
			continue
		}
		parts = append([]string{n.Name()}, parts...)
	}

	// отдаем без самого названия торрента
	if len(parts) > 0 {
		return path.Join(parts[1:]...)
	}
	return ""
}

func (d *TorrDir) Open(name string) (fs.File, error) {
	return Open(d, name)
}

// INode
func (d *TorrDir) Parent() INode                 { return d.parent }
func (d *TorrDir) Torrent() *torr.Torrent        { return d.torr }
func (d *TorrDir) SetTorrent(torr *torr.Torrent) { d.torr = torr }

// DirEntry
func (d *TorrDir) Name() string { return d.info.Name() }
func (d *TorrDir) IsDir() bool  { return true }
func (d *TorrDir) Type() fs.FileMode {
	s, _ := d.Stat()
	return s.Mode()
}
func (d *TorrDir) Info() (fs.FileInfo, error) { return d.info, nil }
func (d *TorrDir) Stat() (fs.FileInfo, error) { return d.info, nil }

// File
func (d *TorrDir) Read(bytes []byte) (int, error) { return 0, fs.ErrInvalid }
func (d *TorrDir) Close() error                   { return nil }
