package torr

import (
	"fmt"
	"io"
	"runtime/debug"
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

// Safe reader implementation
type safeReader struct {
	io.ReadSeeker
	original interface{} // Store the original reader to access its methods
	checkFn  func() bool // returns true if torrent is closed
	streamID string      // for logging
}

// Create a constructor that properly wraps the reader
func newSafeReader(original interface{}, checkFn func() bool, streamID string) *safeReader {
	// Extract ReadSeeker from the original
	var readSeeker io.ReadSeeker
	if rs, ok := original.(io.ReadSeeker); ok {
		readSeeker = rs
	}

	return &safeReader{
		ReadSeeker: readSeeker,
		original:   original,
		checkFn:    checkFn,
		streamID:   streamID,
	}
}

// Override Read with safety checks
func (sr *safeReader) Read(p []byte) (n int, err error) {
	// Check if torrent is closed before attempting to read
	if sr.checkFn() {
		return 0, io.EOF
	}
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			log.TLogln("Recovered from panic in safeReader.Read:", r)
			err = io.EOF
			n = 0
		}
	}()
	// Try to read with a timeout to prevent hanging
	done := make(chan struct{})
	go func() {
		n, err = sr.ReadSeeker.Read(p)
		close(done)
	}()
	select {
	case <-done:
		return n, err
	case <-time.After(15 * time.Second): // 15 second read timeout
		// Check if torrent is still alive
		if sr.checkFn() {
			return 0, io.EOF
		}
		return 0, fmt.Errorf("15s read timeout")
	}
}

// Override Seek with safety checks
func (sr *safeReader) Seek(offset int64, whence int) (int64, error) {
	// Check if torrent is closed before attempting to seek
	if sr.checkFn() {
		return 0, fmt.Errorf("Preload seek error: torrent closed")
	}
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			log.TLogln("Recovered from panic in safeReader.Seek:", r)
		}
	}()
	return sr.ReadSeeker.Seek(offset, whence)
}

// Forward SetResponsive method if it exists
func (sr *safeReader) SetResponsive() {
	if rs, ok := sr.original.(interface{ SetResponsive() }); ok {
		rs.SetResponsive()
	}
}

// Forward SetReadahead method if it exists
func (sr *safeReader) SetReadahead(bytes int64) {
	if rs, ok := sr.original.(interface{ SetReadahead(int64) }); ok {
		rs.SetReadahead(bytes)
	}
}

