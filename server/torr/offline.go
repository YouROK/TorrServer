package torr

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/anacrolix/dms/dlna"
	"github.com/anacrolix/missinggo/v2/httptoo"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"server/log"
	mt "server/mimetype"
	sets "server/settings"
	"server/torr/offline"
	"server/torr/state"
	utils2 "server/utils"
)

// OnOfflineReady is called (in its own goroutine) when an offline torrent has
// been downloaded completely. Notification channels (Telegram bot, webhook...)
// can subscribe by assigning it.
var OnOfflineReady func(hash, title string)

// offlineBaseDir is the directory where torrents switched to offline mode from now on are stored
// (setting OfflineSavePath; a relative path is relative to the config directory).
func offlineBaseDir() string {
	p := strings.TrimSpace(sets.BTsets.OfflineSavePath)
	if p == "" {
		return filepath.Join(sets.Path, "offline")
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(sets.Path, p)
	}
	return filepath.Clean(p)
}

// offlineDirRaw is the directory stored for this torrent ("" = not set).
func (t *Torrent) offlineDirRaw() string {
	t.muOffline.RLock()
	defer t.muOffline.RUnlock()
	return t.offlineDir
}

func (t *Torrent) setOfflineDir(dir string) {
	t.muOffline.Lock()
	t.offlineDir = dir
	t.muOffline.Unlock()
}

// offlineBase is the directory holding the offline data of this torrent.
func (t *Torrent) offlineBase() string {
	if d := t.offlineDirRaw(); d != "" {
		return d
	}
	return offlineBaseDir()
}

// offlineDirOf finds the offline directory of a torrent by hash.
func offlineDirOf(hash metainfo.Hash) string {
	if db := GetTorrentDB(hash); db != nil {
		return db.offlineBase()
	}
	if bts != nil {
		if live := bts.GetTorrent(hash); live != nil {
			return live.offlineBase()
		}
	}
	return offlineBaseDir()
}

// checkOfflineDir creates the directory and makes sure it is writable.
func checkOfflineDir(dir string) error {
	// 0o755/0o644 instead of the world-writable 0o777/0o666: on a
	// multi-user machine the offline copies must not be writable (or even
	// listable) by everyone (review remark N-4)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("offline directory %q is not usable: %w", dir, err)
	}
	f, err := os.CreateTemp(dir, ".write-test-*")
	if err != nil {
		return fmt.Errorf("offline directory %q is not writable: %w", dir, err)
	}
	name := f.Name()
	f.Close()
	_ = os.Remove(name)
	return nil
}

// offlineSpaceMargin is the extra room required on top of the remaining
// torrent size: the sidecar state, filesystem overhead, and the database and
// logs the server keeps writing while the download runs.
const offlineSpaceMargin = 32 << 20

// checkOfflineSpace refuses to start (or resume) an offline download when the
// filesystem that hosts dir cannot visibly hold the data still missing. It is
// a preflight check, not a guarantee: if the free space cannot be determined
// (exotic mount, restricted permissions), the download is not blocked, and a
// disk that fills up mid-download is still handled as an ordinary write error.
func checkOfflineSpace(dir string, remaining int64) error {
	if remaining <= 0 {
		return nil
	}
	free, err := offline.FreeSpace(dir)
	if err != nil {
		log.TLogln("offline: cannot determine the free space on", dir, err)
		return nil
	}
	need := remaining + offlineSpaceMargin
	if free >= need {
		return nil
	}
	return fmt.Errorf("not enough disk space for the offline download: need %s, available %s (%s)", bytesHuman(need), bytesHuman(free), dir)
}

