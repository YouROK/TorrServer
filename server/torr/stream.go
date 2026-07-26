package torr

import (
	// "context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/anacrolix/dms/dlna"
	"github.com/anacrolix/missinggo/v2/httptoo"
	"github.com/anacrolix/torrent"

	"server/ffprobe"
	mt "server/mimetype"
	sets "server/settings"
	"server/torr/state"
	"server/torr/storage/torrstor"
)

// Add atomic counter for concurrent streams
var activeStreams int32

// type contextResponseWriter struct {
// 	http.ResponseWriter
// 	ctx context.Context
// }

// func (w *contextResponseWriter) Write(p []byte) (n int, err error) {
// 	// Check context before each write
// 	select {
// 	case <-w.ctx.Done():
// 		return 0, w.ctx.Err()
// 	default:
// 		return w.ResponseWriter.Write(p)
// 	}
// }

func (t *Torrent) Stream(fileID int, req *http.Request, resp http.ResponseWriter) error {
	// Increment active streams counter
	streamStart := time.Now()
	streamID := atomic.AddInt32(&activeStreams, 1)
	defer atomic.AddInt32(&activeStreams, -1)
	// Stream disconnect timeout (same as torrent)
	streamTimeout := sets.BTsets.TorrentDisconnectTimeout

	if !t.GotInfo() {
		http.NotFound(resp, req)
		return errors.New("torrent doesn't have info yet")
	}
	// Get file information
	st := t.Status()
	var stFile *state.TorrentFileStat
	for _, fileStat := range st.FileStats {
		if fileStat.Id == fileID {
			stFile = fileStat
			break
		}
	}
	if stFile == nil {
		return fmt.Errorf("file with id %v not found", fileID)
	}
	// Find the actual torrent file
	files := t.Files()
	var file *torrent.File
	for _, tfile := range files {
		if tfile.Path() == stFile.Path {
			file = tfile
			break
		}
	}
	if file == nil {
		return fmt.Errorf("file with id %v not found", fileID)
	}
	// Check file size limit
	if int64(sets.MaxSize) > 0 && file.Length() > int64(sets.MaxSize) {
		err := fmt.Errorf("file size exceeded max allowed %d bytes", sets.MaxSize)
		log.Printf("File %s size (%d) exceeded max allowed %d bytes", file.DisplayPath(), file.Length(), sets.MaxSize)
		http.Error(resp, err.Error(), http.StatusForbidden)
		return err
	}
	// Create reader with context for timeout
	reader := t.NewReader(file)
	if reader == nil {
		return errors.New("cannot create torrent reader")
	}
	// Ensure reader is always closed
	defer t.CloseReader(reader)

	if sets.BTsets.ResponsiveMode {
		reader.SetResponsive()
	}
	// Log connection
	host, port, clerr := net.SplitHostPort(req.RemoteAddr)

	if sets.BTsets.EnableDebug {
		if clerr != nil {
			log.Printf("[Stream:%d] Connect client (Active streams: %d)", streamID, atomic.LoadInt32(&activeStreams))
		} else {
			log.Printf("[Stream:%d] Connect client %s:%s (Active streams: %d)",
				streamID, host, port, atomic.LoadInt32(&activeStreams))
		}
	}

	// Our own ffprobe reads through this same endpoint; such a stream must not trigger
	// any position work, otherwise probing would recurse into itself.
	isProbe := req.URL.Query().Get(ProbeMarker) != ""

	// Mark as viewed (never clears an already saved playback position)
	if !isProbe {
		sets.MarkViewed(t.Hash().HexString(), fileID)
	}

	// Keep the position fresh while playing: a paused client that dies without closing
	// the connection would otherwise only be noticed when TCP finally times out.
	stopSaving := make(chan struct{})
	defer close(stopSaving) // registered after CloseReader, so the ticker stops first
	if !isProbe && positionSavingEnabled() {
		go func() {
			ticker := time.NewTicker(saveInterval)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					saveViewedPosition(t, fileID, file, reader, time.Since(streamStart))
				case <-stopSaving:
					return
				}
			}
		}()
	}

	// Set response headers
	resp.Header().Set("Connection", "close")
	// Set response header for Kodi
	resp.Header().Set("Server", "TorrServer (Portable SDK for UPnP devices)")
	// Add timeout header if configured
	if streamTimeout > 0 {
		resp.Header().Set("X-Stream-Timeout", fmt.Sprintf("%d", streamTimeout))
	}
	// Add ETag
	etag := hex.EncodeToString([]byte(fmt.Sprintf("%s/%s", t.Hash().HexString(), file.Path())))
	resp.Header().Set("ETag", httptoo.EncodeQuotedString(etag))
	// DLNA headers
	resp.Header().Set("transferMode.dlna.org", "Streaming")
	// add MimeType
	mime, err := mt.MimeTypeByPath(file.Path())
	if err == nil && mime.IsMedia() {
		resp.Header().Set("content-type", mime.String())
	}
	// DLNA Seek
	if req.Header.Get("getContentFeatures.dlna.org") != "" {
		resp.Header().Set("contentFeatures.dlna.org", dlna.ContentFeatures{
			SupportRange:    true,
			SupportTimeSeek: true,
		}.String())
	}
	// Add support for range requests
	if req.Header.Get("Range") != "" {
		resp.Header().Set("Accept-Ranges", "bytes")
	}
	// // Create a context with timeout if configured
	// ctx := req.Context()
	// if streamTimeout > 0 {
	// 	var cancel context.CancelFunc
	// 	ctx, cancel = context.WithTimeout(ctx, time.Duration(streamTimeout)*time.Second)
	// 	defer cancel()
	// }
	// // Update request with new context
	// req = req.WithContext(ctx)
	// // Handle client disconnections better
	// wrappedResp := &contextResponseWriter{
	// 	ResponseWriter: resp,
	// 	ctx:            ctx,
	// }
	// http.ServeContent(wrappedResp, req, file.Path(), time.Unix(t.Timestamp, 0), reader)

	http.ServeContent(resp, req, file.Path(), time.Unix(t.Timestamp, 0), reader)

	// Auto-save playback position on stream close (resume feature)
	if !isProbe && positionSavingEnabled() {
		saveViewedPosition(t, fileID, file, reader, time.Since(streamStart))
	}

	if sets.BTsets.EnableDebug {
		if clerr != nil {
			log.Printf("[Stream:%d] Disconnect client", streamID)
		} else {
			log.Printf("[Stream:%d] Disconnect client %s:%s", streamID, host, port)
		}
	}
	return nil
}

