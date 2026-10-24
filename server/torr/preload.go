package torr

import (
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"server/ffprobe"

	"server/log"
	"server/settings"
	"server/torr/state"
	utils2 "server/utils"

	"github.com/anacrolix/torrent"
)

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
	if t.Stat != state.TorrentWorking || t.Torrent == nil {
		t.muTorrent.Unlock()
		return
	}
	// drop() (Close, "drop all torrents" in SetSettings) sets t.Torrent to nil at any moment,
	// and "drop all torrents" leaves Stat as is. Use only this pointer below: it stays valid after Drop.
	torrent := t.Torrent
	t.Stat = state.TorrentPreload
	t.muTorrent.Unlock()

	defer func() {
		t.muTorrent.Lock()
		if t.Stat == state.TorrentPreload {
			t.Stat = state.TorrentWorking
		}
		t.muTorrent.Unlock()
		// Очистка по окончании прелоада
		t.BitRate = ""
		t.DurationSeconds = 0
	}()

	file := t.findFileIndex(torrent, index)
	if file == nil {
		file = torrent.Files()[0]
	}

	if size > file.Length() {
		size = file.Length()
	}

	if torrent.Info() == nil {
		return
	}

	timeout := time.Second * time.Duration(settings.BTsets.TorrentDisconnectTimeout)
	if timeout > time.Minute {
		timeout = time.Minute
	}

	// Create a stop channel for the logging goroutine
	logStopChan := make(chan struct{})
	defer close(logStopChan) // Ensure logging stops when function returns

	// Запуск лога в отдельном потоке
	go func(stopChan <-chan struct{}) {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				t.muTorrent.Lock()
				stat := t.Stat
				t.muTorrent.Unlock()

				if stat != state.TorrentPreload {
					return
				}
				st := torrent.Stats()
				statStr := fmt.Sprint(file.Torrent().InfoHash().HexString(), " ",
					utils2.Format(float64(t.PreloadedBytes)), " / ",
					utils2.Format(float64(t.PreloadSize)), " Speed: ",
					utils2.Format(t.DownloadSpeed), " Peers: ",
					st.ActivePeers, " / ",
					st.TotalPeers, " [ Seeds: ",
					st.ConnectedSeeders, " ]")
				log.TLogln("Preload:", statStr)
				t.AddExpiredTime(timeout)
			case <-stopChan:
				return
			}
		}
	}(logStopChan)

	if ffprobe.Exists() {
		link := settings.LoopbackBaseURL() + "/play/" + torrent.InfoHash().HexString() + "/" + strconv.Itoa(index)
		if data, err := ffprobe.ProbeUrl(link); err == nil {
			t.BitRate = data.Format.BitRate
			t.DurationSeconds = data.Format.DurationSeconds
		}
	}

	// Check if torrent was closed
	t.muTorrent.Lock()
	isClosed := t.Stat == state.TorrentClosed
	t.muTorrent.Unlock()

	if isClosed {
		log.TLogln("End preload: torrent closed")
		return
	}

	// startend -> 8/16 MB
	startend := torrent.Info().PieceLength
	if startend < 8<<20 {
		startend = 8 << 20
	}

	readerStart := file.NewReader()
	if readerStart == nil {
		log.TLogln("End preload: null reader")
		return
	}
	defer func() { _ = readerStart.Close() }()

	readerStartEnd, readerEndStart := preloadSpans(size, file.Length(), startend)
	readerStart.SetResponsive()
	readerStart.SetNext(0)
	readerStart.SetReadahead(0)
	readerStart.SetZone(readerStartEnd)
	readerEndEnd := file.Length()

	var wg sync.WaitGroup
	var preloadErr error

	// Tail covers bytes the head does not. Overlap used to skip the end
	// (or, when the budget equalled the window, swap which end was read).
	if readerEndStart < readerEndEnd {
		wg.Add(1)
		go func(from int64) {
			defer wg.Done()

			// Check if we should still preload
			t.muTorrent.Lock()
			shouldPreload := t.Stat == state.TorrentPreload
			t.muTorrent.Unlock()

			if !shouldPreload {
				return
			}

			readerEnd := file.NewReader()
			if readerEnd == nil {
				log.TLogln("Err preload: null reader")
				preloadErr = fmt.Errorf("null reader for end range")
				return
			}
			defer func() { _ = readerEnd.Close() }() // Ensure reader is always closed

			readerEnd.SetResponsive()
			readerEnd.SetNext(0)
			readerEnd.SetReadahead(0)

			_, err := readerEnd.Seek(from, io.SeekStart)
			if err != nil {
				log.TLogln("Err preload seek:", err)
				preloadErr = err
				return
			}
			// Zone is bytes ahead of the read position, so it has to be set after the seek.
			zone := readerEndEnd - from
			if zone < 0 {
				zone = 0
			}
			readerEnd.SetZone(zone)

			offset := from
			buf := make([]byte, 32<<10)
			for offset < readerEndEnd {
				tmp := buf
				if remain := readerEndEnd - offset; remain < int64(len(buf)) {
					tmp = buf[:remain]
				}
				n, err := readerEnd.Read(tmp)
				offset += int64(n)
				if n == 0 || err != nil {
					if err != nil && err != io.EOF {
						log.TLogln("Err preload read:", err)
						preloadErr = err
					}
					break
				}

				// Check if we should continue
				t.muTorrent.Lock()
				shouldContinue := t.Stat == state.TorrentPreload
				t.muTorrent.Unlock()

				if !shouldContinue {
					break
				}
			}
		}(readerEndStart)
	}

	// Main preload section
	pieceLength := torrent.Info().PieceLength
	readahead := pieceLength * 4
	if readerStartEnd < readahead {
		readahead = 0
	}
	readerStart.SetNext(readahead)
	readerStart.SetReadahead(readahead)
	readerStart.SetZone(readerStartEnd)

	offset := int64(0)
	buf := make([]byte, 32<<10)
	for offset < readerStartEnd {
		// Check if we should continue
		t.muTorrent.Lock()
		shouldContinue := t.Stat == state.TorrentPreload
		t.muTorrent.Unlock()

		if !shouldContinue {
			log.TLogln("Preload cancelled")
			break
		}

		tmp := buf
		if remain := readerStartEnd - offset; remain < int64(len(buf)) {
			tmp = buf[:remain]
		}
		n, err := readerStart.Read(tmp)
		offset += int64(n)

		if readahead > 0 && readerStartEnd-offset < readahead {
			readahead = 0
			remain := readerStartEnd - offset
			if remain < 0 {
				remain = 0
			}
			readerStart.SetNext(0)
			readerStart.SetReadahead(0)
			readerStart.SetZone(remain)
		}
		if n == 0 || err != nil {
			if err != nil && err != io.EOF {
				log.TLogln("Error preload:", err)
			}
			break
		}
	}

	// Wait for end range preload to complete
	wg.Wait()

	// Check if end range preload failed
	if preloadErr != nil {
		log.TLogln("End range preload failed:", preloadErr)
	}

	// Final log
	t.muTorrent.Lock()
	finalStat := t.Stat
	t.muTorrent.Unlock()

	if finalStat == state.TorrentPreload {
		st := torrent.Stats()
		log.TLogln("End preload:", file.Torrent().InfoHash().HexString(),
			"Peers:", st.ActivePeers, "/",
			st.TotalPeers, "[ Seeds:",
			st.ConnectedSeeders, "]")
	}
}

// preloadSpans splits a preload into [0, head) and [tail, fileLen).
// tail == fileLen means the head already reaches the end.
// When the windows overlap, tail starts where the head stopped, so both
// edges are read and the shared bytes are read once.
func preloadSpans(size, fileLen, edge int64) (head, tail int64) {
	if fileLen < 0 {
		fileLen = 0
	}
	if size > fileLen {
		size = fileLen
	}
	if size < 0 {
		size = 0
	}
	head = size - edge
	if head <= 0 {
		head = size
	}
	if head > fileLen {
		head = fileLen
	}
	tail = fileLen - edge
	if tail < head {
		tail = head
	}
	return head, tail
}

func (t *Torrent) findFileIndex(torrent *torrent.Torrent, index int) *torrent.File {
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
	for _, file := range torrent.Files() {
		if file.Path() == stFile.Path {
			return file
		}
	}
	return nil
}