// bytesHuman formats a byte count for the error messages (1.3 GiB etc.).
func bytesHuman(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

// offlineErrors keeps the reason why a download was stopped by the server (for
// example the disk turned out to be too small once the metadata arrived); the
// status reports it as offline_error. Kept per hash, so it survives the
// closing of the live torrent.
var offlineErrors sync.Map // metainfo.Hash -> string

func setOfflineError(hash metainfo.Hash, msg string) { offlineErrors.Store(hash, msg) }
func clearOfflineError(hash metainfo.Hash)           { offlineErrors.Delete(hash) }
func offlineErrorOf(hash metainfo.Hash) string {
	if v, ok := offlineErrors.Load(hash); ok {
		return v.(string)
	}
	return ""
}

// recheckOfflineSpace repeats the free space check once the metadata is known:
// a torrent added from a magnet link has no size when it is switched to the
// offline mode, so the check in EnableOffline could not run.
func (t *Torrent) recheckOfflineSpace(gt *torrent.Torrent) error {
	var verified int64
	if t.offStore != nil {
		if _, _, b, ok := t.offStore.Stats(t.Hash()); ok {
			verified = b
		}
	}
	return checkOfflineSpace(t.offlineBase(), gt.Length()-verified)
}

// offlineTotalLength returns the full payload size of the torrent, 0 when the
// metadata is not available (yet).
func offlineTotalLength(db *Torrent) int64 {
	if db == nil || db.TorrentSpec == nil || len(db.TorrentSpec.InfoBytes) == 0 {
		return 0
	}
	info, err := offline.LoadInfo(db.TorrentSpec.InfoBytes)
	if err != nil {
		return 0
	}
	return info.TotalLength()
}

// offlineRemainingBytes returns how many bytes of the torrent are still
// missing on disk: the total size minus the verified data from the saved
// piece state. Without a usable saved state the total size is assumed, which
// is only conservative for the space check.
func offlineRemainingBytes(db *Torrent, hash metainfo.Hash) int64 {
	total := offlineTotalLength(db)
	if total <= 0 {
		return 0
	}
	info, err := offline.LoadInfo(db.TorrentSpec.InfoBytes)
	if err != nil {
		return total
	}
	if verified, ok := offline.SidecarProgress(db.offlineBase(), hash, info.PieceLength, info.NumPieces(), total); ok {
		if rest := total - verified; rest > 0 {
			return rest
		}
		return 0
	}
	return total
}

// OfflineState returns "", state.OfflineDownloading or state.OfflineReady.
func (t *Torrent) OfflineState() string {
	t.muOffline.RLock()
	defer t.muOffline.RUnlock()
	return t.offlineState
}

func (t *Torrent) setOfflineState(s string) {
	t.muOffline.Lock()
	t.offlineState = s
	t.muOffline.Unlock()
}

// OfflineReady reports whether the torrent is fully stored on disk and is served from there.
func (t *Torrent) OfflineReady() bool {
	return t.OfflineState() == state.OfflineReady
}

// FileCount returns the number of files of the torrent. Unlike len(Files()) it
// also works for offline torrents that are not loaded into the torrent client.
func (t *Torrent) FileCount() int {
	if t.OfflineReady() {
		if lay, err := t.offlineLayout(); err == nil {
			return len(lay.files)
		}
		return 0
	}
	return len(t.Files())
}

type offlineFile struct {
	Id     int
	Path   string
	Disk   string
	Length int64
}

type offlineLayout struct {
	name     string
	total    int64
	pieceLen int64
	pieces   int
	files    []offlineFile // sorted like Status() sorts them; Id = index + 1
}

func (l *offlineLayout) byID(id int) *offlineFile {
	for i := range l.files {
		if l.files[i].Id == id {
			return &l.files[i]
		}
	}
	return nil
}

var offlineLayouts sync.Map // metainfo.Hash -> *offlineLayout

func (t *Torrent) offlineLayout() (*offlineLayout, error) {
	hash := t.Hash()
	if v, ok := offlineLayouts.Load(hash); ok {
		return v.(*offlineLayout), nil
	}
	if t.TorrentSpec == nil || len(t.TorrentSpec.InfoBytes) == 0 {
		return nil, errors.New("offline torrent has no metadata")
	}
	info, err := offline.LoadInfo(t.TorrentSpec.InfoBytes)
	if err != nil {
		return nil, err
	}
	list := offline.Files(t.offlineBase(), hash, info)
	lay := &offlineLayout{name: info.BestName(), pieceLen: info.PieceLength, pieces: info.NumPieces(),
		files: make([]offlineFile, 0, len(list))}
	for _, f := range list {
		lay.files = append(lay.files, offlineFile{Path: f.Path, Disk: f.Disk, Length: f.Length})
		lay.total += f.Length
	}
	sort.Slice(lay.files, func(i, j int) bool {
		return utils2.CompareStrings(lay.files[i].Path, lay.files[j].Path)
	})
	for i := range lay.files {
		lay.files[i].Id = i + 1
	}
	offlineLayouts.Store(hash, lay)
	return lay, nil
}

// pausedVerified is the size of the verified data of a paused offline download (from the saved state).
func (t *Torrent) pausedVerified(lay *offlineLayout) int64 {
	b, _ := offline.SidecarProgress(t.offlineBase(), t.Hash(), lay.pieceLen, lay.pieces, lay.total)
	return b
}

// fillOfflineStatus fills the status of a ready offline torrent that has no live client torrent.
func (t *Torrent) fillOfflineStatus(st *state.TorrentStatus) {
	lay, err := t.offlineLayout()
	if err != nil {
		return
	}
	st.Name = lay.name
	st.LoadedSize = lay.total
	st.TorrentSize = lay.total
	if t.OfflineState() == state.OfflinePaused {
		st.LoadedSize = t.pausedVerified(lay)
	}
	for _, f := range lay.files {
		st.FileStats = append(st.FileStats, &state.TorrentFileStat{Id: f.Id, Path: f.Path, Length: f.Length})
	}
}

// streamOffline serves a file of a ready offline torrent straight from disk.
func (t *Torrent) streamOffline(fileID int, req *http.Request, resp http.ResponseWriter) error {
	lay, err := t.offlineLayout()
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return err
	}
	file := lay.byID(fileID)
	if file == nil {
		err := fmt.Errorf("file with id %v not found", fileID)
		http.Error(resp, err.Error(), http.StatusNotFound)
		return err
	}
	if int64(sets.MaxSize) > 0 && file.Length > int64(sets.MaxSize) {
		err := fmt.Errorf("file size exceeded max allowed %d bytes", sets.MaxSize)
		http.Error(resp, err.Error(), http.StatusForbidden)
		return err
	}
	fh, err := os.Open(file.Disk)
	if err != nil {
		log.TLogln("offline: cannot open", file.Disk, err)
		http.Error(resp, "offline file is not available", http.StatusNotFound)
		return err
	}
	defer fh.Close()

	t.prepareStreamResponse(resp, req, fileID, file.Path)
	http.ServeContent(resp, req, file.Path, time.Unix(t.Timestamp, 0), fh)
	return nil
}

