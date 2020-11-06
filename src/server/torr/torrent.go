package torr

import (
	"io"
	"log"
	"sync"
	"time"

	"server/settings"
	"server/torr/utils"
	utils2 "server/utils"

	"server/torr/reader"
	"server/torr/storage/torrstor"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/labstack/gommon/bytes"
)

type TorrentStatus int

func (t TorrentStatus) String() string {
	switch t {
	case TorrentAdded:
		return "Torrent added"
	case TorrentGettingInfo:
		return "Torrent getting info"
	case TorrentPreload:
		return "Torrent preload"
	case TorrentWorking:
		return "Torrent working"
	case TorrentClosed:
		return "Torrent closed"
	default:
		return "Torrent unknown status"
	}
}

const (
	TorrentAdded = TorrentStatus(iota)
	TorrentGettingInfo
	TorrentPreload
	TorrentWorking
	TorrentClosed
)

type Torrent struct {
	*torrent.Torrent

	status TorrentStatus

	muTorrent sync.Mutex

	bt    *BTServer
	cache *torrstor.Cache

	lastTimeSpeed       time.Time
	DownloadSpeed       float64
	UploadSpeed         float64
	BytesReadUsefulData int64
	BytesWrittenData    int64

	PreloadSize    int64
	PreloadedBytes int64

	hash metainfo.Hash

	expiredTime time.Time

	closed <-chan struct{}

	progressTicker *time.Ticker
}

func NewTorrent(spec *torrent.TorrentSpec, bt *BTServer) (*Torrent, error) {

	switch settings.BTsets.RetrackersMode {
	case 1:
		spec.Trackers = append(spec.Trackers, [][]string{utils.GetDefTrackers()}...)
	case 2:
		spec.Trackers = nil
	case 3:
		spec.Trackers = [][]string{utils.GetDefTrackers()}
	}

	goTorrent, _, err := bt.client.AddTorrentSpec(spec)
	if err != nil {
		return nil, err
	}

	bt.mu.Lock()
	defer bt.mu.Unlock()
	if tor, ok := bt.torrents[spec.InfoHash]; ok {
		return tor, nil
	}

	torr := new(Torrent)
	torr.Torrent = goTorrent
	torr.status = TorrentAdded
	torr.lastTimeSpeed = time.Now()
	torr.bt = bt
	torr.hash = spec.InfoHash
	torr.closed = goTorrent.Closed()

	go torr.watch()

	bt.torrents[spec.InfoHash] = torr
	return torr, nil
}

func (t *Torrent) WaitInfo() bool {
	if t.Torrent == nil {
		return false
	}

	// Close torrent if not info while 10 minutes
	tm := time.NewTimer(time.Minute * 10)

	select {
	case <-t.Torrent.GotInfo():
		t.cache = t.bt.storage.GetCache(t.hash)
		return true
	case <-t.closed:
		return false
	case <-tm.C:
		return false
	}
}

func (t *Torrent) GotInfo() bool {
	if t.status == TorrentClosed {
		return false
	}
	t.status = TorrentGettingInfo
	if t.WaitInfo() {
		t.status = TorrentWorking
		t.expiredTime = time.Now().Add(time.Minute * 5)
		return true
	} else {
		t.Close()
		return false
	}
}

func (t *Torrent) watch() {
	t.progressTicker = time.NewTicker(time.Second)
	defer t.progressTicker.Stop()

	for {
		select {
		case <-t.progressTicker.C:
			go t.progressEvent()
		case <-t.closed:
			t.Close()
			return
		}
	}
}

func (t *Torrent) progressEvent() {
	if t.expired() {
		t.drop()
		return
	}

	t.muTorrent.Lock()
	if t.Torrent != nil && t.Torrent.Info() != nil {
		st := t.Torrent.Stats()
		deltaDlBytes := st.BytesReadUsefulData.Int64() - t.BytesReadUsefulData
		deltaUpBytes := st.BytesWrittenData.Int64() - t.BytesWrittenData
		deltaTime := time.Since(t.lastTimeSpeed).Seconds()

		t.DownloadSpeed = float64(deltaDlBytes) / deltaTime
		t.UploadSpeed = float64(deltaUpBytes) / deltaTime

		t.BytesWrittenData = st.BytesWrittenData.Int64()
		t.BytesReadUsefulData = st.BytesReadUsefulData.Int64()
	} else {
		t.DownloadSpeed = 0
		t.UploadSpeed = 0
	}
	t.muTorrent.Unlock()

	t.lastTimeSpeed = time.Now()
	t.updateRA()
}

