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

// torrentNames names the directory of each torrent. Torrents whose titles give
// the same name get the start of their hash appended, or only the first of
// them could be opened.
func torrentNames(torrs []*torr.Torrent) []string {
	names := make([]string, len(torrs))
	count := map[string]int{}
	for i, t := range torrs {
		names[i] = torrentName(t)
		count[names[i]]++
	}
	for i, t := range torrs {
		if count[names[i]] > 1 {
			names[i] += " [" + t.Hash().HexString()[:8] + "]"
		}
	}
	return names
}

func (d *CategoryDir) ReadDir(n int) ([]fs.DirEntry, error) {
	nodes := []fs.DirEntry{}
	var torrs []*torr.Torrent
	for _, t := range torr.ListTorrent() {
		if categoryName(t.Category) != d.Name() {
			continue
		}
		if settings.BTsets.ShowFSActiveTorr && !t.GotInfo() {
			continue
		}
		torrs = append(torrs, t)
	}
	for i, name := range torrentNames(torrs) {
		nodes = append(nodes, NewTorrDir(nil, name, torrs[i]))
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
