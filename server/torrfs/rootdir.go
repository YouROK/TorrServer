package torrfs

import (
	"io/fs"
	"path"
	"strings"
	"time"

	"server/torr"
)

type RootDir struct {
	info fs.FileInfo
}

func NewRootDir() *RootDir {
	return &RootDir{
		info: info{
			name:  "/",
			size:  4096,
			mode:  0o555,
			mtime: time.Unix(477033600, 0),
			isDir: true,
		},
	}
}

func (d *RootDir) Open(name string) (fs.File, error) {
	name = path.Clean(name)
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Path: name, Err: fs.ErrInvalid}
	}

	if name == "." || name == "/" {
		return d, nil
	}

	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}

	return Open(d, name)
}

func (d *RootDir) Stat() (fs.FileInfo, error) {
	return d.info, nil
}

func (d *RootDir) ReadDir(n int) ([]fs.DirEntry, error) {
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
		nodes[cat] = NewCategoryDir(cat)
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

// INode
func (d *RootDir) Parent() INode                 { return nil }
func (d *RootDir) Torrent() *torr.Torrent        { return nil }
func (d *RootDir) SetTorrent(torr *torr.Torrent) {}

// DirEntry
func (d *RootDir) Name() string { return d.info.Name() }
func (d *RootDir) IsDir() bool  { return true }
func (d *RootDir) Type() fs.FileMode {
	s, _ := d.Stat()
	return s.Mode()
}
func (d *RootDir) Info() (fs.FileInfo, error) { return d.info, nil }

// File
func (d *RootDir) Read(bytes []byte) (int, error) { return 0, fs.ErrInvalid }
func (d *RootDir) Close() error                   { return nil }
