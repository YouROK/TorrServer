package torrfs

import (
	"io/fs"
	"time"

	sets "server/settings"
	"server/torr"
	"server/torr/storage/torrstor"

	"github.com/anacrolix/torrent"
)

type TorrFile struct {
	parent INode

	info fs.FileInfo

	torr   *torr.Torrent
	file   *torrent.File
	reader *torrstor.Reader
}

type TorrFileHandle struct {
	*TorrFile
	r *torrstor.Reader
}

func NewTorrFile(parent INode, name string, file *torrent.File) *TorrFile {
	f := &TorrFile{
		file:   file,
		parent: parent,
		torr:   parent.Torrent(),
		info: info{
			name:  name,
			size:  file.Length(),
			mode:  0o444,
			mtime: time.Unix(parent.Torrent().Timestamp, 0),
			isDir: false,
		},
	}
	return f
}

func (f *TorrFile) Open(name string) (fs.File, error) {
	r := f.Torrent().NewReader(f.file)
	if r == nil {
		return nil, fs.ErrInvalid
	}
	if sets.BTsets.ResponsiveMode {
		r.SetResponsive()
	}
	return &TorrFileHandle{TorrFile: f, r: r}, nil
}

// INode
func (f *TorrFile) Parent() INode                 { return f.parent }
func (f *TorrFile) Torrent() *torr.Torrent        { return f.torr }
func (f *TorrFile) SetTorrent(torr *torr.Torrent) { f.torr = torr }

// DirEntry
func (f *TorrFile) Name() string { return f.info.Name() }
func (f *TorrFile) IsDir() bool  { return false }
func (f *TorrFile) Type() fs.FileMode {
	s, _ := f.Stat()
	return s.Mode()
}
func (f *TorrFile) Info() (fs.FileInfo, error)           { return f.info, nil }
func (f *TorrFile) Stat() (fs.FileInfo, error)           { return f.info, nil }
func (f *TorrFile) Read(p []byte) (int, error)           { return 0, fs.ErrInvalid }
func (f *TorrFile) Close() error                         { return nil }
func (f *TorrFile) ReadDir(n int) ([]fs.DirEntry, error) { return nil, fs.ErrInvalid }

func (h *TorrFileHandle) Read(p []byte) (int, error) {
	return h.r.Read(p)
}

func (h *TorrFileHandle) Seek(off int64, whence int) (int64, error) {
	return h.r.Seek(off, whence)
}

func (h *TorrFileHandle) Close() error {
	h.torr.CloseReader(h.r)
	return nil
}