// readOfflineFile is ReadFile for a ready offline torrent.
func (t *Torrent) readOfflineFile(fileID int, maxSize int64) ([]byte, error) {
	lay, err := t.offlineLayout()
	if err != nil {
		return nil, err
	}
	file := lay.byID(fileID)
	if file == nil {
		return nil, fmt.Errorf("file with id %v not found", fileID)
	}
	if file.Length > maxSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d)", file.Length, maxSize)
	}
	fh, err := os.Open(file.Disk)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	buf := make([]byte, file.Length)
	if _, err := io.ReadFull(fh, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// prepareStreamResponse marks the file as viewed and sets the response headers
// shared by every offline stream (from disk and from a running download).
func (t *Torrent) prepareStreamResponse(resp http.ResponseWriter, req *http.Request, fileID int, path string) {
	hash := t.Hash().HexString()

	// Mark as viewed without resetting existing TimeCode (same as Stream)
	var timecode float64
	for _, v := range sets.ListViewed(hash) {
		if v.FileIndex == fileID {
			timecode = v.TimeCode
			break
		}
	}
	sets.SetViewed(&sets.Viewed{Hash: hash, FileIndex: fileID, TimeCode: timecode})

	if sets.BTsets.EnableDebug {
		host, port, clerr := net.SplitHostPort(req.RemoteAddr)
		if clerr == nil {
			log.TLogln("[Offline stream] connect", host+":"+port, path)
		}
	}

	resp.Header().Set("Connection", "close")
	resp.Header().Set("Server", "TorrServer (Portable SDK for UPnP devices)")
	if timeout := sets.BTsets.TorrentDisconnectTimeout; timeout > 0 {
		resp.Header().Set("X-Stream-Timeout", fmt.Sprintf("%d", timeout))
	}
	etag := hex.EncodeToString([]byte(fmt.Sprintf("%s/%s", hash, path)))
	resp.Header().Set("ETag", httptoo.EncodeQuotedString(etag))
	resp.Header().Set("transferMode.dlna.org", "Streaming")
	if mime, merr := mt.MimeTypeByPath(path); merr == nil && mime.IsMedia() {
		resp.Header().Set("content-type", mime.String())
	}
	if req.Header.Get("getContentFeatures.dlna.org") != "" {
		resp.Header().Set("contentFeatures.dlna.org", dlna.ContentFeatures{
			SupportRange:    true,
			SupportTimeSeek: true,
		}.String())
	}
	if req.Header.Get("Range") != "" {
		resp.Header().Set("Accept-Ranges", "bytes")
	}
}

// offlineStreamReadahead is how far ahead of the player the client fetches
// pieces while an offline torrent is being downloaded and watched at once.
const offlineStreamReadahead = 16 << 20

// closeOnce makes closing a reader safe from several places (the deferred
// close and the watchdog goroutine that aborts a blocked read).
func closeOnce(c io.Closer) func() {
	var once sync.Once
	return func() { once.Do(func() { _ = c.Close() }) }
}

// offlineLiveFile returns the client file with the given id of a torrent that
// is being downloaded. Ids are the positions in the list sorted by path, the
// same order the status reports them in.
func (t *Torrent) offlineLiveFile(fileID int) (*torrent.File, error) {
	t.muTorrent.Lock()
	gt := t.Torrent
	t.muTorrent.Unlock()
	if gt == nil || gt.Info() == nil {
		return nil, errors.New("torrent doesn't have info yet")
	}
	files := append([]*torrent.File(nil), gt.Files()...)
	sort.Slice(files, func(i, j int) bool {
		return utils2.CompareStrings(files[i].Path(), files[j].Path())
	})
	if fileID < 1 || fileID > len(files) {
		return nil, fmt.Errorf("file with id %v not found", fileID)
	}
	return files[fileID-1], nil
}

// streamOfflineLive plays a file of an offline torrent that is still being
// downloaded. The offline storage has no stream cache, so the file is read
// through the reader of the BitTorrent client: it requests the pieces under
// the playback position with priority (on top of the sequential download of
// the whole torrent) and reads the pieces already verified from disk.
func (t *Torrent) streamOfflineLive(fileID int, req *http.Request, resp http.ResponseWriter) error {
	atomic.AddInt32(&activeStreams, 1)
	defer atomic.AddInt32(&activeStreams, -1)

	if !t.GotInfo() {
		http.NotFound(resp, req)
		return errors.New("torrent doesn't have info yet")
	}
	file, err := t.offlineLiveFile(fileID)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusNotFound)
		return err
	}
	if int64(sets.MaxSize) > 0 && file.Length() > int64(sets.MaxSize) {
		err := fmt.Errorf("file size exceeded max allowed %d bytes", sets.MaxSize)
		http.Error(resp, err.Error(), http.StatusForbidden)
		return err
	}

	reader := file.NewReader()
	reader.SetResponsive()
	reader.SetReadahead(offlineStreamReadahead)
	closeReader := closeOnce(reader)
	defer closeReader()

	// A read waiting for a piece blocks until the piece arrives: close the
	// reader when the client disconnects or the torrent is closed, which
	// makes the blocked read return.
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-req.Context().Done():
		case <-t.closed:
		case <-done:
			return
		}
		closeReader()
	}()

	t.prepareStreamResponse(resp, req, fileID, file.Path())
	http.ServeContent(resp, req, file.Path(), time.Unix(t.Timestamp, 0), reader)
	return nil
}