func (t *Torrent) updateRA() {
	if t.BytesReadUsefulData > settings.BTsets.PreloadBufferSize {
		pieceLen := t.Torrent.Info().PieceLength
		adj := pieceLen * int64(t.Torrent.Stats().ActivePeers) / int64(1+t.cache.ReadersLen())
		switch {
		case adj < pieceLen:
			adj = pieceLen
		case adj > pieceLen*4:
			adj = pieceLen * 4
		}
		t.cache.AdjustRA(adj)
	}
}

func (t *Torrent) expired() bool {
	return t.cache.ReadersLen() == 0 && t.expiredTime.Before(time.Now()) && (t.status == TorrentWorking || t.status == TorrentClosed)
}

func (t *Torrent) Files() []*torrent.File {
	if t.Torrent != nil && t.Torrent.Info() != nil {
		files := t.Torrent.Files()
		return files
	}
	return nil
}

func (t *Torrent) Hash() metainfo.Hash {
	return t.hash
}

func (t *Torrent) Status() TorrentStatus {
	return t.status
}

func (t *Torrent) Length() int64 {
	if t.Info() == nil {
		return 0
	}
	return t.Torrent.Length()
}

func (t *Torrent) NewReader(file *torrent.File, readahead int64) *reader.Reader {
	if t.status == TorrentClosed {
		return nil
	}
	reader := reader.NewReader(t, file, readahead)
	return reader
}

func (t *Torrent) CloseReader(reader *reader.Reader) {
	reader.Close()
	t.cache.RemReader(reader)
	t.expiredTime = time.Now().Add(time.Second * time.Duration(settings.BTsets.TorrentDisconnectTimeout))
}

func (t *Torrent) GetCache() *torrstor.Cache {
	return t.cache
}

func (t *Torrent) Preload(file *torrent.File, size int64) {
	if size < 0 {
		return
	}

	if t.status == TorrentGettingInfo {
		t.WaitInfo()
		// wait change status
		time.Sleep(100 * time.Millisecond)
	}

	t.muTorrent.Lock()
	if t.status != TorrentWorking {
		t.muTorrent.Unlock()
		return
	}

	if size == 0 {
		size = settings.BTsets.PreloadBufferSize
	}
	if size == 0 {
		t.muTorrent.Unlock()
		return
	}
	t.status = TorrentPreload
	t.muTorrent.Unlock()

	defer func() {
		if t.status == TorrentPreload {
			t.status = TorrentWorking
		}
	}()

	buff5mb := int64(5 * 1024 * 1024)
	startPreloadLength := size
	endPreloadOffset := int64(0)
	if startPreloadLength > buff5mb {
		endPreloadOffset = file.Offset() + file.Length() - buff5mb
	}

	readerPre := t.NewReader(file, startPreloadLength)
	if readerPre == nil {
		return
	}
	defer func() {
		t.CloseReader(readerPre)
		t.expiredTime = time.Now().Add(time.Minute * 5)
	}()

	if endPreloadOffset > 0 {
		readerPost := t.NewReader(file, 1)
		if readerPre == nil {
			return
		}
		readerPost.Seek(endPreloadOffset, io.SeekStart)
		readerPost.SetReadahead(buff5mb)
		defer func() {
			t.CloseReader(readerPost)
			t.expiredTime = time.Now().Add(time.Minute * 5)
		}()
	}

	if size > file.Length() {
		size = file.Length()
	}

	t.PreloadSize = size
	var lastSize int64 = 0
	errCount := 0
	for t.status == TorrentPreload {
		t.expiredTime = time.Now().Add(time.Minute * 5)
		t.PreloadedBytes = t.Torrent.BytesCompleted()
		log.Println("Preload:", file.Torrent().InfoHash().HexString(), bytes.Format(t.PreloadedBytes), "/", bytes.Format(t.PreloadSize), "Speed:", utils2.Format(t.DownloadSpeed), "Peers:[", t.Torrent.Stats().ConnectedSeeders, "]", t.Torrent.Stats().ActivePeers, "/", t.Torrent.Stats().TotalPeers)
		if t.PreloadedBytes >= t.PreloadSize {
			return
		}

		if lastSize == t.PreloadedBytes {
			errCount++
		} else {
			lastSize = t.PreloadedBytes
			errCount = 0
		}
		if errCount > 120 {
			return
		}
		time.Sleep(time.Second)
	}
}

func (t *Torrent) drop() {
	t.muTorrent.Lock()
	if t.Torrent != nil {
		t.Torrent.Drop()
		t.Torrent = nil
	}
	t.muTorrent.Unlock()
}

func (t *Torrent) Close() {
	t.status = TorrentClosed
	t.bt.mu.Lock()
	defer t.bt.mu.Unlock()

	if _, ok := t.bt.torrents[t.hash]; ok {
		delete(t.bt.torrents, t.hash)
	}

	t.drop()
}