const (
	// A real viewing session streams for a while; probes and metadata preloads are short.
	minSessionSeconds = 20
	// Reading to within this margin of the end means the file was watched to the end.
	eofMargin = 8 << 20
	// How often the position is refreshed while a stream is still running.
	saveInterval = 30 * time.Second
)

// ProbeMarker is a query flag added to the URL ffprobe is pointed at. A stream carrying
// it is our own probe, so it must never start another probe — without this guard probing
// would stream from ourselves and recurse.
const ProbeMarker = "tsprobe"

// durEntry is a cached media duration, or the time of the last failed attempt to get it.
type durEntry struct {
	seconds float64
	lastTry time.Time
}

// durations caches durations per "hash:fileIndex"; lastSaved holds the last stored byte
// offset per file so an idle (paused) stream is not rewritten over and over.
var (
	durations sync.Map
	lastSaved sync.Map
)

// probeSlot limits probing to one ffprobe process at a time; probeRetryDelay keeps a file
// that cannot be probed from spawning a process on every save.
var probeSlot = make(chan struct{}, 1)

const probeRetryDelay = 10 * time.Minute

func durationKey(hash string, fileID int) string {
	return hash + ":" + strconv.Itoa(fileID)
}

// SetDuration records a media duration discovered elsewhere (preload already runs ffprobe).
func SetDuration(hash string, fileID int, seconds float64) {
	if seconds > 0 {
		durations.Store(durationKey(hash, fileID), durEntry{seconds: seconds})
	}
}

func getDuration(hash string, fileID int) float64 {
	if v, ok := durations.Load(durationKey(hash, fileID)); ok {
		if e, ok := v.(durEntry); ok {
			return e.seconds
		}
	}
	return 0
}

// positionSavingEnabled reports whether playback positions can be saved. It requires
// ffprobe, since the position is stored in seconds and that needs the real duration.
func positionSavingEnabled() bool {
	return sets.BTsets != nil && sets.BTsets.SavePosition && ffprobe.Exists()
}

