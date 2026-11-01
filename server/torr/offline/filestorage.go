package offline

import (
	"container/list"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"

	"server/log"
)

const (
	maxOpenFiles  = 64
	flushInterval = 2 * time.Second
)

var errClosed = errors.New("offline storage closed")

// Storage is a storage.ClientImpl that keeps torrent data as ordinary files.
// It is passed to the torrent client per torrent through TorrentSpec.Storage.
type Storage struct {
	base string

	mu       sync.Mutex
	torrents map[metainfo.Hash]*torrentStore
	dirs     map[metainfo.Hash]string // per-torrent base directory
}

// NewStorage creates a storage; base is the default directory for torrents
// that have no directory of their own (see Use).
func NewStorage(base string) *Storage {
	return &Storage{
		base:     base,
		torrents: make(map[metainfo.Hash]*torrentStore),
		dirs:     make(map[metainfo.Hash]string),
	}
}

func (s *Storage) Base() string { return s.base }

// Use makes the torrent keep its data under dir (set before the torrent is added to the client).
func (s *Storage) Use(ih metainfo.Hash, dir string) {
	s.mu.Lock()
	s.dirs[ih] = dir
	s.mu.Unlock()
}

func (s *Storage) baseFor(ih metainfo.Hash) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if d := s.dirs[ih]; d != "" {
		return d
	}
	return s.base
}

func (s *Storage) OpenTorrent(info *metainfo.Info, ih metainfo.Hash) (storage.TorrentImpl, error) {
	base := s.baseFor(ih)
	if err := os.MkdirAll(TorrentRoot(base, ih), 0o755); err != nil {
		return nil, err
	}
	ts := newTorrentStore(s, base, info, ih)
	s.mu.Lock()
	old := s.torrents[ih]
	s.torrents[ih] = ts
	s.mu.Unlock()
	if old != nil {
		_ = old.Close()
	}
	return ts, nil
}

// CloseHash closes all file handles of one torrent (needed before its files
// can be deleted on Windows).
func (s *Storage) CloseHash(ih metainfo.Hash) {
	s.mu.Lock()
	ts := s.torrents[ih]
	s.mu.Unlock()
	if ts != nil {
		_ = ts.Close()
	}
}

// Flush saves the piece state of an open torrent after the downloaded data has
// been fsynced; it returns an error when either step failed.
func (s *Storage) Flush(ih metainfo.Hash) error {
	s.mu.Lock()
	ts := s.torrents[ih]
	s.mu.Unlock()
	if ts == nil {
		return nil
	}
	return ts.saveCompletion()
}

// Stats returns the number of verified pieces, the number of pieces and the size of the
// verified data of an open torrent.
func (s *Storage) Stats(ih metainfo.Hash) (complete, total int, verified int64, ok bool) {
	s.mu.Lock()
	ts := s.torrents[ih]
	s.mu.Unlock()
	if ts == nil {
		return 0, 0, 0, false
	}
	return ts.completeCount(), len(ts.state), ts.verifiedBytes(), true
}

// NextIncomplete returns the indexes of the first n pieces that are not verified yet.
func (s *Storage) NextIncomplete(ih metainfo.Hash, n int) []int {
	s.mu.Lock()
	ts := s.torrents[ih]
	s.mu.Unlock()
	if ts == nil || n <= 0 {
		return nil
	}
	ts.cmu.Lock()
	defer ts.cmu.Unlock()
	out := make([]int, 0, n)
	for i, st := range ts.state {
		if st != pieceComplete {
			out = append(out, i)
			if len(out) >= n {
				break
			}
		}
	}
	return out
}

func (s *Storage) Close() error {
	s.mu.Lock()
	list := make([]*torrentStore, 0, len(s.torrents))
	for _, ts := range s.torrents {
		list = append(list, ts)
	}
	s.mu.Unlock()
	for _, ts := range list {
		_ = ts.Close()
	}
	return nil
}

func (s *Storage) forget(ts *torrentStore) {
	s.mu.Lock()
	if s.torrents[ts.hash] == ts {
		delete(s.torrents, ts.hash)
	}
	s.mu.Unlock()
}

// Piece states.
const (
	pieceUnknown     = 0 // nothing is known, the client verifies it by hashing
	pieceComplete    = 1
	pieceNotComplete = 2
)

// fileHandle is one open file of a torrent. refs counts the users of the
// handle right now: a pinned handle (refs > 0) is never closed by the
// eviction and is used for its I/O without holding any lock - os.File.ReadAt
// and WriteAt take an explicit offset and are safe for concurrent use.
type fileHandle struct {
	f    *os.File
	rw   bool
	refs int
	elem *list.Element // position in the LRU list, nil for a non-cached handle
}

