package torrstor

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/anacrolix/torrent/metainfo"

	"server/log"
	"server/settings"
	"server/torr/utils"
)

type diskPieceRef struct {
	hash     string
	pieceId  int
	size     int64
	accessed int64
}

var (
	cleanGlobalMu sync.Mutex

	// diskUsed is the size of the disk cache in bytes, kept in memory because it is
	// read on every completed piece and every priority update, and the cache may hold
	// tens of thousands of piece files. Pieces adjust it as they are written and
	// released; a full scan resyncs it at most once per diskUsageResync to absorb
	// files removed behind the cache's back (RemTorrent, manual cleanup).
	diskUsed     atomic.Int64
	diskUsedSync atomic.Int64 // unix nanos of the last full scan, 0 means never
)

const (
	diskUsageResync = time.Minute
	// cleanGlobal frees this share of the quota below the limit, so that the next
	// few downloaded pieces do not each trigger another scan of the cache directory
	diskCacheHeadroomDiv = 50
)

// IsLongTermCache reports whether cached pieces are kept on disk until the global
// quota DiskCacheSize is full.
func IsLongTermCache() bool {
	return settings.BTsets != nil && settings.BTsets.LongTermCache && settings.BTsets.UseDisk
}

func diskCacheUsed() int64 {
	last := diskUsedSync.Load()
	if last == 0 || time.Since(time.Unix(0, last)) > diskUsageResync {
		if diskUsedSync.CompareAndSwap(last, time.Now().UnixNano()) {
			diskUsed.Store(diskCacheUsage())
		}
	}
	return diskUsed.Load()
}

func addDiskCacheUsed(delta int64) {
	if delta != 0 {
		diskUsed.Add(delta)
	}
}

func (s *Storage) cleanGlobal() {
	if !IsLongTermCache() || settings.BTsets.DiskCacheSize <= 0 {
		return
	}
	quota := settings.BTsets.DiskCacheSize
	if diskCacheUsed() <= quota {
		return
	}

	if !cleanGlobalMu.TryLock() {
		return
	}
	defer cleanGlobalMu.Unlock()

	candidates := scanDiskCachePieces()
	var total int64
	for _, ref := range candidates {
		total += ref.size
	}
	diskUsed.Store(total)
	diskUsedSync.Store(time.Now().UnixNano())
	if total <= quota {
		return
	}

	toRemove := total - (quota - quota/diskCacheHeadroomDiv)
	protected, partial := s.activeDiskPieces()

	evictable := make([]diskPieceRef, 0, len(candidates))
	for _, ref := range candidates {
		if protected[ref.hash][ref.pieceId] {
			continue
		}
		if ref.size > 0 {
			evictable = append(evictable, ref)
		}
	}

	// Oldest first; a piece an open torrent is still downloading goes last. The torrent
	// client never re-requests chunks it has received, so evicting their bytes makes a
	// reader that comes back to them read zeros.
	sort.Slice(evictable, func(i, j int) bool {
		pi, pj := partial[evictable[i].hash][evictable[i].pieceId], partial[evictable[j].hash][evictable[j].pieceId]
		if pi != pj {
			return pj
		}
		return evictable[i].accessed < evictable[j].accessed
	})

	// eviction only frees space: the torrent card, its metainfo and its pins stay
	// until the user removes the torrent (RemTorrent deletes the whole directory)
	var removed int64
	for _, ref := range evictable {
		if removed >= toRemove {
			break
		}
		if s.evictDiskPiece(ref) {
			removed += ref.size
		}
	}

	if removed > 0 {
		utils.FreeOSMemGC()
	}
}

func diskCacheUsage() int64 {
	var total int64
	for _, ref := range scanDiskCachePieces() {
		total += ref.size
	}
	return total
}

// scanDiskCachePieces lists the cached piece files of every torrent.
func scanDiskCachePieces() []diskPieceRef {
	path := settings.BTsets.TorrentsSavePath
	if path == "" || path == "/" {
		return nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}
	refs := make([]diskPieceRef, 0)
	for _, entry := range entries {
		if !entry.IsDir() || !isInfoHashName(entry.Name()) {
			continue
		}
		refs = appendHashDirPieces(refs, path, entry.Name())
	}
	return refs
}

// appendHashDirPieces appends the cached piece files of one torrent: files named by
// a piece index (not the info and pins files) that hold data.
func appendHashDirPieces(refs []diskPieceRef, path, hashHex string) []diskPieceRef {
	files, err := os.ReadDir(filepath.Join(path, hashHex))
	if err != nil {
		return refs
	}
	for _, f := range files {
		if f.IsDir() || isCacheMetaFile(f.Name()) {
			continue
		}
		pieceId, err := strconv.Atoi(f.Name())
		if err != nil || pieceId < 0 {
			continue
		}
		info, err := f.Info()
		if err != nil || info.Size() <= 0 {
			continue
		}
		refs = append(refs, diskPieceRef{
			hash:     hashHex,
			pieceId:  pieceId,
			size:     info.Size(),
			accessed: info.ModTime().Unix(),
		})
	}
	return refs
}

// activeDiskPieces reports, for the torrents open in this storage, the pieces that
// global eviction must keep (see Cache.getProtectedPieceIds) and the pieces that are
// only partially downloaded.
func (s *Storage) activeDiskPieces() (protected, partial map[string]map[int]bool) {
	protected = make(map[string]map[int]bool)
	partial = make(map[string]map[int]bool)

	s.mu.Lock()
	caches := make([]*Cache, 0, len(s.caches))
	for _, cache := range s.caches {
		if cache != nil && !cache.isClosed.Load() {
			caches = append(caches, cache)
		}
	}
	s.mu.Unlock()

	for _, cache := range caches {
		hashHex := cache.hash.HexString()
		protected[hashHex] = cache.getProtectedPieceIds()
		part := make(map[int]bool)
		for id, p := range cache.getPieces() {
			if p.Size > 0 && !p.Complete {
				part[id] = true
			}
		}
		partial[hashHex] = part
	}
	return protected, partial
}

func (s *Storage) evictDiskPiece(ref diskPieceRef) bool {
	cache := s.GetCache(metainfo.NewHashFromHex(ref.hash))
	if cache != nil && !cache.isClosed.Load() {
		// Release goes through the torrent so the client learns the piece is gone;
		// DiskPiece.Release also updates diskUsed
		if piece, ok := cache.getPieces()[ref.pieceId]; ok && piece.Size > 0 {
			cache.removePiece(piece)
			return true
		}
	}

	piecePath := filepath.Join(settings.BTsets.TorrentsSavePath, ref.hash, strconv.Itoa(ref.pieceId))
	if err := os.Remove(piecePath); err != nil {
		if !os.IsNotExist(err) {
			log.TLogln("Error remove cache piece:", err)
		}
		return false
	}
	addDiskCacheUsed(-ref.size)
	return true
}

// CachedPieceSizes returns the size of every cached piece file of a torrent.
func CachedPieceSizes(hashHex string) map[int]int64 {
	sizes := make(map[int]int64)
	path := settings.BTsets.TorrentsSavePath
	if path == "" || path == "/" {
		return sizes
	}
	for _, ref := range appendHashDirPieces(nil, path, hashHex) {
		sizes[ref.pieceId] = ref.size
	}
	return sizes
}
