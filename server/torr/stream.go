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
	"strings"
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
	isProbe := req.URL.Query().Get(probeMarker) != ""

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
	// How often the position is refreshed while a stream is still running.
	saveInterval = 30 * time.Second
	// Within this much of the end the film counts as watched to the end, and the position
	// is stored as the duration rather than as a point just short of it, so the film is not
	// offered to be resumed at its credits.
	endMargin = 15.0
)

// probeMarker is a query flag added to the URL ffprobe is pointed at. A stream carrying
// it is our own probe, so it must never start another probe — without this guard probing
// would stream from ourselves and recurse.
const probeMarker = "tsprobe"

// durEntry is a cached media duration, or the time of the last failed attempt to get it.
// start is the timestamp the container gives its first frame — zero for most formats, but
// MPEG-TS clocks begin wherever the muxer left them.
type durEntry struct {
	seconds float64
	start   float64
	lastTry time.Time
}

// Both maps are keyed by fileKey. durations caches media durations; lastSaved holds the
// last stored byte offset per file, so an idle (paused) stream is not rewritten over and over.
var (
	durations sync.Map
	lastSaved sync.Map
)

// probeSlot limits probing to one ffprobe process at a time; probeRetryDelay keeps a file
// that cannot be probed from spawning a process on every save.
var probeSlot = make(chan struct{}, 1)

const probeRetryDelay = 10 * time.Minute

func fileKey(hash string, fileID int) string {
	return hash + ":" + strconv.Itoa(fileID)
}

// forgetTorrent drops what was remembered about a torrent's files once it is closed.
func forgetTorrent(hash string) {
	for _, m := range []*sync.Map{&durations, &lastSaved} {
		m.Range(func(k, _ any) bool {
			if key, ok := k.(string); ok && strings.HasPrefix(key, hash+":") {
				m.Delete(k)
			}
			return true
		})
	}
}

// probeLink is the local stream URL ffprobe reads a file through. It carries the marker that
// tells the stream handler this is a probe, not a viewing session: without it the probe would
// start the position ticker and save on close, over the real resume point.
func probeLink(hash string, fileID int) string {
	link := "http://127.0.0.1:" + sets.Port
	if sets.Ssl {
		link = "https://127.0.0.1:" + sets.SslPort
	}
	return link + "/play/" + hash + "/" + strconv.Itoa(fileID) + "?" + probeMarker + "=1"
}

// setDuration records a media duration discovered elsewhere (preload already runs ffprobe).
func setDuration(hash string, fileID int, seconds float64) {
	if seconds > 0 {
		durations.Store(fileKey(hash, fileID), durEntry{seconds: seconds})
	}
}

func getDuration(hash string, fileID int) float64 {
	seconds, _ := getTiming(hash, fileID)
	return seconds
}

// getTiming returns the media duration and the timestamp its first frame carries.
func getTiming(hash string, fileID int) (seconds, start float64) {
	if v, ok := durations.Load(fileKey(hash, fileID)); ok {
		if e, ok := v.(durEntry); ok {
			return e.seconds, e.start
		}
	}
	return 0, 0
}

// positionSavingEnabled reports whether playback positions can be saved. It requires
// ffprobe, since the position is stored in seconds and that needs the real duration.
func positionSavingEnabled() bool {
	return sets.BTsets != nil && sets.BTsets.SavePosition && ffprobe.Exists()
}

func probeDuration(hash string, fileID int) {
	key := fileKey(hash, fileID)
	if v, ok := durations.Load(key); ok {
		if e, ok := v.(durEntry); ok {
			if e.seconds > 0 || time.Since(e.lastTry) < probeRetryDelay {
				return // already known, or attempted too recently
			}
		}
	}
	select {
	case probeSlot <- struct{}{}:
		defer func() { <-probeSlot }()
	default:
		return // another probe is running, try again later
	}
	// Claimed only once the attempt is actually being made. Recording it before taking the
	// slot means a probe skipped because another was already running still blocks every retry
	// for the next ten minutes, and the position for that file goes unsaved meanwhile.
	durations.Store(key, durEntry{lastTry: time.Now()})

	data, err := ffprobe.ProbeUrl(probeLink(hash, fileID))
	if err != nil || data == nil || data.Format == nil || data.Format.DurationSeconds <= 0 {
		return // the claimed attempt stands; retried after probeRetryDelay
	}
	durations.Store(key, durEntry{seconds: data.Format.DurationSeconds, start: data.Format.StartTimeSeconds})
}

