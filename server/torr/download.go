package torr

import (
	"errors"
	"fmt"

	"github.com/anacrolix/torrent/metainfo"

	"server/log"
	"server/settings"
	"server/torr/state"
	"server/torr/storage/torrstor"
)

// Downloading a file or a whole torrent pins its pieces in the long-term disk cache
// (see torrstor/pin.go). A torrent that has at least seedShare of one of its files
// cached stays open to seed; one with less closes on timeout as usual, and its cached
// pieces stay on disk until the global eviction needs the space.

const (
	seedShare = 0.5
	// files smaller than this (subtitles, nfo, covers) are cached in full by the
	// pieces they share with the main file, so they say nothing about what to seed
	seedMinFileSize = 50 << 20
)

var (
	ErrDownloadNeedsLongTermCache = errors.New("download to cache requires the long-term disk cache")
	ErrDownloadNoInfo             = errors.New("torrent info is not loaded")
	ErrDownloadFileNotFound       = errors.New("file not found in torrent")
)

// Download starts downloading the file with the given status id (FileStats[].Id)
// in full, or the whole torrent when index <= 0.
func (t *Torrent) Download(index int) error {
	if !torrstor.IsLongTermCache() {
		return ErrDownloadNeedsLongTermCache
	}
	cache := t.GetCache()
	if cache == nil || t.Torrent == nil || t.Torrent.Info() == nil {
		return ErrDownloadNoInfo
	}
	r, size, err := t.downloadRange(index)
	if err != nil {
		return err
	}
	if quota := settings.BTsets.DiskCacheSize; quota > 0 && size > quota {
		return fmt.Errorf("size %d exceeds the disk cache quota %d", size, quota)
	}
	log.TLogln("Download to cache:", t.Hash().HexString(), "index", index)
	cache.PinPieces(r)
	return nil
}

// CancelDownload stops a download started by Download with the same index.
func (t *Torrent) CancelDownload(index int) error {
	cache := t.GetCache()
	if cache == nil || t.Torrent == nil || t.Torrent.Info() == nil {
		return ErrDownloadNoInfo
	}
	r, _, err := t.downloadRange(index)
	if err != nil {
		return err
	}
	cache.UnpinPieces(r)
	return nil
}

func (t *Torrent) downloadRange(index int) (torrstor.PieceRange, int64, error) {
	torrent := t.Torrent
	if index <= 0 {
		return torrstor.PieceRange{Begin: 0, End: torrent.NumPieces()}, torrent.Length(), nil
	}
	file := t.findFileIndex(torrent, index)
	if file == nil {
		return torrstor.PieceRange{}, 0, ErrDownloadFileNotFound
	}
	r := fileRange(torrent.Info().PieceLength, file.Offset(), file.Length())
	return r, file.Length(), nil
}

func fileRange(pieceLength, offset, length int64) torrstor.PieceRange {
	if pieceLength <= 0 || length <= 0 {
		return torrstor.PieceRange{}
	}
	return torrstor.PieceRange{
		Begin: int(offset / pieceLength),
		End:   int((offset + length + pieceLength - 1) / pieceLength),
	}
}

// fileCompletedBytes returns the bytes of a file covered by complete pieces.
func fileCompletedBytes(pieceLength, offset, length int64, complete func(int) bool) int64 {
	r := fileRange(pieceLength, offset, length)
	var done int64
	for id := r.Begin; id < r.End; id++ {
		if !complete(id) {
			continue
		}
		begin := max(int64(id)*pieceLength, offset)
		end := min(int64(id+1)*pieceLength, offset+length)
		if end > begin {
			done += end - begin
		}
	}
	return done
}

func seedableShare(length, completed int64) bool {
	return length >= seedMinFileSize && float64(completed) >= float64(length)*seedShare
}

// pieceCompletion snapshots which pieces the client has.
func (t *Torrent) pieceCompletion() []bool {
	complete := make([]bool, 0, t.Torrent.NumPieces())
	for _, run := range t.Torrent.PieceStateRuns() {
		for i := 0; i < run.Length; i++ {
			complete = append(complete, run.Complete)
		}
	}
	return complete
}

// fillDownloadStatus adds per-file cache progress and the download state to st.
// The caller holds muTorrent and has filled st.FileStats.
func (t *Torrent) fillDownloadStatus(st *state.TorrentStatus) {
	info := t.Torrent.Info()
	if info == nil || !torrstor.IsLongTermCache() {
		return
	}
	complete := t.pieceCompletion()
	isComplete := func(id int) bool { return id < len(complete) && complete[id] }
	offsets := make(map[string]int64, len(st.FileStats))
	for _, f := range t.Files() {
		offsets[f.Path()] = f.Offset()
	}
	for _, fs := range st.FileStats {
		offset, ok := offsets[fs.Path]
		if !ok {
			continue
		}
		fs.Loaded = fileCompletedBytes(info.PieceLength, offset, fs.Length, isComplete)
		if t.cache != nil {
			fs.Downloading = t.cache.PinnedPieces(fileRange(info.PieceLength, offset, fs.Length)) > 0
		}
	}
	if t.cache != nil {
		st.Downloading = t.cache.HasPinned()
	}
}

// seedable reports whether enough of one file is cached to keep seeding it.
func (t *Torrent) seedable() bool {
	if !torrstor.IsLongTermCache() || t.Torrent == nil || t.Torrent.Info() == nil {
		return false
	}
	info := t.Torrent.Info()
	complete := t.pieceCompletion()
	isComplete := func(id int) bool { return id < len(complete) && complete[id] }
	for _, f := range t.Files() {
		if seedableShare(f.Length(), fileCompletedBytes(info.PieceLength, f.Offset(), f.Length(), isComplete)) {
			return true
		}
	}
	return false
}

// seedableOnDisk is seedable for a torrent that isn't loaded, judged by its
// metainfo and the piece files in the cache directory.
func seedableOnDisk(hashHex string, infoBytes []byte) bool {
	if len(infoBytes) == 0 {
		infoBytes = torrstor.LoadMetainfo(hashHex)
	}
	if len(infoBytes) == 0 {
		return false
	}
	mi := metainfo.MetaInfo{InfoBytes: infoBytes}
	info, err := mi.UnmarshalInfo()
	if err != nil || info.PieceLength <= 0 {
		return false
	}
	sizes := torrstor.CachedPieceSizes(hashHex)
	if len(sizes) == 0 {
		return false
	}
	total := info.TotalLength()
	isComplete := func(id int) bool {
		want := min(info.PieceLength, total-int64(id)*info.PieceLength)
		return want > 0 && sizes[id] == want
	}
	var offset int64
	for _, f := range info.UpvertedFiles() {
		if seedableShare(f.Length, fileCompletedBytes(info.PieceLength, offset, f.Length, isComplete)) {
			return true
		}
		offset += f.Length
	}
	return false
}

// ResumeCachedTorrents opens the torrents that have an unfinished download or
// enough cached data to seed, so they keep downloading and seeding after a restart.
func ResumeCachedTorrents() {
	if !torrstor.IsLongTermCache() {
		return
	}
	for _, tor := range ListTorrentsDB() {
		hashHex := tor.Hash().HexString()
		var infoBytes []byte
		if tor.TorrentSpec != nil {
			infoBytes = tor.TorrentSpec.InfoBytes
		}
		if torrstor.HasPins(hashHex) || seedableOnDisk(hashHex, infoBytes) {
			log.TLogln("Resume cached torrent:", hashHex)
			startLoadFromDB(tor)
		}
	}
}