// Forward Close method if it exists
func (sr *safeReader) Close() error {
	if closer, ok := sr.original.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (t *Torrent) Preload(index int, size int64) {
	// recovery from panic
	defer func() {
		if r := recover(); r != nil {
			log.TLogln("Recovered from panic in Preload:", r)
			// Stack trace for debugging
			debug.PrintStack()
		}
	}()

	if size <= 0 {
		return
	}
	t.PreloadSize = size

	// First, check if torrent is already closed
	t.muTorrent.Lock()
	if t.Stat == state.TorrentClosed {
		t.muTorrent.Unlock()
		log.TLogln("Preload skipped: torrent already closed")
		return
	}
	t.muTorrent.Unlock()

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
		t.muTorrent.Lock()
		if t.Stat == state.TorrentPreload {
			t.Stat = state.TorrentWorking
		}
		t.muTorrent.Unlock()
		// Очистка по окончании прелоада
		t.BitRate = ""
		t.DurationSeconds = 0
	}()

	file := t.findFileIndex(index)
	if file == nil {
		file = t.Files()[0]
	}

	if size > file.Length() {
		size = file.Length()
	}

	if t.Info() == nil {
		return
	}

	// Helper function to check if torrent is closed
	isTorrentClosed := func() bool {
		t.muTorrent.Lock()
		defer t.muTorrent.Unlock()
		return t.Stat == state.TorrentClosed || t.Torrent == nil
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

				if isTorrentClosed() {
					return
				}

				statStr := fmt.Sprint(file.Torrent().InfoHash().HexString(), " ",
					utils2.Format(float64(t.PreloadedBytes)), "/",
					utils2.Format(float64(t.PreloadSize)), " Speed:",
					utils2.Format(t.DownloadSpeed), " Peers:",
					t.Torrent.Stats().ActivePeers, "/",
					t.Torrent.Stats().TotalPeers, " [Seeds:",
					t.Torrent.Stats().ConnectedSeeders, "]")
				log.TLogln("Preload:", statStr)
				t.AddExpiredTime(timeout)
			case <-stopChan:
				return
			}
		}
	}(logStopChan)

	if ffprobe.Exists() {
		link := "http://127.0.0.1:" + settings.Port + "/play/" + t.Hash().HexString() + "/" + strconv.Itoa(index)
		if settings.Ssl {
			link = "https://127.0.0.1:" + settings.SslPort + "/play/" + t.Hash().HexString() + "/" + strconv.Itoa(index)
		}
		if data, err := ffprobe.ProbeUrl(link); err == nil {
			t.BitRate = data.Format.BitRate
			t.DurationSeconds = data.Format.DurationSeconds
		}
	}

	// Check if torrent was closed
	if isTorrentClosed() {
		log.TLogln("End preload: torrent closed")
		return
	}

	// startend -> 8/16 MB
	startend := t.Info().PieceLength
	if startend < 8<<20 {
		startend = 8 << 20
	}

	// Create the reader and wrap it immediately
	rawReader := file.NewReader()
	if rawReader == nil {
		log.TLogln("End preload: null reader")
		return
	}

	// Wrap the reader with our safe reader
	readerStart := newSafeReader(rawReader, isTorrentClosed, "preload-start")
	defer readerStart.Close()

	readerStart.SetResponsive()
	readerStart.SetReadahead(0)
	readerStartEnd := size - startend

	if readerStartEnd < 0 {
		// Если конец начального ридера оказался за началом
		readerStartEnd = size
	}
	if readerStartEnd > file.Length() {
		// Если конец начального ридера оказался после конца файла
		readerStartEnd = file.Length()
	}

	readerEndStart := file.Length() - startend
	readerEndEnd := file.Length()

	var wg sync.WaitGroup
	var preloadEndErr error

	// Start end range preload if needed
	if readerEndStart > readerStartEnd {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Check if we should still preload
			t.muTorrent.Lock()
			shouldPreload := t.Stat == state.TorrentPreload
			t.muTorrent.Unlock()

			if !shouldPreload || isTorrentClosed() {
				return
			}

			rawReaderEnd := file.NewReader()
			if rawReaderEnd == nil {
				preloadEndErr = fmt.Errorf("null reader for end range")
				return
			}

			// Wrap the end reader too
			readerEnd := newSafeReader(rawReaderEnd, isTorrentClosed, "preload-end")
			defer readerEnd.Close()

			readerEnd.SetResponsive()
			readerEnd.SetReadahead(0)

			_, err := readerEnd.Seek(readerEndStart, io.SeekStart)
			if err != nil {
				preloadEndErr = err
				return
			}

			offset := readerEndStart
			tmp := make([]byte, 32768)
			for offset+int64(len(tmp)) < readerEndEnd {
				// The safeReader will check isTorrentClosed internally
				n, err := readerEnd.Read(tmp)
				if err != nil {
					if err != io.EOF {
						preloadEndErr = err
					}
					break
				}
				offset += int64(n)

				// Check if we should continue
				t.muTorrent.Lock()
				shouldContinue := t.Stat == state.TorrentPreload
				t.muTorrent.Unlock()

				if !shouldContinue {
					break
				}
			}
		}()
	}

	// Main preload section
	pieceLength := t.Info().PieceLength
	readahead := pieceLength * 4
	if readerStartEnd < readahead {
		readahead = 0
	}
	readerStart.SetReadahead(readahead)

	offset := int64(0)
	tmp := make([]byte, 32768)
	for offset+int64(len(tmp)) < readerStartEnd {
		// Check if we should continue
		t.muTorrent.Lock()
		shouldContinue := t.Stat == state.TorrentPreload
		t.muTorrent.Unlock()

		if !shouldContinue {
			log.TLogln("Preload cancelled")
			break
		}

		// The safeReader will handle torrent closed checks internally
		n, err := readerStart.Read(tmp)
		if err != nil {
			if err != io.EOF {
				log.TLogln("Error preload:", err)
			}
			break
		}
		offset += int64(n)

		if readahead > 0 && readerStartEnd-(offset+int64(len(tmp))) < readahead {
			readahead = 0
			readerStart.SetReadahead(0)
		}
	}

	// Wait for end range preload to complete
	wg.Wait()

	// Check if end range preload failed
	if preloadEndErr != nil {
		log.TLogln("End range preload failed:", preloadEndErr)
	}

	// Final log - check if torrent still exists
	if !isTorrentClosed() {
		t.muTorrent.Lock()
		finalStat := t.Stat
		t.muTorrent.Unlock()

		if finalStat == state.TorrentPreload {
			log.TLogln("End preload:", file.Torrent().InfoHash().HexString(),
				"Peers:", t.Torrent.Stats().ActivePeers, "/",
				t.Torrent.Stats().TotalPeers, "[ Seeds:",
				t.Torrent.Stats().ConnectedSeeders, "]")
		}
	}
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