// readOfflineLiveFile is ReadFile for an offline torrent that is still being
// downloaded (subtitles and other small files).
func (t *Torrent) readOfflineLiveFile(ctx context.Context, fileID int, maxSize int64) ([]byte, error) {
	file, err := t.offlineLiveFile(fileID)
	if err != nil {
		return nil, err
	}
	if file.Length() > maxSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d)", file.Length(), maxSize)
	}
	reader := file.NewReader()
	reader.SetResponsive()
	closeReader := closeOnce(reader)
	defer closeReader()

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
		case <-t.closed:
		case <-done:
			return
		}
		closeReader()
	}()

	buf := make([]byte, file.Length())
	if _, err := io.ReadFull(reader, buf); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, err
	}
	return buf, nil
}

const (
	// Amount of data that is requested at the same time. Pieces are started roughly in order
	// instead of all at once, so only a bounded amount of data sits in unfinished pieces (the
	// client forgets those after a restart and downloads them again).
	offlineWindowBytes = 128 << 20
	// The window is widened when no piece gets verified for this long (scarce swarm).
	offlineWindowStall = 20 * time.Second
)

// offlineWindow limits the pieces that are requested to the first few unverified ones.
type offlineWindow struct {
	base, size   int
	total        int
	prioritized  map[int]struct{}
	lastVerified int
	lastProgress time.Time
}

func newOfflineWindow(pieceLen int64, pieces int) *offlineWindow {
	base := 8
	if pieceLen > 0 {
		if n := int(offlineWindowBytes / pieceLen); n > base {
			base = n
		}
	}
	if base > pieces {
		base = pieces
	}
	return &offlineWindow{base: base, size: base, total: pieces,
		prioritized: make(map[int]struct{}), lastProgress: time.Now()}
}

