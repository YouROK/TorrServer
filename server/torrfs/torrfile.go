package torrfs

import (
	"io/fs"
	"time"

	"github.com/anacrolix/torrent"
)

type TorrFile struct {
	INode
	file *torrent.File
}

// TODO не показывает файлы
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

func (f *TorrFile) Read(p []byte) (int, error) {
	return 0, fs.ErrInvalid
}

func (f *TorrFile) Close() error {
	return nil
}
