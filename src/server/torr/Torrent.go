package torr

import (
	"io"
	"log"
	"sort"
	"sync"
	"time"

	"server/settings"
	"server/utils"

	"server/torr/reader"
	"server/torr/storage/memcache"

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
	muReader  sync.Mutex

	bt    *BTServer
	cache *memcache.Cache

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

func NewTorrent(magnet metainfo.Magnet, infobytes []byte, bt *BTServer) (*Torrent, error) {
	switch settings.Get().RetrackersMode {
	case 1:
		magnet.Trackers = append(magnet.Trackers, utils.GetDefTrackers()...)
	case 2:
		magnet.Trackers = nil
	case 3:
		magnet.Trackers = utils.GetDefTrackers()
	}
	goTorrent, _, err := bt.client.AddTorrentSpec(&torrent.TorrentSpec{
		InfoBytes:   infobytes,
		Trackers:    [][]string{magnet.Trackers},
		DisplayName: magnet.DisplayName,
		InfoHash:    magnet.InfoHash,
	})

	if err != nil {
		return nil, err
	}

	bt.mu.Lock()
	defer bt.mu.Unlock()
	if tor, ok := bt.torrents[magnet.InfoHash]; ok {
		return tor, nil
	}

	torr := new(Torrent)
	torr.Torrent = goTorrent
	torr.status = TorrentAdded
	torr.lastTimeSpeed = time.Now()
	torr.bt = bt
	torr.hash = magnet.InfoHash
	torr.closed = goTorrent.Closed()

	go torr.watch()

	bt.torrents[magnet.InfoHash] = torr
	return torr, nil
}

func (t *Torrent) WaitInfo() bool {
	if t.Torrent == nil {
		return false
	}

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
	if t.BytesReadUsefulData > settings.Get().PreloadBufferSize {
		adj := int64((int(t.cache.GetState().PiecesLength) * t.Torrent.Stats().ActivePeers) / (1 + t.cache.ReadersLen()))
		switch {
		case adj < t.cache.GetState().PiecesLength:
			adj = t.cache.GetState().PiecesLength
		case adj > t.cache.GetState().PiecesLength*4:
			adj = t.cache.GetState().PiecesLength * 4
		}
		t.cache.AdjustRA(adj)
		//log.Println("Status:", t.Name(), "S:", fmt.Sprintf("%8s", utils.Format(t.DownloadSpeed)), "P:", fmt.Sprintf("%2d", t.Torrent.Stats().ActivePeers), "/", fmt.Sprintf("%2d", t.Torrent.Stats().TotalPeers), "R:", t.cache.ReadersLen(), "RA:", utils.Format(float64(adj)))
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
	t.muReader.Lock()

	if t.status == TorrentClosed {
		return nil
	}

	defer t.muReader.Unlock()
	reader := reader.NewReader(file)
	if readahead <= 0 {
		readahead = t.cache.GetState().PiecesLength
	}
	reader.SetReadahead(readahead)
	t.cache.AddReader(reader)
	return reader
}

func (t *Torrent) CloseReader(reader *reader.Reader) {
	t.muReader.Lock()
	reader.Close()
	t.cache.RemReader(reader)
	t.expiredTime = time.Now().Add(time.Second * time.Duration(settings.Get().TorrentDisconnectTimeout))
	t.muReader.Unlock()
}

func (t *Torrent) GetCache() *memcache.Cache {
	return t.cache
}

func (t *Torrent) Preload(file *torrent.File, size int64) {
	if size < 0 {
		return
	}

	if t.status == TorrentGettingInfo {
		t.WaitInfo()
	}

	t.muTorrent.Lock()
	if t.status != TorrentWorking {
		t.muTorrent.Unlock()
		return
	}

	if size == 0 {
		size = settings.Get().PreloadBufferSize
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
		if readerPost == nil {
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
		log.Println("Preload:", file.Torrent().InfoHash().HexString(), bytes.Format(t.PreloadedBytes), "/", bytes.Format(t.PreloadSize), "Speed:", utils.Format(t.DownloadSpeed), "Peers:[", t.Torrent.Stats().ConnectedSeeders, "]", t.Torrent.Stats().ActivePeers, "/", t.Torrent.Stats().TotalPeers)
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

	t.muReader.Lock()
	defer t.muReader.Unlock()

	if _, ok := t.bt.torrents[t.hash]; ok {
		delete(t.bt.torrents, t.hash)
	}

	t.drop()
}

func (t *Torrent) Stats() TorrentStats {
	t.muTorrent.Lock()
	defer t.muTorrent.Unlock()

	st := TorrentStats{}

	st.Name = t.Name()
	st.Hash = t.hash.HexString()
	st.TorrentStatus = t.status
	st.TorrentStatusString = t.status.String()

	if t.Torrent != nil {
		st.LoadedSize = t.Torrent.BytesCompleted()
		st.TorrentSize = t.Length()
		st.PreloadedBytes = t.PreloadedBytes
		st.PreloadSize = t.PreloadSize
		st.DownloadSpeed = t.DownloadSpeed
		st.UploadSpeed = t.UploadSpeed

		tst := t.Torrent.Stats()
		st.BytesWritten = tst.BytesWritten.Int64()
		st.BytesWrittenData = tst.BytesWrittenData.Int64()
		st.BytesRead = tst.BytesRead.Int64()
		st.BytesReadData = tst.BytesReadData.Int64()
		st.BytesReadUsefulData = tst.BytesReadUsefulData.Int64()
		st.ChunksWritten = tst.ChunksWritten.Int64()
		st.ChunksRead = tst.ChunksRead.Int64()
		st.ChunksReadUseful = tst.ChunksReadUseful.Int64()
		st.ChunksReadWasted = tst.ChunksReadWasted.Int64()
		st.PiecesDirtiedGood = tst.PiecesDirtiedGood.Int64()
		st.PiecesDirtiedBad = tst.PiecesDirtiedBad.Int64()
		st.TotalPeers = tst.TotalPeers
		st.PendingPeers = tst.PendingPeers
		st.ActivePeers = tst.ActivePeers
		st.ConnectedSeeders = tst.ConnectedSeeders
		st.HalfOpenPeers = tst.HalfOpenPeers

		files := t.Files()

		sort.Slice(files, func(i, j int) bool {
			return files[i].Path() < files[j].Path()
		})

		for i, f := range files {
			st.FileStats = append(st.FileStats, TorrentFileStat{
				Id:     i,
				Path:   f.Path(),
				Length: f.Length(),
			})
		}
	}
	return st
}