// progress updates the window size bookkeeping: on any new verified piece the
// window collapses back to its base size, after offlineWindowStall without
// progress it doubles (scarce swarm) up to the number of pieces of the torrent.
// It is separate from update() so it can be tested without a torrent client.
func (w *offlineWindow) progress(complete int, now time.Time) {
	if complete > w.lastVerified {
		w.lastVerified = complete
		w.lastProgress = now
		w.size = w.base
	} else if now.Sub(w.lastProgress) > offlineWindowStall && w.size < w.total {
		w.size *= 2
		if w.size > w.total {
			w.size = w.total
		}
		w.lastProgress = now
	}
}

func (w *offlineWindow) update(gt *torrent.Torrent, off *offline.Storage, hash metainfo.Hash) {
	complete, _, _, ok := off.Stats(hash)
	if !ok {
		return
	}
	w.progress(complete, time.Now())

	want := off.NextIncomplete(hash, w.size)
	wantSet := make(map[int]struct{}, len(want))
	for _, i := range want {
		wantSet[i] = struct{}{}
		if _, ok := w.prioritized[i]; !ok {
			gt.Piece(i).SetPriority(torrent.PiecePriorityNormal)
			w.prioritized[i] = struct{}{}
		}
	}
	for i := range w.prioritized {
		if _, ok := wantSet[i]; !ok {
			gt.Piece(i).SetPriority(torrent.PiecePriorityNone)
			delete(w.prioritized, i)
		}
	}
}

// runOffline drives the download of an offline torrent: waits (without a
// timeout) for the metadata, asks for all pieces and finishes the torrent when
// everything is on disk. It is started once per live Torrent by NewTorrent.
func (t *Torrent) runOffline() {
	if !t.GotInfo() {
		return
	}
	t.muTorrent.Lock()
	gt := t.Torrent
	t.muTorrent.Unlock()
	if gt == nil {
		return
	}
	if err := t.recheckOfflineSpace(gt); err != nil {
		// not enough room for what is still missing: stop instead of dying on ENOSPC later
		hash := t.Hash()
		log.TLogln("offline: pausing", hash.HexString(), err)
		setOfflineError(hash, err.Error())
		go func() {
			if _, perr := PauseOffline(hash.HexString()); perr != nil {
				log.TLogln("offline: cannot pause", hash.HexString(), perr)
			}
		}()
		return
	}
	log.TLogln("offline: start download", t.Hash().HexString(), "dir", t.offlineBase(),
		"completed", gt.BytesCompleted(), "of", gt.Length())
	var win *offlineWindow
	if info := gt.Info(); t.offStore != nil && info != nil {
		win = newOfflineWindow(info.PieceLength, info.NumPieces())
		win.update(gt, t.offStore, t.Hash())
	} else {
		gt.DownloadAll()
	}

	lastLog := time.Now()
	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-t.closed:
			log.TLogln("offline: torrent closed", t.Hash().HexString(), "completed", gt.BytesCompleted())
			return
		case <-tick.C:
			if t.OfflineState() != state.OfflineDownloading {
				return
			}
			if win != nil {
				win.update(gt, t.offStore, t.Hash())
			}
			if time.Since(lastLog) >= 15*time.Second {
				lastLog = time.Now()
				verified := ""
				if t.offStore != nil {
					if c, n, _, ok := t.offStore.Stats(t.Hash()); ok {
						wsz := 0
						if win != nil {
							wsz = win.size
						}
						verified = fmt.Sprintf("| verified pieces %d of %d, window %d", c, n, wsz)
					}
				}
				log.TLogln("offline: progress", t.Hash().HexString(), gt.BytesCompleted(), "of", gt.Length(), verified)
			}
			if gt.Info() != nil && gt.BytesMissing() == 0 {
				if t.finishOffline() {
					return
				}
				// the data could not be flushed to the disk: try again on the next tick
			}
		}
	}
}