type torrentStore struct {
	owner    *Storage
	hash     metainfo.Hash
	files    []File
	pieceLen int64
	total    int64

	// hmu guards handles, lru and closed. The handles are acquired and
	// released under the lock, but used for their I/O without it: reads of
	// different (and even the same) files run in parallel, which the single
	// I/O mutex of the first version serialized for the whole torrent
	// (review remark N-1). The eviction is least-recently-used
	// (review remark N-2).
	hmu     sync.Mutex
	handles map[int]*fileHandle
	lru     *list.List // front = most recently used, values are file indexes
	closed  bool

	// smu guards dirtyFiles: the files written to since the last flush. They are
	// fsynced before the piece state that vouches for their data is saved.
	smu        sync.Mutex
	dirtyFiles map[int]struct{}

	// cmu guards the completion state.
	cmu        sync.Mutex
	state      []uint8
	dirty      bool
	sidecar    string
	lastErrLog time.Time
	stopOnce   sync.Once
	closeOnce  sync.Once
	stop       chan struct{}
}

func newTorrentStore(owner *Storage, base string, info *metainfo.Info, ih metainfo.Hash) *torrentStore {
	ts := &torrentStore{
		owner:    owner,
		hash:     ih,
		files:    Files(base, ih, info),
		pieceLen: info.PieceLength,
		handles:  make(map[int]*fileHandle),
		lru:      list.New(),
		state:    make([]uint8, info.NumPieces()),
		sidecar:  sidecarPath(base, ih),
		stop:     make(chan struct{}),
	}
	if n := len(ts.files); n > 0 {
		ts.total = ts.files[n-1].Offset + ts.files[n-1].Length
	}
	done, found := ts.loadCompletion()
	if found && done > 0 && done == len(ts.state) && !ts.filesIntact() {
		// The saved state claims a complete download, but the files on disk are
		// missing or have unexpected sizes: the state is stale (the data was
		// deleted, lost or the directory was moved). Distrust it and verify
		// every piece by hashing; otherwise the download would "finish"
		// instantly with the same broken state (review remark V-1).
		log.TLogln("offline: saved state says complete but the files on disk do not match, verifying by hashing", ih.HexString())
		for i := range ts.state {
			ts.state[i] = pieceUnknown
		}
		done, found = 0, false
	}
	if found {
		log.TLogln("offline: opened", ih.HexString(), "- verified pieces from the saved state:", done, "of", len(ts.state))
	} else if !ts.anyFileOnDisk() {
		// a new download: there is nothing to verify, do not hash every piece
		for i := range ts.state {
			ts.state[i] = pieceNotComplete
		}
		ts.dirty = true
		log.TLogln("offline: opened", ih.HexString(), "- new download")
	} else {
		log.TLogln("offline: opened", ih.HexString(), "- no saved state, pieces found on disk are verified by hashing")
	}
	go ts.flushLoop()
	return ts
}

// anyFileOnDisk reports whether at least one file of the torrent already exists.
func (ts *torrentStore) anyFileOnDisk() bool {
	for _, f := range ts.files {
		if _, err := os.Stat(f.Disk); err == nil {
			return true
		}
	}
	return false
}

// filesIntact reports whether every non-empty file exists with the expected size.
// It is the torrentStore-side equivalent of FilesIntact (layout.go) and works on
// the file list computed at open time.
func (ts *torrentStore) filesIntact() bool {
	for _, f := range ts.files {
		if f.Length == 0 {
			continue
		}
		st, err := os.Stat(f.Disk)
		if err != nil || st.IsDir() || st.Size() != f.Length {
			return false
		}
	}
	return true
}

func sidecarPath(base string, hash metainfo.Hash) string {
	return filepath.Join(base, hash.HexString()+".pieces")
}

const sidecarMagic = "TSOFF1\n"

// loadCompletion restores the verified-pieces bitmap saved earlier. The saved state is trusted
// in both directions: a piece that is not marked complete is downloaded again without hashing
// it first. (Hashing every unfinished piece at start made a resumed download crawl for minutes
// and could crash the client when the torrent was closed meanwhile.) Without a saved state the
// client hashes every piece it finds on disk.
func (ts *torrentStore) loadCompletion() (complete int, found bool) {
	buf, err := os.ReadFile(ts.sidecar)
	if err != nil {
		return 0, false
	}
	bits, ok := parseSidecar(buf, len(ts.state))
	if !ok {
		return 0, false // missing or damaged (e.g. torn write): verify by hashing instead
	}
	for i, b := range bits {
		if b == '1' {
			ts.state[i] = pieceComplete
			complete++
		} else {
			ts.state[i] = pieceNotComplete
		}
	}
	return complete, true
}

