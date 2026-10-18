package torrfs

import (
	"io/fs"
	"time"

	"server/settings"

	"server/torr"
)

type CategoryDir struct {
	info fs.FileInfo
}

func NewCategoryDir(category string) *CategoryDir {
	category = categoryName(category)
	d := &CategoryDir{
		info: info{
			name:  category,
			size:  4096,
			mode:  0o555,
			mtime: time.Unix(477033666, 0),
			isDir: true,
		},
	}
	return d
}

func categoryName(category string) string {
	category = SanitizeName(category)
	if category == "" {
		return "other"
	}
	return category
}

func torrentName(t *torr.Torrent) string {
	name := SanitizeName(t.Title)
	if name == "" {
		name = t.Hash().String()
	}
	return name
}

func (d *CategoryDir) Stat() (fs.FileInfo, error) {
	return d.info, nil
}

func (d *CategoryDir) ReadDir(n int) ([]fs.DirEntry, error) {
	nodes := []fs.DirEntry{}
	torrs := torr.ListTorrent()
	for _, t := range torrs {
		if categoryName(t.Category) != d.Name() {
			continue
		}
		if settings.BTsets.ShowFSActiveTorr && !t.GotInfo() {
			continue
		}
		td := NewTorrDir(nil, torrentName(t), t)
		nodes = append(nodes, td)
	}

	return nodes, nil
}

// INode
func (d *CategoryDir) Open(name string) (fs.File, error) { return Open(d, name) }
func (d *CategoryDir) Parent() INode                     { return nil }
func (d *CategoryDir) Torrent() *torr.Torrent            { return nil }
func (d *CategoryDir) SetTorrent(torr *torr.Torrent)     {}

// DirEntry
func (d *CategoryDir) Name() string { return d.info.Name() }
func (d *CategoryDir) IsDir() bool  { return true }
func (d *CategoryDir) Type() fs.FileMode {
	s, _ := d.Stat()
	return s.Mode()
}
func (d *CategoryDir) Info() (fs.FileInfo, error) { return d.info, nil }

// File
func (d *CategoryDir) Read(bytes []byte) (int, error) { return 0, fs.ErrInvalid }
func (d *CategoryDir) Close() error                   { return nil }