// finishOffline declares a fully downloaded torrent ready. It returns false
// only when the download must be retried because the downloaded data could not
// be flushed to the disk; true means there is nothing more for the caller to do.
func (t *Torrent) finishOffline() bool {
	unlock := lockOfflineTransition(t.Hash())
	defer unlock()
	if t.OfflineState() != state.OfflineDownloading {
		// the download was paused or disabled while we were finishing: never
		// overwrite the state the user has chosen
		return true
	}
	hash := t.Hash()
	title := t.Title

	// The data and the piece state must be on stable storage BEFORE the torrent
	// is declared ready: after a power loss a ready torrent is only checked by
	// the file sizes, so unflushed pieces would turn into silent corruption.
	if t.offStore != nil {
		if err := t.offStore.Flush(hash); err != nil {
			log.TLogln("offline: cannot flush the downloaded data to the disk, will retry", hash.HexString(), err)
			return false
		}
	}

	t.setOfflineState(state.OfflineReady)
	clearOfflineError(hash)
	AddTorrentDB(t) // persists the ready state together with the metadata
	log.TLogln("offline: ready", hash.HexString(), title)

	// Leave the BitTorrent network: from now on the data is served from disk.
	// Closing the torrent also closes its storage, which saves the completion bitmap.
	t.Close()

	if cb := OnOfflineReady; cb != nil {
		go cb(hash.HexString(), title)
	}
	return true
}

func startOffline(db *Torrent) (*Torrent, error) {
	for attempt := 0; attempt < 3; attempt++ {
		tr, err := NewTorrent(db.TorrentSpec, bts)
		if err != nil {
			return nil, err
		}
		if tr.OfflineState() != state.OfflineDownloading {
			// A regular torrent got registered in the meantime (for example a status request
			// loaded it from the DB right after the old instance was closed). Its storage
			// cannot be replaced, so drop it and add the torrent again with the offline storage.
			log.TLogln("offline: torrent was loaded in the normal mode, reloading", db.Hash().HexString())
			closeAndWait(tr)
			continue
		}
		tr.Title = db.Title
		tr.Poster = db.Poster
		tr.Category = db.Category
		tr.Data = db.Data
		tr.Size = db.Size
		if db.Timestamp != 0 {
			tr.Timestamp = db.Timestamp
		}
		return tr, nil
	}
	return nil, errors.New("cannot start the offline download, try again")
}

func closeAndWait(t *Torrent) {
	closed := t.closed
	if bts.RemoveTorrent(t.Hash()) {
		select {
		case <-closed:
		case <-time.After(5 * time.Second):
			log.TLogln("Warning: timeout waiting for torrent close:", t.Hash().HexString())
		}
	}
}

func removeOfflineData(hash metainfo.Hash, dir string) {
	offlineLayouts.Delete(hash)
	clearOfflineError(hash)
	if bts != nil {
		bts.mu.RLock()
		off := bts.offline
		bts.mu.RUnlock()
		if off != nil {
			off.CloseHash(hash)
		}
	}
	if err := offline.RemoveData(dir, hash); err != nil {
		log.TLogln("offline: cannot remove data:", err)
	}
}

// restartOfflineDownload switches a torrent whose offline data went missing back to
// downloading and starts the download again. Review remark V-1: the saved piece
// state must be removed together with the data — the stale sidecar marks every
// piece complete, so the new download would "finish" instantly with no files on
// disk and the torrent would be reported ready while nothing can be played.
func restartOfflineDownload(db *Torrent, hash metainfo.Hash) (*Torrent, error) {
	log.TLogln("offline: files are missing, downloading again", hash.HexString())
	// closed outside of the transition lock: finishOffline of a concurrently
	// finishing download needs the lock, and closeAndWait waits for it
	if live := bts.GetTorrent(hash); live != nil && live.OfflineState() != "" {
		closeAndWait(live)
	}
	unlock := lockOfflineTransition(hash)
	defer unlock()
	// The download starts from zero, so the whole torrent must fit; check
	// before anything is removed, so that a full disk does not destroy the
	// old (stale) data with nowhere to download to.
	// The lock is released by the deferred unlock only: a second explicit
	// unlock() here would unlock an already unlocked mutex, which is a fatal,
	// unrecoverable runtime error (and, if another goroutine took the lock in
	// between, would release the lock of that goroutine).
	if err := checkOfflineSpace(db.offlineBase(), offlineTotalLength(db)); err != nil {
		return nil, err
	}
	removeOfflineData(hash, db.offlineBase()) // removes the sidecar, the layout cache and the files
	db.setOfflineState(state.OfflineDownloading)
	AddTorrentDB(db)
	return startOffline(db)
}

// offlineTransitions serialises the offline state transitions of one torrent
// (review remark S-3): EnableOffline, PauseOffline, ResumeOfflineDownload,
// DisableOffline and finishOffline all read the current state and write a new
// one. Without a lock two transitions interleave: a pause racing a download
// that finishes at the same moment would downgrade the ready state to paused,
// a disable racing a finish could resurrect the offline mode. The lock is held
// only for the read-check-write section on the DB; closing the torrent happens
// outside of it, so nobody waits for a torrent close while holding the lock.
var offlineTransitions sync.Map // metainfo.Hash -> *sync.Mutex