// parseSidecar returns the bitmap of a saved state (current format with a CRC, or the legacy one).
func parseSidecar(buf []byte, pieces int) ([]byte, bool) {
	bits := buf
	if len(buf) == len(sidecarMagic)+pieces+4 && string(buf[:len(sidecarMagic)]) == sidecarMagic {
		body := buf[:len(buf)-4]
		if crc32.ChecksumIEEE(body) != binary.BigEndian.Uint32(buf[len(buf)-4:]) {
			return nil, false
		}
		bits = body[len(sidecarMagic):]
	}
	if len(bits) != pieces {
		return nil, false
	}
	return bits, true
}

// SidecarProgress returns the size of the verified data saved for a torrent that is not open
// (for example a paused one).
func SidecarProgress(base string, hash metainfo.Hash, pieceLen int64, pieces int, total int64) (int64, bool) {
	buf, err := os.ReadFile(sidecarPath(base, hash))
	if err != nil {
		return 0, false
	}
	bits, ok := parseSidecar(buf, pieces)
	if !ok {
		return 0, false
	}
	n := 0
	for _, b := range bits {
		if b == '1' {
			n++
		}
	}
	bytes := int64(n) * pieceLen
	if pieces > 0 && bits[pieces-1] == '1' {
		bytes -= int64(pieces)*pieceLen - total // the last piece is usually shorter
	}
	if bytes < 0 {
		bytes = 0
	}
	if bytes > total {
		bytes = total
	}
	return bytes, true
}

// completeCount returns the number of verified pieces.
func (ts *torrentStore) completeCount() int {
	ts.cmu.Lock()
	defer ts.cmu.Unlock()
	n := 0
	for _, st := range ts.state {
		if st == pieceComplete {
			n++
		}
	}
	return n
}

// verifiedBytes is the size of the verified pieces, i.e. the data that survives a restart.
func (ts *torrentStore) verifiedBytes() int64 {
	ts.cmu.Lock()
	defer ts.cmu.Unlock()
	n := 0
	for _, st := range ts.state {
		if st == pieceComplete {
			n++
		}
	}
	b := int64(n) * ts.pieceLen
	if last := len(ts.state) - 1; last >= 0 && ts.state[last] == pieceComplete {
		b -= int64(len(ts.state))*ts.pieceLen - ts.total // the last piece is usually shorter
	}
	if b < 0 {
		b = 0
	}
	if b > ts.total {
		b = ts.total
	}
	return b
}

// writeSidecar overwrites the saved state in place and fsyncs it. A rename of a temporary file is
// deliberately not used: on Windows it fails whenever a virus scanner or indexer holds the file,
// which left the saved state stale for a long time. Retries cover short sharing violations; the CRC
// protects against a torn write.
func writeSidecar(path string, buf []byte) (err error) {
	for attempt := 0; attempt < 5; attempt++ {
		if err = writeFileSync(path, buf); err == nil {
			return nil
		}
		time.Sleep(time.Duration(attempt+1) * 50 * time.Millisecond)
	}
	return err
}

func writeFileSync(path string, buf []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	_, err = f.Write(buf)
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}

// markFileDirty remembers that the file was written to and must be fsynced.
func (ts *torrentStore) markFileDirty(i int) {
	ts.smu.Lock()
	if ts.dirtyFiles == nil {
		ts.dirtyFiles = make(map[int]struct{})
	}
	ts.dirtyFiles[i] = struct{}{}
	ts.smu.Unlock()
}

// syncFile fsyncs one file through its own descriptor: fsync works per file,
// not per descriptor. A file that does not exist (any more) has nothing to flush.
func syncFile(path string) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	err = f.Sync()
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}

