package torr

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"server/log"
	"server/settings"
	"server/torr/state"
	cacheSt "server/torr/storage/state"
	"server/torr/storage/torrstor"
	"server/torr/utils"
	utils2 "server/utils"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

type Torrent struct {
	Title  string
	Poster string
	Data   string
	*torrent.TorrentSpec

	Stat	  state.TorrentStat
	Timestamp int64
	Size	  int64

	*torrent.Torrent
	muTorrent sync.Mutex

	bt	*BTServer
	cache *torrstor.Cache

	lastTimeSpeed	   time.Time
	DownloadSpeed	   float64
	UploadSpeed		 float64
	BytesReadUsefulData int64
	BytesWrittenData	int64

	PreloadSize	int64
	PreloadedBytes int64

	expiredTime time.Time

	closed <-chan struct{}

	progressTicker *time.Ticker
}

func NewTorrent(spec *torrent.TorrentSpec, bt *BTServer) (*Torrent, error) {
	// TODO panic when settings sets
	if bt == nil {
		return nil, errors.New("BT client not connected")
	}
		spec.Trackers = [][]string{[]string{"http://retracker.local"}}

		if settings.BTsets.RetrackersFromFile {
				spec.Trackers = append(spec.Trackers, [][]string{utils.LoadFromFile()}...)
		}

		if settings.BTsets.RetrackersFromNGOSang {
				spec.Trackers = append(spec.Trackers, [][]string{utils.LoadNGOSang()}...)
		}

		if settings.BTsets.RetrackersFromNewTrackon {
				spec.Trackers = append(spec.Trackers, [][]string{utils.LoadNewTrackon()}...)
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
	torr.Stat = state.TorrentAdded
	torr.lastTimeSpeed = time.Now()
	torr.bt = bt
	torr.closed = goTorrent.Closed()
	torr.TorrentSpec = spec
	torr.AddExpiredTime(time.Minute)
	torr.Timestamp = time.Now().Unix()

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
		t.cache.SetTorrent(t.Torrent)
		return true
	case <-t.closed:
		return false
	case <-tm.C:
		return false
	}
}

func (t *Torrent) GotInfo() bool {
	if t.Stat == state.TorrentClosed {
		return false
	}
	t.Stat = state.TorrentGettingInfo
	if t.WaitInfo() {
		t.Stat = state.TorrentWorking
		t.AddExpiredTime(time.Minute * 5)
		return true
	} else {
		t.Close()
		return false
	}
}

func (t *Torrent) AddExpiredTime(duration time.Duration) {
	t.expiredTime = time.Now().Add(duration)
}

func (t *Torrent) watch() {
	t.progressTicker = time.NewTicker(time.Second)
	defer t.progressTicker.Stop()

	for {
		select {
		case <-t.progressTicker.C:
			go t.progressEvent()
		case <-t.closed:
			return
		}
	}
}

func (t *Torrent) progressEvent() {
	if t.expired() {
		log.TLogln("Torrent close by timeout", t.Torrent.InfoHash().HexString())
		t.bt.RemoveTorrent(t.Hash())
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

		t.PreloadedBytes = t.Torrent.BytesCompleted()
	} else {
		t.DownloadSpeed = 0
		t.UploadSpeed = 0
	}
	t.muTorrent.Unlock()

	t.lastTimeSpeed = time.Now()
	t.updateRA()
}

func (t *Torrent) updateRA() {
	t.muTorrent.Lock()
	defer t.muTorrent.Unlock()
	if t.Torrent != nil && t.Torrent.Info() != nil {
		pieceLen := t.Torrent.Info().PieceLength
		adj := pieceLen * int64(t.Torrent.Stats().ActivePeers) / int64(1+t.cache.Readers())
		switch {
		case adj < pieceLen:
			adj = pieceLen
		case adj > pieceLen*4:
			adj = pieceLen * 4
		}
		go t.cache.AdjustRA(adj)
	}
}

