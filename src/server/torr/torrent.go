package torr

import (
	"io"
	"sort"
	"sync"
	"time"

	"server/log"
	"server/settings"
	"server/torr/state"
	"server/torr/utils"
	utils2 "server/utils"

	"server/torr/storage/torrstor"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

type Torrent struct {
	///// info for db
	Title  string
	Poster string
	*torrent.TorrentSpec

	Status state.TorrentStatus
	/////

	*torrent.Torrent
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
	torr.Status = state.TorrentAdded
	torr.lastTimeSpeed = time.Now()
	torr.bt = bt
	torr.closed = goTorrent.Closed()
	torr.TorrentSpec = spec
	torr.expiredTime = time.Now().Add(time.Minute)

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
		t.cache = t.bt.storage.GetCache(t.Hash())
		return true
	case <-t.closed:
		return false
	case <-tm.C:
		return false
	}
}

func (t *Torrent) GotInfo() bool {
	if t.Status == state.TorrentClosed {
		return false
	}
	t.Status = state.TorrentGettingInfo
	if t.WaitInfo() {
		t.Status = state.TorrentWorking
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
		log.TLogln("Torrent close by timeout", t.Torrent.InfoHash().HexString())
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
	return t.cache.ReadersLen() == 0 && t.expiredTime.Before(time.Now()) && (t.Status == state.TorrentWorking || t.Status == state.TorrentClosed)
}

func (t *Torrent) Files() []*torrent.File {
	if t.Torrent != nil && t.Torrent.Info() != nil {
		files := t.Torrent.Files()
		return files
	}
	return nil
}

func (t *Torrent) Hash() metainfo.Hash {
	if t.Torrent != nil {
		t.Torrent.InfoHash()
	}
	if t.TorrentSpec != nil {
		return t.TorrentSpec.InfoHash
	}
	return [20]byte{}
}

func (t *Torrent) Length() int64 {
	if t.Info() == nil {
		return 0
	}
	return t.Torrent.Length()
}

func (t *Torrent) NewReader(file *torrent.File, readahead int64) *Reader {
	if t.Status == state.TorrentClosed {
		return nil
	}
	reader := NewReader(t, file, readahead)
	return reader
}

func (t *Torrent) CloseReader(reader *Reader) {
	reader.Close()
	t.cache.RemReader(reader)
	t.expiredTime = time.Now().Add(time.Second * time.Duration(settings.BTsets.TorrentDisconnectTimeout))
}

func (t *Torrent) GetCache() *torrstor.Cache {
	return t.cache
}

func (t *Torrent) Preload(index int, size int64) {
	if size < 0 {
		return
	}

	if t.Status == state.TorrentGettingInfo {
		t.WaitInfo()
		// wait change status
		time.Sleep(100 * time.Millisecond)
	}

	t.muTorrent.Lock()
	if t.Status != state.TorrentWorking {
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
	t.Status = state.TorrentPreload
	t.muTorrent.Unlock()

	defer func() {
		if t.Status == state.TorrentPreload {
			t.Status = state.TorrentWorking
		}
	}()

	if index < 0 || index >= len(t.Files()) {
		index = 0
	}
	file := t.Files()[index]

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
	for t.Status == state.TorrentPreload {
		t.expiredTime = time.Now().Add(time.Minute * 5)
		t.PreloadedBytes = t.Torrent.BytesCompleted()
		log.TLogln("Preload:", file.Torrent().InfoHash().HexString(), utils2.Format(float64(t.PreloadedBytes)), "/", utils2.Format(float64(t.PreloadSize)), "Speed:", utils2.Format(t.DownloadSpeed), "Peers:[", t.Torrent.Stats().ConnectedSeeders, "]", t.Torrent.Stats().ActivePeers, "/", t.Torrent.Stats().TotalPeers)
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
	t.Status = state.TorrentClosed
	t.bt.mu.Lock()
	defer t.bt.mu.Unlock()

	if _, ok := t.bt.torrents[t.Hash()]; ok {
		delete(t.bt.torrents, t.Hash())
	}

	t.drop()
}

func (t *Torrent) Stats() *state.TorrentStats {
	t.muTorrent.Lock()
	defer t.muTorrent.Unlock()

	st := new(state.TorrentStats)

	st.TorrentStatus = t.Status
	st.TorrentStatusString = t.Status.String()
	st.Title = t.Title
	st.Poster = t.Poster

	if t.TorrentSpec != nil {
		st.Hash = t.TorrentSpec.InfoHash.HexString()
	}
	if t.Torrent != nil {
		st.Name = t.Torrent.Name()
		st.Hash = t.Torrent.InfoHash().HexString()
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
			st.FileStats = append(st.FileStats, state.TorrentFileStat{
				Id:     i + 1,
				Path:   f.Path(),
				Length: f.Length(),
			})
		}
	}
	return st
}
