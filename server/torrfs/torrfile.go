package torrfs

import (
	"io/fs"
	"server/torr/storage/torrstor"
	"time"

	"github.com/anacrolix/torrent"
)

type TorrFile struct {
	INode
	file   *torrent.File
	reader *torrstor.Reader
}

type TorrFileHandle struct {
	*TorrFile
	r *torrstor.Reader
}

func NewTorrFile(parent INode, name string, file *torrent.File) *TorrFile {
	f := &TorrFile{
		file: file,
		INode: &Node{
			parent: parent,
			torr:   parent.Torrent(),
			info: info{
				name:  name,
				size:  file.Length(),
				mode:  0444,
				mtime: time.Unix(parent.Torrent().Timestamp, 0),
				isDir: false,
			},
		},
	}
	return f
}

func (f *TorrFile) Open(name string) (fs.File, error) {
	r := f.Torrent().NewReader(f.file)
	if r == nil {
		return nil, fs.ErrInvalid
	}
	return &TorrFileHandle{TorrFile: f, r: r}, nil
}

func (h *TorrFileHandle) Read(p []byte) (int, error) {
	return h.r.Read(p)
}

func (h *TorrFileHandle) Seek(off int64, whence int) (int64, error) {
	return h.r.Seek(off, whence)
}

func (h *TorrFileHandle) Close() error {
	h.r.Close()
	return nil
}