func (t *Torrent) expired() bool {
	return t.cache.Readers() == 0 && t.expiredTime.Before(time.Now()) && (t.Stat == state.TorrentWorking || t.Stat == state.TorrentClosed)
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

func (t *Torrent) NewReader(file *torrent.File) *torrstor.Reader {
	if t.Stat == state.TorrentClosed {
		return nil
	}
	reader := t.cache.NewReader(file)
	return reader
}

func (t *Torrent) CloseReader(reader *torrstor.Reader) {
	t.cache.CloseReader(reader)
	t.AddExpiredTime(time.Second * time.Duration(settings.BTsets.TorrentDisconnectTimeout))
}

func (t *Torrent) GetCache() *torrstor.Cache {
	return t.cache
}

func (t *Torrent) Preload(index int, size int64) {
	if size <= 0 {
		return
	}
	t.PreloadSize = size

	if t.Stat == state.TorrentGettingInfo {
		if !t.WaitInfo() {
			return
		}
		// wait change status
		time.Sleep(100 * time.Millisecond)
	}

	t.muTorrent.Lock()
	if t.Stat != state.TorrentWorking {
		t.muTorrent.Unlock()
		return
	}

	t.Stat = state.TorrentPreload
	t.muTorrent.Unlock()

	defer func() {
		if t.Stat == state.TorrentPreload {
			t.Stat = state.TorrentWorking
		}
	}()

	file := t.findFileIndex(index)
	if file == nil {
		file = t.Files()[0]
	}

	if t.Info() != nil {
		pl := t.Info().PieceLength
		mb5 := int64(5 * 1024 * 1024)

		pieceFileStart := int(file.Offset() / pl)
		pieceFileEnd := int((file.Offset() + file.Length()) / pl)
		readerPieceBefore := int((file.Offset() + size - mb5) / pl)
		readerPieceAfter := int((file.Offset() + file.Length() - mb5) / pl)

		lastStat := time.Now().Add(-time.Second)

		for true {
			t.muTorrent.Lock()
			if t.Torrent == nil {
				return
			}
			t.PreloadedBytes = t.Torrent.BytesCompleted()
			t.muTorrent.Unlock()

			stat := fmt.Sprint(file.Torrent().InfoHash().HexString(), " ", utils2.Format(float64(t.PreloadedBytes)), "/", utils2.Format(float64(t.PreloadSize)), " Speed:", utils2.Format(t.DownloadSpeed), " Peers:[", t.Torrent.Stats().ConnectedSeeders, "]", t.Torrent.Stats().ActivePeers, "/", t.Torrent.Stats().TotalPeers)
			if time.Since(lastStat) > time.Second {
				log.TLogln("Preload:", stat)
				lastStat = time.Now()
			}

			isComplete := true
			if readerPieceBefore >= pieceFileStart {
				for i := pieceFileStart; i < readerPieceBefore; i++ {
					if !t.Piece(i).State().Complete {
						isComplete = false
						if t.Piece(i).State().Priority == torrent.PiecePriorityNone {
							t.Piece(i).SetPriority(torrent.PiecePriorityNormal)
						}
					}
				}
			}
			if readerPieceAfter <= pieceFileEnd {
				for i := readerPieceAfter; i <= pieceFileEnd; i++ {
					if !t.Piece(i).State().Complete {
						isComplete = false
						if t.Piece(i).State().Priority == torrent.PiecePriorityNone {
							t.Piece(i).SetPriority(torrent.PiecePriorityNormal)
						}
					}
				}
			}
			t.AddExpiredTime(time.Second * time.Duration(settings.BTsets.TorrentDisconnectTimeout))
			if isComplete {
				break
			}
			time.Sleep(time.Millisecond * 10)
		}
	}
	log.TLogln("End preload:", file.Torrent().InfoHash().HexString(), "Peers:[", t.Torrent.Stats().ConnectedSeeders, "]", t.Torrent.Stats().ActivePeers, "/", t.Torrent.Stats().TotalPeers)
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
	t.Stat = state.TorrentClosed

	t.bt.mu.Lock()
	if _, ok := t.bt.torrents[t.Hash()]; ok {
		delete(t.bt.torrents, t.Hash())
	}
	t.bt.mu.Unlock()

	t.drop()
}

func (t *Torrent) Status() *state.TorrentStatus {
	t.muTorrent.Lock()
	defer t.muTorrent.Unlock()

	st := new(state.TorrentStatus)

	st.Stat = t.Stat
	st.StatString = t.Stat.String()
	st.Title = t.Title
	st.Poster = t.Poster
	st.Data = t.Data
	st.Timestamp = t.Timestamp
	st.TorrentSize = t.Size

	if t.TorrentSpec != nil {
		st.Hash = t.TorrentSpec.InfoHash.HexString()
	}
	if t.Torrent != nil {
		st.Name = t.Torrent.Name()
		st.Hash = t.Torrent.InfoHash().HexString()
		st.LoadedSize = t.Torrent.BytesCompleted()

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

		if t.Torrent.Info() != nil {
			st.TorrentSize = t.Torrent.Length()

			files := t.Files()
			sort.Slice(files, func(i, j int) bool {
				return files[i].Path() < files[j].Path()
			})
			for i, f := range files {
				st.FileStats = append(st.FileStats, &state.TorrentFileStat{
					Id:	 i + 1, // in web id 0 is undefined
					Path:   f.Path(),
					Length: f.Length(),
				})
			}
		}
	}
	return st
}

func (t *Torrent) CacheState() *cacheSt.CacheState {
	if t.Torrent != nil && t.cache != nil {
		st := t.cache.GetState()
		st.Torrent = t.Status()
		return st
	}
	return nil
}

func (t *Torrent) findFileIndex(index int) *torrent.File {
	st := t.Status()
	var stFile *state.TorrentFileStat
	for _, f := range st.FileStats {
		if index == f.Id {
			stFile = f
			break
		}
	}
	if stFile == nil {
		return nil
	}
	for _, file := range t.Files() {
		if file.Path() == stFile.Path {
			return file
		}
	}
	return nil
}
