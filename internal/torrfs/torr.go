package torrfs

import (
	"io"
	"time"

	"silo/internal/torrent"
	"silo/internal/user"
)

type TorrNode struct {
	fs   *TorrFS
	user *user.User
	ut   *user.UserTorrent
	rec  *torrent.TorrentRecord
	name string
	path string
}

func (t *TorrNode) Name() string                     { return t.name }
func (t *TorrNode) Path() string                     { return t.path }
func (t *TorrNode) IsDir() bool                      { return true }
func (t *TorrNode) Size() int64                      { return 0 }
func (t *TorrNode) ModTime() time.Time               { return t.ut.AddedAt }
func (t *TorrNode) MimeType() string                 { return "" }
func (t *TorrNode) Open() (io.ReadSeekCloser, error) { return nil, ErrIsDir }

func (t *TorrNode) Children() ([]Node, error) {
	b := newTreeBuilder()
	for _, f := range t.rec.Files {
		b.insert(f.Path, f)
	}
	return buildChildrenFromTree(t.fs, t.user, t.ut.TorrentHash, t.ut.AddedAt, b, t.path), nil
}