func lockOfflineTransition(hash metainfo.Hash) (unlock func()) {
	v, _ := offlineTransitions.LoadOrStore(hash, &sync.Mutex{})
	mu := v.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
}

// EnableOffline switches the torrent to offline mode: it is downloaded
// completely into the offline directory and, once finished, served from there.
func EnableOffline(hashHex string) (*Torrent, error) {
	if sets.ReadOnly {
		return nil, errors.New("read-only DB mode")
	}
	hash := metainfo.NewHashFromHex(hashHex)

	live := bts.GetTorrent(hash)
	db := GetTorrentDB(hash)
	if db == nil && live != nil {
		SaveTorrentToDB(live)
		db = GetTorrentDB(hash)
	}
	if db == nil || db.TorrentSpec == nil {
		return nil, errors.New("torrent not found, add it first")
	}

	// The transition lock serialises this decision against the other offline
	// transitions and against finishOffline (review remark S-3).
	unlock := lockOfflineTransition(hash)
	switch db.OfflineState() {
	case state.OfflineReady:
		// The torrent is marked ready, but its files may be gone (manual
		// cleanup, disk failure, moved directory) while the server is running:
		// verify instead of returning a torrent that would only 404 on play
		// (review remark S-4). Safe since V-1: restartOfflineDownload removes
		// the stale piece state together with the data.
		if db.TorrentSpec == nil || len(db.TorrentSpec.InfoBytes) == 0 {
			unlock()
			return db, nil
		}
		if info, err := offline.LoadInfo(db.TorrentSpec.InfoBytes); err != nil || offline.FilesIntact(db.offlineBase(), hash, info) {
			unlock()
			return db, nil
		}
		unlock()
		return restartOfflineDownload(db, hash)
	case state.OfflinePaused:
		unlock() // ResumeOfflineDownload takes the lock itself
		return ResumeOfflineDownload(hashHex)
	case state.OfflineDownloading:
		unlock()
		if live != nil {
			return live, nil
		}
		return startOffline(db)
	}

	// New offline torrents use the directory from the settings; it is remembered per torrent.
	dir := offlineBaseDir()
	if err := checkOfflineDir(dir); err != nil {
		unlock()
		return nil, err
	}
	// A fresh offline download needs room for the whole torrent: refuse the
	// switch before the state is written and the client is reloaded, so the
	// user gets a clear error instead of a download that dies on ENOSPC.
	if err := checkOfflineSpace(dir, offlineTotalLength(db)); err != nil {
		unlock()
		return nil, err
	}
	// The state is saved BEFORE the live torrent is closed: from that moment on anything that
	// (re)loads the torrent from the DB, e.g. a status request of the web UI, already creates it
	// with the offline storage.
	clearOfflineError(hash)
	db.setOfflineDir(dir)
	db.setOfflineState(state.OfflineDownloading)
	AddTorrentDB(db)
	unlock()

	// The storage of a torrent cannot be replaced while it is loaded: reload it.
	if live != nil {
		closeAndWait(live)
	}
	log.TLogln("offline: enable", hash.HexString(), dir)
	return startOffline(db)
}

// PauseOffline stops a running offline download. The data downloaded so far stays on disk and
// the torrent leaves the BitTorrent network until ResumeOfflineDownload is called.
func PauseOffline(hashHex string) (*Torrent, error) {
	if sets.ReadOnly {
		return nil, errors.New("read-only DB mode")
	}
	hash := metainfo.NewHashFromHex(hashHex)
	// The transition lock is held for the whole check-and-write: a download
	// finishing at the same moment (finishOffline) either completed under this
	// lock - then the fresh DB state is ready and is never downgraded to
	// paused - or completes after this write, which is equally correct
	// (review remark S-3).
	unlock := lockOfflineTransition(hash)
	db := GetTorrentDB(hash)
	if db == nil {
		unlock()
		return nil, errors.New("torrent not found")
	}
	switch db.OfflineState() {
	case state.OfflinePaused:
		unlock()
		return db, nil
	case state.OfflineDownloading:
	default:
		unlock()
		return nil, errors.New("the torrent is not being downloaded for offline viewing")
	}

	// The paused state is saved BEFORE the live torrent is closed (see EnableOffline). When the
	// torrent is loaded, it is saved as it is, so the metadata fetched so far is kept as well.
	live := bts.GetTorrent(hash)
	if live != nil && live.OfflineState() == state.OfflineDownloading {
		live.setOfflineState(state.OfflinePaused)
		AddTorrentDB(live)
	} else {
		db.setOfflineState(state.OfflinePaused)
		AddTorrentDB(db)
	}
	unlock()
	if live != nil {
		closeAndWait(live) // closing the torrent also saves the verified pieces
	}
	log.TLogln("offline: paused", hash.HexString())
	if paused := GetTorrentDB(hash); paused != nil {
		return paused, nil
	}
	return db, nil
}