// linearMargin is taken off a position worked out from the average bitrate. Such a guess can
// be a minute out on a file whose bitrate moves, and being early is the harmless direction:
// a resume that lands short replays a few seconds, one that overshoots skips them unseen.
const linearMargin = 15

// safetyMargin is taken off a position read from the container's timestamps, so that a
// reckoning which is only nearly exact still errs behind the picture rather than past it.
//
// Ten seconds rather than a token amount, because of what a client holds that cannot be seen
// from here. Past the buffer being tracked there is a decoder and an output queue, and those
// hold a second or two of picture whatever else is going on. While the buffer is large that
// sits inside the noise; when the buffer runs dry — a supply that cannot keep up, or a server
// restarted under a playing client — it is the whole of the difference, and measured against
// a real player the position then sat two seconds past the picture with five taken off.
const safetyMargin = 10

// playbackTime is the time on screen for a streaming reader.
//
// It prefers the timestamps the container carries, read out of the stream the client is
// already pulling — exact whatever the bitrate does, and free. Only a file that carries none
// falls back to the average bitrate, and that fallback deliberately reports early.
//
// Whichever source is used, the answer is capped by the clock: the picture cannot have moved
// further than real time has since this session started reading. An estimate that claims
// otherwise is wrong in the one direction that matters, skipping past unwatched picture.
func playbackTime(reader *torrstor.Reader, pb torrstor.Playback, hash string, fileID int) (float64, string) {
	file := reader.File()
	if file == nil {
		return 0, ""
	}
	flen := file.Length()
	dur, start := getTiming(hash, fileID)
	index := reader.TimeIndex()
	index.SetOrigin(start)
	index.SetDuration(dur)

	var sec float64
	var source string
	switch {
	case pb.HasTime:
		// A few seconds are taken off. The reckoning is bounded but not exact — timestamps
		// arrive a cluster apart and the position is read between them — and measured against
		// a real player it has come out a fraction of a second past the picture. Landing
		// short replays a moment; landing past skips it unseen, and there is no reason to
		// leave that possible for the sake of three tenths of a second.
		sec, source = pb.ScreenSec-safetyMargin, index.Source()
	case dur > 0 && flen > 0:
		// No timestamps in this container: fall back to the average bitrate, which is a
		// guess, so it is deliberately reported early.
		sec, source = float64(pb.Screen)/float64(flen)*dur-linearMargin, "estimate"
	default:
		return 0, ""
	}

	// The first timestamp at or after the anchor. Asking for the one before it would reach
	// back to the file header, which another connection indexed at time zero, and the
	// ceiling would then clamp every position to the length of the session.
	began, known := index.TimeFrom(pb.Anchor)
	if !known && dur > 0 && flen > 0 {
		began, known = float64(pb.Anchor)/float64(flen)*dur, true
	}
	if ceiling := began + pb.Session; known && sec > ceiling {
		sec = ceiling
	}

	// Whatever this session reckons on its own, the picture cannot have got further than an
	// earlier connection to the same file left it, plus the time since. A player that drops
	// its connection and opens another starts a session that knows nothing of what it holds,
	// and without this the position jumps forward by a whole buffer.
	if furthest, ok := reader.FurthestScreen(); ok && sec > furthest {
		sec = furthest
	}
	if sec < 0 {
		sec = 0
	}
	if dur > 0 && sec > dur-endMargin {
		sec = dur // watched to the end
	}
	return sec, source
}

