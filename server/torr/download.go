package torr

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"

	"server/log"
	"server/settings"
	"server/torr/state"
	"server/torr/storage/torrstor"
	utils2 "server/utils"
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
	t.keepInDB()
	return nil
}

// keepInDB saves a torrent that has data in the long-term cache, so that its card
// stays with its cache. A torrent added without saving (a stream or play link from
// an external player) lives only in memory, and a settings change or a restart
// closes every torrent and reopens the DB ones only: without this its card would
// vanish while its cached pieces stay on disk. Called once a torrent has written
// its first piece (progressEvent) and when a download to the cache starts.
func (t *Torrent) keepInDB() {
	if settings.ReadOnly || t.Torrent == nil || t.inDB.Load() {
		return
	}
	if GetTorrentDB(t.Hash()) == nil {
		if t.Title == "" {
			t.Title = t.Torrent.Name()
		}
		SaveTorrentToDB(t)
	}
	t.inDB.Store(true)
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
	dbTorrs := ListTorrentsDB()
	adoptCachedTorrents(dbTorrs)
	for _, tor := range dbTorrs {
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

// adoptCachedTorrents adds to the DB the torrents that have cached pieces or an
// unfinished download but no DB entry (cached by a build that did not keep them,
// see keepInDB), so their cards come back. The new entries are added to dbTorrs.
func adoptCachedTorrents(dbTorrs map[metainfo.Hash]*Torrent) {
	if settings.ReadOnly {
		return
	}
	for _, hashHex := range torrstor.CachedTorrents() {
		hash := metainfo.NewHashFromHex(hashHex)
		if _, ok := dbTorrs[hash]; ok {
			continue
		}
		if len(torrstor.CachedPieceSizes(hashHex)) == 0 && !torrstor.HasPins(hashHex) {
			continue
		}
		tor := cachedTorrent(hash)
		if tor == nil {
			continue
		}
		log.TLogln("Add cached torrent to DB:", hashHex)
		AddTorrentDB(tor)
		dbTorrs[hash] = tor
	}
}

// cachedTorrent builds a DB torrent from the metainfo saved in its cache directory.
func cachedTorrent(hash metainfo.Hash) *Torrent {
	infoBytes := torrstor.LoadMetainfo(hash.HexString())
	if len(infoBytes) == 0 {
		return nil
	}
	mi := metainfo.MetaInfo{InfoBytes: infoBytes}
	info, err := mi.UnmarshalInfo()
	if err != nil {
		return nil
	}
	tor := &Torrent{
		TorrentSpec: &torrent.TorrentSpec{InfoHash: hash, InfoBytes: infoBytes, DisplayName: info.BestName()},
		Title:       info.BestName(),
		Size:        info.TotalLength(),
		Stat:        state.TorrentInDB,
		Timestamp:   time.Now().Unix(),
	}
	files := new(tsFiles)
	files.TorrServer.Files = fileStatsFromInfo(&info)
	if buf, err := json.Marshal(files); err == nil {
		tor.Data = string(buf)
	}
	return tor
}

// fileStatsFromInfo lists the files of a torrent as Status does for a loaded one:
// the same paths, the same order, and so the same ids.
func fileStatsFromInfo(info *metainfo.Info) []*state.TorrentFileStat {
	files := make([]*state.TorrentFileStat, 0, len(info.UpvertedFiles()))
	for _, fi := range info.UpvertedFiles() {
		files = append(files, &state.TorrentFileStat{
			Path:   strings.Join(append([]string{info.BestName()}, fi.BestPath()...), "/"),
			Length: fi.Length,
		})
	}
	sort.Slice(files, func(i, j int) bool {
		return utils2.CompareStrings(files[i].Path, files[j].Path)
	})
	for i, f := range files {
		f.Id = i + 1
	}
	return files
}
