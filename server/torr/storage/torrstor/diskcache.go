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

func isLongTermCache() bool {
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
	if !isLongTermCache() || settings.BTsets.DiskCacheSize <= 0 {
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

	var removed int64
	affected := make(map[string]bool)
	for _, ref := range evictable {
		if removed >= toRemove {
			break
		}
		if s.evictDiskPiece(ref) {
			removed += ref.size
			affected[ref.hash] = true
		}
	}

	for hashHex := range affected {
		s.onTorrentCacheChanged(hashHex)
	}

	if removed > 0 {
		utils.FreeOSMemGC()
	}
}

func diskCacheUsage() int64 {
	path := settings.BTsets.TorrentsSavePath
	if path == "" || path == "/" {
		return 0
	}

	var total int64
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}

	for _, entry := range entries {
		if !entry.IsDir() || len(entry.Name()) != 40 {
			continue
		}
		hashDir := filepath.Join(path, entry.Name())
		files, err := os.ReadDir(hashDir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() || isCacheMetaFile(f.Name()) {
				continue
			}
			info, err := f.Info()
			if err != nil {
				continue
			}
			total += info.Size()
		}
	}
	return total
}

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
		if !entry.IsDir() || len(entry.Name()) != 40 {
			continue
		}
		hashDir := filepath.Join(path, entry.Name())
		files, err := os.ReadDir(hashDir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() || isCacheMetaFile(f.Name()) {
				continue
			}
			pieceId, err := strconv.Atoi(f.Name())
			if err != nil {
				continue
			}
			info, err := f.Info()
			if err != nil || info.Size() == 0 {
				continue
			}
			refs = append(refs, diskPieceRef{
				hash:     entry.Name(),
				pieceId:  pieceId,
				size:     info.Size(),
				accessed: info.ModTime().Unix(),
			})
		}
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

func (s *Storage) onTorrentCacheChanged(hashHex string) {
	if !hasDiskCacheForHash(hashHex) {
		settings.RemTorrent(metainfo.NewHashFromHex(hashHex))
		hashDir := filepath.Join(settings.BTsets.TorrentsSavePath, hashHex)
		_ = os.Remove(hashDir)
		log.TLogln("Remove torrent card after cache eviction:", hashHex)
	}
}

func hasDiskCacheForHash(hashHex string) bool {
	hashDir := filepath.Join(settings.BTsets.TorrentsSavePath, hashHex)
	entries, err := os.ReadDir(hashDir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if isCacheMetaFile(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err == nil && info.Size() > 0 {
			return true
		}
	}
	return false
}

// CachedPieceSizes returns the size of every cached piece file of a torrent.
func CachedPieceSizes(hashHex string) map[int]int64 {
	sizes := make(map[int]int64)
	entries, err := os.ReadDir(filepath.Join(settings.BTsets.TorrentsSavePath, hashHex))
	if err != nil {
		return sizes
	}
	for _, entry := range entries {
		if entry.IsDir() || isCacheMetaFile(entry.Name()) {
			continue
		}
		id, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}
		if info, err := entry.Info(); err == nil && info.Size() > 0 {
			sizes[id] = info.Size()
		}
	}
	return sizes
}

// RemoveDiskCache deletes the cached pieces and the unfinished download of a
// torrent that is no longer open, keeping its metainfo.
func RemoveDiskCache(hashHex string) {
	path := settings.BTsets.TorrentsSavePath
	if path == "" || path == "/" || len(hashHex) != 40 {
		return
	}
	dir := filepath.Join(path, hashHex)
	var removed int64
	for id, size := range CachedPieceSizes(hashHex) {
		if err := os.Remove(filepath.Join(dir, strconv.Itoa(id))); err == nil {
			removed += size
		}
	}
	_ = os.Remove(pinsPath(hashHex))
	addDiskCacheUsed(-removed)
	if removed > 0 {
		log.TLogln("Remove cache of sleeping torrent:", hashHex, removed, "bytes")
	}
}
