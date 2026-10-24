package upload

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/anacrolix/torrent"

	"server/log"
	sets "server/settings"
	"server/tgbot/config"
	"server/torr"
	"server/torr/state"
	"server/torr/storage/torrstor"
)

var ErrStopped = errors.New("stopped")

type TorrFile struct {
	hash   string
	name   string
	wrk    *Worker
	offset int64
	size   int64
	mu     sync.Mutex

	reader *torrstor.Reader
}

func NewTorrFile(wrk *Worker, stFile *state.TorrentFileStat) (*TorrFile, error) {
	uid := int64(0)
	if wrk.c != nil && wrk.c.Sender() != nil {
		uid = wrk.c.Sender().ID
	}
	const limitCloud = int64(50 * 1024 * 1024)
	const limitLocal = int64(2 * 1024 * 1024 * 1024)
	official := config.Cfg == nil || config.OfficialBotAPI(config.Cfg.HostTG)
	if official && stFile.Length > limitCloud {
		return nil, errors.New(tr(uid, "upload_file_too_large_50mb"))
	}
	if !official && stFile.Length > limitLocal {
		return nil, errors.New(tr(uid, "upload_file_too_large_2gb"))
	}

	tf := new(TorrFile)
	tf.hash = wrk.torrentHash
	tf.name = filepath.Base(stFile.Path)
	tf.wrk = wrk
	tf.size = stFile.Length

	t := torr.GetTorrent(wrk.torrentHash)
	t.WaitInfo()

	files := t.Files()
	var file *torrent.File
	for _, tfile := range files {
		if tfile.Path() == stFile.Path {
			file = tfile
			break
		}
	}
	if file == nil {
		return nil, fmt.Errorf("file with id %v not found", stFile.Id)
	}
	if int64(sets.MaxSize) > 0 && file.Length() > int64(sets.MaxSize) {
		log.TLogln("tg upload err size", file.DisplayPath(), "max", sets.MaxSize)
		return nil, fmt.Errorf("file size exceeded max allowed %d bytes", sets.MaxSize)
	}

	reader := t.NewReader(file)
	if reader == nil {
		return nil, errors.New("cannot create torrent reader")
	}
	if sets.BTsets != nil && sets.BTsets.ResponsiveMode {
		reader.SetResponsive()
	}
	tf.reader = reader

	return tf, nil
}

func (t *TorrFile) Read(p []byte) (n int, err error) {
	if t.wrk.isCancelled {
		return 0, ErrStopped
	}
	n, err = t.reader.Read(p)
	t.mu.Lock()
	t.offset += int64(n)
	t.mu.Unlock()
	return
}

func (t *TorrFile) Remaining() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.size - t.offset
}

func (t *TorrFile) progress() (offset, size int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.offset, t.size
}

func (t *TorrFile) Close() {
	if t.reader != nil {
		t.reader.Close()
		t.reader = nil
	}
}