// PlaybackState reports what a streaming reader is showing, for the UI. It is the same
// arithmetic the saver uses, so what is on screen in the web interface is what would be
// stored. Duration is only filled in when it is already known: probing is the saver's job.
func PlaybackState(hash string, fileID int, reader *torrstor.Reader) *state.PlaybackStatus {
	if reader.File() == nil {
		return nil
	}
	pb := reader.Playback()
	if !pb.Started {
		return nil
	}
	st := &state.PlaybackStatus{
		FileIndex:      fileID,
		Head:           pb.Head,
		Buffer:         pb.Buffer,
		BufferMeasured: pb.HasTime,
		BufferSeconds:  pb.Held,
		SessionSeconds: pb.Session,
		Viewing:        isViewing(pb, pb.Session),
		Position:       pb.Screen,
	}
	if dur := getDuration(hash, fileID); dur > 0 {
		st.Duration = dur
		st.TimeCode, st.Source = playbackTime(reader, pb, hash, fileID)
	}
	return st
}

// screenMoved is how far the picture has travelled from where this session began, in bytes.
// Zero means nothing has been shown yet and there is nothing worth storing.
func screenMoved(pb torrstor.Playback) int64 {
	if !pb.HasTime {
		// Without a measurement the buffer is a guess from the settings, and the old rule is
		// the only one available: the session has to have outrun it.
		return pb.Head - pb.Anchor - pb.Buffer
	}
	return pb.Screen - pb.Anchor
}

// isViewing tells a viewing session from a probe or a metadata preload: it has streamed for
// a while, and its picture has moved past the byte it started on.
//
// The second half used to ask for a whole buffer's worth beyond the anchor, on the grounds
// that a real session plays past everything it buffered. On a high bitrate that is a matter
// of seconds; on a low one it is not. A cartoon episode at four megabits holds three hundred
// megabytes as ten minutes of film, so nothing was stored — and no duration probed — until
// ten minutes of a twenty-two minute episode had gone by. Half the episode had no resume
// point at all. What the condition was really guarding against is a client that has not
// played anything yet, and the picture, worked out from the file's own timestamps, can
// simply be asked whether it has moved.
func isViewing(pb torrstor.Playback, active float64) bool {
	return active >= minSessionSeconds && screenMoved(pb) > 0
}

// saveViewedPosition stores where playback actually was: the read head minus the client's
// buffer, converted to seconds using the real media duration. held is how long the client
// has kept this stream open, which is what separates a viewing session from a quick probe.
func saveViewedPosition(t *Torrent, fileID int, file *torrent.File, reader *torrstor.Reader, held time.Duration) {
	pb := reader.Playback()
	if file.Length() <= 0 || !pb.Started {
		return
	}
	hash := t.Hash().HexString()

	// Two independent time signals, the larger taken, because either one alone can mislead:
	// how long the client held the stream depends on TCP back pressure, while the span of
	// actual reads depends on how fast the torrent supplies.
	if !isViewing(pb, max(pb.Session, held.Seconds())) {
		return
	}

	// Nothing new to store (e.g. the stream is paused and the position is unchanged).
	key := fileKey(hash, fileID)
	if prev, ok := lastSaved.Load(key); ok {
		if off, ok := prev.(int64); ok && off == pb.Screen {
			return
		}
	}
	lastSaved.Store(key, pb.Screen)

	// The gates above are passed only by a genuine viewing session, so the duration is
	// looked up (and probed, if still unknown) at most once per watched file. Done in the
	// background: the client is already gone and the reader must not be held open for it.
	go func() {
		dur := getDuration(hash, fileID)
		if dur <= 0 {
			probeDuration(hash, fileID)
			dur = getDuration(hash, fileID)
		}
		if dur <= 0 { // no real duration => the position cannot be expressed in seconds
			return
		}
		pb := reader.Playback() // the probe may have taken a while; store where it is now
		timecode, _ := playbackTime(reader, pb, hash, fileID)
		sets.SetViewed(&sets.Viewed{Hash: hash, FileIndex: fileID, TimeCode: timecode, Duration: dur})
		how := fmt.Sprintf("%dMB from the average bitrate", pb.Buffer>>20)
		if pb.HasTime {
			how = fmt.Sprintf("%dMB / %.0fs measured", pb.Buffer>>20, pb.Held)
		}
		log.Printf("[Position] %s:%d saved %.0fs of %.0fs (head %dMB, buffer %s)",
			hash[:8], fileID, timecode, dur, pb.Head>>20, how)
	}()
}

// GetActiveStreams returns number of currently active streams
func GetActiveStreams() int32 {
	return atomic.LoadInt32(&activeStreams)
}
