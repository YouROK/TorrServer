package upload

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"

	"github.com/anacrolix/torrent"

	sets "server/settings"
	"server/tgbot/config"
	"server/torr"
	"server/torr/state"
	"server/torr/storage/torrstor"
)

var ERR_STOPPED = errors.New("stopped")

type TorrFile struct {
	hash   string
	name   string
	wrk    *Worker
	offset int64
	size   int64
	id     int

	reader *torrstor.Reader
}

func NewTorrFile(wrk *Worker, stFile *state.TorrentFileStat) (*TorrFile, error) {
	if config.Cfg.HostTG != "" && stFile.Length > 2*1024*1024*1024 {
		return nil, errors.New("Размер файла должен быть больше 2GB")
	}
	if config.Cfg.HostTG == "" && stFile.Length > 50*1024*1024 {
		return nil, errors.New("Размер файла должен быть больше 50MB\nЧтобы закачивать файлы до 2GB нужно в tg.cfg указать host к <a href='https://github.com/tdlib/telegram-bot-api'>telegram bot api</a>")
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
		log.Println("file", file.DisplayPath(), "size exceeded max allowed", sets.MaxSize, "bytes")
		return nil, fmt.Errorf("file size exceeded max allowed %d bytes", sets.MaxSize)
	}

	reader := t.NewReader(file)
	if reader == nil {
		return nil, errors.New("cannot create torrent reader")
	}
	if sets.BTsets.ResponsiveMode {
		reader.SetResponsive()
	}
	tf.reader = reader

	return tf, nil
}

func (t *TorrFile) Read(p []byte) (n int, err error) {
	if t.wrk.isCancelled {
		return 0, ERR_STOPPED
	}
	n, err = t.reader.Read(p)
	t.offset += int64(n)
	return
}

func (t *TorrFile) Loaded() int64 {
	return t.size - t.offset
}

func (t *TorrFile) Close() {
	if t.reader != nil {
		t.reader.Close()
		t.reader = nil
	}
}