// syncDirty fsyncs the files written to since the last call. Files that could
// not be synced stay dirty for the next attempt.
func (ts *torrentStore) syncDirty() error {
	ts.smu.Lock()
	idx := make([]int, 0, len(ts.dirtyFiles))
	for i := range ts.dirtyFiles {
		idx = append(idx, i)
	}
	ts.dirtyFiles = nil
	ts.smu.Unlock()

	var firstErr error
	for _, i := range idx {
		if err := syncFile(ts.files[i].Disk); err != nil {
			ts.markFileDirty(i)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// saveCompletion saves the piece state. The state must never get ahead of the
// data: the files written to are fsynced first, and the state is saved only if
// that worked (otherwise it stays dirty and is retried). Without the fsync a
// power loss could leave a state that vouches for pieces that never reached the disk.
func (ts *torrentStore) saveCompletion() error {
	ts.cmu.Lock()
	if !ts.dirty {
		ts.cmu.Unlock()
		return nil
	}
	buf := make([]byte, len(sidecarMagic)+len(ts.state)+4)
	copy(buf, sidecarMagic)
	for i, st := range ts.state {
		if st == pieceComplete {
			buf[len(sidecarMagic)+i] = '1'
		} else {
			buf[len(sidecarMagic)+i] = '0'
		}
	}
	binary.BigEndian.PutUint32(buf[len(buf)-4:], crc32.ChecksumIEEE(buf[:len(buf)-4]))
	ts.dirty = false
	ts.cmu.Unlock()

	// every piece in buf was written before it was marked complete, so every
	// file that has data of these pieces is already in the dirty set
	err := ts.syncDirty()
	if err == nil {
		err = writeSidecar(ts.sidecar, buf)
	}
	if err != nil {
		ts.cmu.Lock()
		ts.dirty = true // try again on the next tick
		quiet := time.Since(ts.lastErrLog) < 30*time.Second
		if !quiet {
			ts.lastErrLog = time.Now()
		}
		ts.cmu.Unlock()
		if !quiet {
			log.TLogln("offline: cannot save the piece state of", ts.hash.HexString(), err)
		}
		return err
	}
	return nil
}

func (ts *torrentStore) flushLoop() {
	t := time.NewTicker(flushInterval)
	defer t.Stop()
	for {
		select {
		case <-ts.stop:
			return
		case <-t.C:
			ts.saveCompletion()
		}
	}
}

// Close implements storage.TorrentImpl. It is safe to call more than once.
func (ts *torrentStore) Close() error {
	ts.closeOnce.Do(func() {
		ts.stopOnce.Do(func() { close(ts.stop) })
		ts.saveCompletion()
		ts.saveCompletion() // a failed save leaves the state dirty: one more try
		log.TLogln("offline: storage closed", ts.hash.HexString(),
			"- verified pieces:", ts.completeCount(), "of", len(ts.state))

		ts.hmu.Lock()
		ts.closed = true
		for k, h := range ts.handles {
			_ = h.f.Close()
			delete(ts.handles, k)
		}
		ts.lru.Init()
		ts.hmu.Unlock()

		ts.owner.forget(ts)
	})
	return nil
}

func (ts *torrentStore) Piece(p metainfo.Piece) storage.PieceImpl {
	return &pieceImpl{ts: ts, idx: p.Index(), off: p.Offset()}
}

// fileIndex returns the file that contains the byte at pos, or -1.
func (ts *torrentStore) fileIndex(pos int64) int {
	i := sort.Search(len(ts.files), func(i int) bool {
		return ts.files[i].Offset+ts.files[i].Length > pos
	})
	if i >= len(ts.files) {
		return -1
	}
	return i
}

// openFile opens the file of file index i; hmu must be held (it is only called
// from acquire, which holds it).
func (ts *torrentStore) openFile(i int, write bool) (*os.File, error) {
	path := ts.files[i].Disk
	if write {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		return os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	}
	return os.Open(path)
}

// acquire returns the handle of the file i with its refs incremented. When a
// write is requested and only a read-only handle is cached, it is upgraded
// (possible only while nobody uses it) or a separate descriptor is opened for
// this one write - both are positioned I/O, so that is safe. ts.hmu must not
// be held.
func (ts *torrentStore) acquire(i int, write bool) (*fileHandle, error) {
	ts.hmu.Lock()
	defer ts.hmu.Unlock()
	if ts.closed {
		return nil, errClosed
	}
	if h := ts.handles[i]; h != nil {
		if h.rw || !write {
			h.refs++
			ts.lru.MoveToFront(h.elem)
			return h, nil
		}
		if h.refs == 0 {
			// upgrade the cached read-only handle to a writable one
			_ = h.f.Close()
			ts.lru.Remove(h.elem)
			delete(ts.handles, i)
		} else {
			// the read-only handle is pinned by concurrent readers: open a
			// separate descriptor for this write and do not cache it
			f, err := ts.openFile(i, true)
			if err != nil {
				return nil, err
			}
			return &fileHandle{f: f, rw: true, refs: 1}, nil
		}
	}
	if len(ts.handles) >= maxOpenFiles {
		ts.evictLocked()
	}
	f, err := ts.openFile(i, write)
	if err != nil {
		return nil, err
	}
	h := &fileHandle{f: f, rw: write, refs: 1}
	h.elem = ts.lru.PushFront(i)
	ts.handles[i] = h
	return h, nil
}

// release marks the handle as no longer used by the caller.
func (ts *torrentStore) release(h *fileHandle) {
	ts.hmu.Lock()
	defer ts.hmu.Unlock()
	h.refs--
	if h.elem == nil {
		// A non-cached handle (the separate descriptor opened for a write while
		// the cached read-only one was pinned) is known to nobody but its user:
		// neither the eviction nor Close can see it, so the last release must
		// close it. Without this every such write leaked a file descriptor.
		if h.refs <= 0 {
			_ = h.f.Close()
		}
		return
	}
	if ts.closed {
		// Close already closed every cached handle: only the bookkeeping is left
		return
	}
	ts.lru.MoveToFront(h.elem)
}

// evictLocked closes the least recently used unpinned handles until there is
// room for one more. Pinned handles (refs > 0) are skipped: a handle that is
// being read from must not be closed underneath its user. The open file limit
// is therefore soft, bounded by maxOpenFiles plus the number of concurrent
// users. ts.hmu must be held.
func (ts *torrentStore) evictLocked() {
	for len(ts.handles) >= maxOpenFiles {
		var victim *list.Element
		for v := ts.lru.Back(); v != nil; v = v.Prev() {
			if h := ts.handles[v.Value.(int)]; h != nil && h.refs == 0 {
				victim = v
				break
			}
		}
		if victim == nil {
			return // everything is pinned: exceed the limit until they are released
		}
		i := victim.Value.(int)
		_ = ts.handles[i].f.Close()
		ts.lru.Remove(victim)
		delete(ts.handles, i)
	}
}

func (ts *torrentStore) readAt(b []byte, global int64) (n int, err error) {
	for n < len(b) {
		pos := global + int64(n)
		i := ts.fileIndex(pos)
		if i < 0 {
			return n, io.EOF
		}
		f := ts.files[i]
		inFile := pos - f.Offset
		want := int64(len(b) - n)
		if rem := f.Length - inFile; want > rem {
			want = rem
		}
		h, e := ts.acquire(i, false)
		if e != nil {
			if errors.Is(e, os.ErrNotExist) || errors.Is(e, errClosed) {
				// a missing file simply means "no data yet"
				return n, io.EOF
			}
			// anything else (too many open files, permissions...) is a real
			// error and must not be mistaken for missing data
			return n, e
		}
		m, e := h.f.ReadAt(b[n:n+int(want)], inFile)
		ts.release(h)
		n += m
		if e != nil && int64(m) < want {
			return n, e
		}
	}
	return n, nil
}

func (ts *torrentStore) writeAt(b []byte, global int64) (n int, err error) {
	for n < len(b) {
		pos := global + int64(n)
		i := ts.fileIndex(pos)
		if i < 0 {
			return n, io.ErrShortWrite
		}
		f := ts.files[i]
		inFile := pos - f.Offset
		want := int64(len(b) - n)
		if rem := f.Length - inFile; want > rem {
			want = rem
		}
		h, e := ts.acquire(i, true)
		if e != nil {
			return n, e
		}
		m, e := h.f.WriteAt(b[n:n+int(want)], inFile)
		ts.release(h)
		n += m
		if m > 0 {
			ts.markFileDirty(i)
		}
		if e != nil {
			return n, e
		}
	}
	return n, nil
}

type pieceImpl struct {
	ts  *torrentStore
	idx int
	off int64 // offset of the piece inside the torrent data stream
}

func (p *pieceImpl) ReadAt(b []byte, off int64) (int, error)  { return p.ts.readAt(b, p.off+off) }
func (p *pieceImpl) WriteAt(b []byte, off int64) (int, error) { return p.ts.writeAt(b, p.off+off) }

func (p *pieceImpl) setState(st uint8) {
	p.ts.cmu.Lock()
	if p.idx >= 0 && p.idx < len(p.ts.state) && p.ts.state[p.idx] != st {
		p.ts.state[p.idx] = st
		p.ts.dirty = true
	}
	p.ts.cmu.Unlock()
}

func (p *pieceImpl) MarkComplete() error    { p.setState(pieceComplete); return nil }
func (p *pieceImpl) MarkNotComplete() error { p.setState(pieceNotComplete); return nil }

func (p *pieceImpl) Completion() storage.Completion {
	p.ts.cmu.Lock()
	defer p.ts.cmu.Unlock()
	if p.idx < 0 || p.idx >= len(p.ts.state) {
		return storage.Completion{}
	}
	switch p.ts.state[p.idx] {
	case pieceComplete:
		return storage.Completion{Complete: true, Ok: true}
	case pieceNotComplete:
		return storage.Completion{Complete: false, Ok: true}
	}
	return storage.Completion{}
}