func probeDuration(hash string, fileID int) {
	key := durationKey(hash, fileID)
	if v, ok := durations.Load(key); ok {
		if e, ok := v.(durEntry); ok {
			if e.seconds > 0 || time.Since(e.lastTry) < probeRetryDelay {
				return // already known, or attempted too recently
			}
		}
	}
	durations.Store(key, durEntry{lastTry: time.Now()}) // claim the attempt
	select {
	case probeSlot <- struct{}{}:
		defer func() { <-probeSlot }()
	default:
		return // another probe is running, try again later
	}

	link := "http://127.0.0.1:" + sets.Port + "/play/" + hash + "/" + strconv.Itoa(fileID)
	if sets.Ssl {
		link = "https://127.0.0.1:" + sets.SslPort + "/play/" + hash + "/" + strconv.Itoa(fileID)
	}
	data, err := ffprobe.ProbeUrl(link + "?" + ProbeMarker + "=1")
	if err != nil || data == nil || data.Format == nil || data.Format.DurationSeconds <= 0 {
		return // the claimed attempt stands; retried after probeRetryDelay
	}
	durations.Store(key, durEntry{seconds: data.Format.DurationSeconds})
}

// saveViewedPosition stores where playback actually was: the read head minus the client's
// buffer, converted to seconds using the real media duration. held is how long the client
// has kept this stream open, which is what separates a viewing session from a quick probe.
func saveViewedPosition(t *Torrent, fileID int, file *torrent.File, reader *torrstor.Reader, held time.Duration) {
	flen := file.Length()
	anchor, ok := reader.Anchor()
	if flen <= 0 || !ok {
		return
	}
	hash := t.Hash().HexString()

	buffer := int64(sets.BTsets.BufferSizeMB) * 1024 * 1024
	if sets.BTsets.AutoBuffer {
		if measured, ok := reader.BufferEstimate(); ok {
			buffer = measured
		}
	}
	if buffer <= 0 {
		buffer = 32 << 20
	}

	head := reader.Offset() // last byte handed to the client = what is on screen + its buffer
	// Ignore probes and metadata preloads: a real session pulls data over a stretch of time
	// and plays past everything it buffered. Two independent time signals are used because
	// either one alone can be misleading: how long the client held the stream depends on TCP
	// back pressure, while the span of actual reads depends on how fast the torrent supplies.
	active := reader.SessionSeconds()
	if held.Seconds() > active {
		active = held.Seconds()
	}
	if active < minSessionSeconds || head-anchor <= buffer {
		return
	}

	screen := head - buffer
	if screen < anchor {
		screen = anchor
	}
	if head >= flen-eofMargin { // watched to the end
		screen = flen
	}
	if screen > flen {
		screen = flen
	}

	// Nothing new to store (e.g. the stream is paused and the position is unchanged).
	key := durationKey(hash, fileID)
	if prev, ok := lastSaved.Load(key); ok {
		if off, ok := prev.(int64); ok && off == screen {
			return
		}
	}
	lastSaved.Store(key, screen)

	// The gates above are passed only by a genuine viewing session, so the duration is
	// looked up (and probed, if still unknown) at most once per watched file. Done in the
	// background: the client is already gone and the reader must not be held open for it.
	debug := sets.BTsets.EnableDebug
	go func() {
		dur := getDuration(hash, fileID)
		if dur <= 0 {
			probeDuration(hash, fileID)
			dur = getDuration(hash, fileID)
		}
		if dur <= 0 { // no real duration => the position cannot be expressed in seconds
			return
		}
		timecode := float64(screen) / float64(flen) * dur
		sets.SetViewed(&sets.Viewed{
			Hash:      hash,
			FileIndex: fileID,
			TimeCode:  timecode,
			Offset:    screen,
			Length:    flen,
			Duration:  dur,
		})
		if debug {
			log.Printf("[Stream] saved position hash=%s idx=%d time=%.0fs/%.0fs (head=%dMB buf=%dMB)",
				hash[:8], fileID, timecode, dur, head>>20, buffer>>20)
		}
	}()
}

// GetActiveStreams returns number of currently active streams
func GetActiveStreams() int32 {
	return atomic.LoadInt32(&activeStreams)
}