// ResumeOfflineDownload continues a paused offline download.
func ResumeOfflineDownload(hashHex string) (*Torrent, error) {
	if sets.ReadOnly {
		return nil, errors.New("read-only DB mode")
	}
	hash := metainfo.NewHashFromHex(hashHex)
	unlock := lockOfflineTransition(hash)
	db := GetTorrentDB(hash)
	if db == nil || db.TorrentSpec == nil {
		unlock()
		return nil, errors.New("torrent not found")
	}
	switch db.OfflineState() {
	case state.OfflineDownloading:
		unlock()
		if live := bts.GetTorrent(hash); live != nil {
			return live, nil
		}
		return startOffline(db)
	case state.OfflinePaused:
	default:
		unlock()
		return nil, errors.New("the offline download is not paused")
	}
	if err := checkOfflineDir(db.offlineBase()); err != nil {
		unlock()
		return nil, err
	}
	// Only the still-missing part counts: a half-downloaded torrent resumed on
	// a full disk must fail with a clear message instead of ENOSPC halfway
	// through a write.
	if err := checkOfflineSpace(db.offlineBase(), offlineRemainingBytes(db, hash)); err != nil {
		unlock()
		return nil, err
	}
	live := bts.GetTorrent(hash) // may be loaded for online use while the download was paused
	clearOfflineError(hash)
	db.setOfflineState(state.OfflineDownloading)
	AddTorrentDB(db)
	unlock()
	if live != nil {
		closeAndWait(live)
	}
	log.TLogln("offline: resume", hash.HexString())
	return startOffline(db)
}

// DisableOffline switches the torrent back to the normal (streaming) mode and
// deletes the offline data.
func DisableOffline(hashHex string) (*Torrent, error) {
	if sets.ReadOnly {
		return nil, errors.New("read-only DB mode")
	}
	hash := metainfo.NewHashFromHex(hashHex)
	unlock := lockOfflineTransition(hash)
	db := GetTorrentDB(hash)
	if db == nil {
		unlock()
		return nil, errors.New("torrent not found")
	}
	dir := db.offlineBase()
	if db.OfflineState() != "" {
		// saved first, so nothing reloads the torrent in offline mode while it is being switched off;
		// finishOffline waits for this lock and skips itself once the state is not downloading
		db.setOfflineState("")
		db.setOfflineDir("")
		AddTorrentDB(db)
	}
	live := bts.GetTorrent(hash)
	liveIsOffline := live != nil && live.OfflineState() != ""
	if liveIsOffline {
		live.setOfflineState("")
	}
	unlock()
	if liveIsOffline {
		closeAndWait(live)
	}
	removeOfflineData(hash, dir)
	return db, nil
}

// ResumeOffline restarts unfinished offline downloads and re-checks finished
// ones. It runs after every (re)connect of the torrent client.
func ResumeOffline() {
	// Connect starts this before the DB exists in tests (and possibly before
	// it is open in embedded setups): without a DB there is nothing to resume.
	if bts == nil || !sets.IsDBOpen() {
		return
	}
	for hash, db := range ListTorrentsDB() {
		if st := db.OfflineState(); st != "" {
			log.TLogln("offline: resume check", hash.HexString(), "state", st, "dir", db.offlineBase())
		}
		switch db.OfflineState() {
		case state.OfflineDownloading:
			if bts.GetTorrent(hash) == nil {
				if _, err := startOffline(db); err != nil {
					log.TLogln("offline: cannot resume", hash.HexString(), err)
				}
			}
		case state.OfflinePaused:
			// stays paused until the user resumes it
		case state.OfflineReady:
			if db.TorrentSpec == nil || len(db.TorrentSpec.InfoBytes) == 0 {
				continue
			}
			info, err := offline.LoadInfo(db.TorrentSpec.InfoBytes)
			if err != nil || offline.FilesIntact(db.offlineBase(), hash, info) {
				continue
			}
			if _, err := restartOfflineDownload(db, hash); err != nil {
				log.TLogln("offline: cannot resume", hash.HexString(), err)
			}
		}
	}
}
