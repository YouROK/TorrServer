package torr

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anacrolix/torrent"

	"server/settings"
	"server/torr/offlinetest"
	"server/torr/state"
)

func connectShow(t *testing.T) (*Torrent, *offlinetest.Seeder) {
	t.Helper()
	return connectSeeded(t, offlinetest.WriteShow(t))
}

func connectSeeded(t *testing.T, show *offlinetest.Show) (*Torrent, *offlinetest.Seeder) {
	t.Helper()
	offlinetest.Prepare(t)
	return connectPrepared(t, show)
}

// connectPrepared assumes offlinetest.Prepare already ran so the caller can
// tweak BTsets (e.g. CacheSize) before Connect snapshots storage capacity.
func connectPrepared(t *testing.T, show *offlinetest.Show) (*Torrent, *offlinetest.Seeder) {
	t.Helper()
	bt := NewBTS()
	InitApiHelper(bt)
	if err := bt.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(bt.Disconnect)

	seed := offlinetest.Seed(t, show)
	tor, err := AddTorrent(show.Spec, "Show", "", "", "tv")
	if err != nil {
		t.Fatalf("AddTorrent: %v", err)
	}
	if !tor.GotInfo() {
		t.Fatal("GotInfo")
	}
	p := seed.Peer()
	if p.Port == 0 {
		t.Fatal("seeder has no listen port")
	}
	tor.AddPeers([]torrent.Peer{p})
	return tor, seed
}

func waitBytes(t *testing.T, tor *Torrent, n int) []byte {
	t.Helper()
	var file *torrent.File
	for _, f := range tor.Files() {
		if strings.HasSuffix(f.Path(), "episode1.mkv") {
			file = f
			break
		}
	}
	if file == nil {
		t.Fatal("no mkv")
	}
	r := tor.NewReader(file)
	if r == nil {
		t.Fatal("NewReader")
	}
	defer tor.CloseReader(r)
	r.SetResponsive()
	buf := make([]byte, n)
	done := make(chan error, 1)
	go func() {
		_, err := io.ReadFull(r, buf)
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("read from local seeder timed out")
	}
	return buf
}

func TestOfflineSeedReadAndLifecycle(t *testing.T) {
	tor, _ := connectShow(t)
	got := waitBytes(t, tor, 64)
	if string(got) != strings.Repeat("M", 64) {
		t.Fatalf("payload %q", got[:8])
	}

	if tor.Hash().HexString() == "" || tor.Length() <= 0 {
		t.Fatalf("hash/len %s %d", tor.Hash(), tor.Length())
	}
	st := tor.Status()
	if st.Stat != state.TorrentWorking || len(st.FileStats) != 4 || st.StatString == "" {
		t.Fatalf("status %+v", st)
	}
	if tor.CacheState() == nil {
		t.Fatal("cache state")
	}
	rows, _, _ := SnapshotCaches()
	if len(rows) == 0 {
		t.Fatal("cache snapshot")
	}
	snap := SnapshotClientStatus()
	if snap.TorrentCount == 0 {
		t.Fatal("client snapshot")
	}

	hash := tor.Hash().HexString()
	if GetTorrent(hash) == nil {
		t.Fatal("GetTorrent")
	}
	if len(ListTorrent()) == 0 {
		t.Fatal("ListTorrent")
	}
	SaveTorrentToDB(tor)
	if GetTorrentDB(tor.Hash()) == nil {
		t.Fatal("db")
	}

	SetTorrent(hash, "Renamed", "", "movie", `{"k":1}`)
	if tor.Title != "Renamed" || tor.Category != "movie" {
		t.Fatalf("set %s %s", tor.Title, tor.Category)
	}

	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	req.RemoteAddr = "127.0.0.1:9"
	req.Header.Set("Range", "bytes=0-1023")
	req.Header.Set("getContentFeatures.dlna.org", "1")
	w := httptest.NewRecorder()
	if err := tor.Stream(st.FileStats[0].Id, req, w); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusPartialContent && w.Code != http.StatusOK {
		t.Fatalf("stream %d %s", w.Code, w.Body.Bytes())
	}
	if w.Body.Len() == 0 {
		t.Fatal("empty stream")
	}
	if GetActiveStreams() < 0 {
		t.Fatal("streams")
	}

	// second file id that does not exist
	w = httptest.NewRecorder()
	if err := tor.Stream(99, req, w); err == nil {
		t.Fatal("expected missing file")
	}

	settings.MaxSize = 1
	w = httptest.NewRecorder()
	if err := tor.Stream(1, req, w); err == nil || w.Code != http.StatusForbidden {
		t.Fatalf("max size %v %d", err, w.Code)
	}
	settings.MaxSize = 0

	Preload(tor, 1)
	if tor.Status().Stat != state.TorrentWorking {
		t.Fatalf("after preload %v", tor.Status().Stat)
	}

	// cancel a preload that is still running by closing after start
	done := make(chan struct{})
	go func() {
		tor.Preload(1, 32<<20)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("preload hung")
	}

	DropTorrent(hash)
	if bts.GetTorrent(tor.Hash()) != nil {
		t.Fatal("still loaded after drop")
	}
	// GetTorrent reloads from the DB in the background. LoadTorrent on that same
	// row at the same time races with the reload, so load the row directly.
	row := GetTorrentDB(tor.Hash())
	if row == nil {
		t.Fatal("db row gone")
	}
	loaded := LoadTorrent(row)
	if loaded == nil || !loaded.GotInfo() {
		t.Fatal("LoadTorrent")
	}

	RemTorrent(hash)
	if GetTorrent(hash) != nil || GetTorrentDB(tor.Hash()) != nil {
		t.Fatal("removed torrent still listed")
	}
}

func TestOfflineReadersConcurrent(t *testing.T) {
	tor, _ := connectShow(t)
	file := tor.Files()[0]
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := tor.NewReader(file)
			if r == nil {
				t.Error("nil reader")
				return
			}
			defer tor.CloseReader(r)
			buf := make([]byte, 128)
			if _, err := r.Seek(0, io.SeekStart); err != nil {
				t.Error(err)
				return
			}
			if _, err := r.Read(buf); err != nil && err != io.EOF {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}

func TestPreloadReadsFirstAndLastPiece(t *testing.T) {
	tor, _ := connectShow(t)
	var id int
	var file *torrent.File
	for _, f := range tor.Status().FileStats {
		if strings.HasSuffix(f.Path, "episode1.mkv") {
			id = f.Id
			break
		}
	}
	for _, f := range tor.Files() {
		if strings.HasSuffix(f.Path(), "episode1.mkv") {
			file = f
			break
		}
	}
	if file == nil || id == 0 {
		t.Fatal("mkv")
	}
	tor.Preload(id, file.Length())
	cache := tor.GetCache()
	if cache == nil {
		t.Fatal("cache")
	}
	// A piece fully inside the file and inside the last 32KiB used to stay
	// empty: the preload loop stopped one buffer short of the end.
	pl := file.Torrent().Info().PieceLength
	fileOff := file.Offset()
	fileEnd := fileOff + file.Length()
	tailFrom := fileEnd - 32<<10
	headPiece, tailPiece := -1, -1
	for p := int(fileOff / pl); p <= int((fileEnd-1)/pl); p++ {
		p0 := int64(p) * pl
		p1 := p0 + pl
		if p0 < fileOff || p1 > fileEnd {
			continue
		}
		if headPiece < 0 && p1 <= fileOff+32<<10 {
			headPiece = p
		}
		if tailPiece < 0 && p0 >= tailFrom {
			tailPiece = p
		}
	}
	if headPiece < 0 || tailPiece < 0 {
		t.Fatalf("fixture has no whole piece in the head/tail, head=%d tail=%d", headPiece, tailPiece)
	}
	pieces := cache.GetState().Pieces
	for _, pid := range []int{headPiece, tailPiece} {
		it, ok := pieces[pid]
		if !ok || !it.Completed || it.Size < it.Length {
			t.Fatalf("piece %d after preload: ok=%v %+v", pid, ok, it)
		}
	}
}

// A file longer than the preload budget used to skip its tail: the head stopped
// one window early and the tail reader did not start, or it stopped 32KiB short.
// 48MiB file, 32MiB budget, 8MiB edge, 1MiB pieces: head is pieces 0..23,
// tail is pieces 40..47, piece 32 sits in the gap and must stay empty.
func TestPreloadReadsTailPastBudget(t *testing.T) {
	const (
		piece   int64 = 1 << 20
		fileLen int64 = 48 << 20
		budget  int64 = 32 << 20
	)
	show := offlinetest.WriteSized(t, "movie.mkv", fileLen, piece)
	offlinetest.Prepare(t)
	// Head+tail ≈ budget. offlinetest defaults CacheSize to budget, so
	// cleanPieces can drop mid-head (piece 23 is past the 8MiB keep window).
	// Preload uses anacrolix readers, which never pin Cache.readers ranges.
	settings.BTsets.CacheSize = budget + 16<<20
	tor, _ := connectPrepared(t, show)
	if len(tor.Files()) != 1 {
		t.Fatalf("files %d", len(tor.Files()))
	}
	tor.Preload(1, budget)
	cache := tor.GetCache()
	if cache == nil {
		t.Fatal("cache")
	}
	pieces := cache.GetState().Pieces
	for _, id := range []int{0, 23, 47} {
		it, ok := pieces[id]
		if !ok || !it.Completed || it.Size < it.Length {
			t.Fatalf("piece %d: ok=%v %+v", id, ok, it)
		}
	}
	if it, ok := pieces[32]; ok && it.Completed {
		t.Fatalf("gap piece 32 was loaded: %+v", it)
	}
}

func TestPreloadZeroAndClosed(t *testing.T) {
	tor, _ := connectShow(t)
	file := tor.Files()[0]
	tor.Preload(1, 0)
	tor.Close()
	tor.Preload(1, 1024)
	if tor.NewReader(file) != nil {
		t.Fatal("reader on closed torrent")
	}
}

func TestConfigureProxyModes(t *testing.T) {
	offlinetest.Prepare(t)
	bt := NewBTS()
	settings.BTsets.DownloadRateLimit = 64
	settings.BTsets.UploadRateLimit = 32
	settings.Args.ProxyURL = "://bad"
	bt.configure()
	settings.Args.ProxyURL = "ftp://127.0.0.1:1"
	if err := bt.configureProxy(); err == nil {
		t.Fatal("ftp proxy")
	}
	for _, mode := range []string{"", "tracker", "peers", "full"} {
		settings.Args.ProxyURL = "socks5://127.0.0.1:1080"
		settings.Args.ProxyMode = mode
		if err := bt.configureProxy(); err != nil {
			t.Fatal(mode, err)
		}
	}
	if !isPrivateIP([]byte{10, 0, 0, 1}) || isPrivateIP([]byte{8, 8, 8, 8}) {
		t.Fatal("private ip")
	}
}

func TestReadOnlyGuardsAndShutdownEmbedded(t *testing.T) {
	offlinetest.Prepare(t)
	settings.ReadOnly = true
	SetSettings(settings.BTsets)
	SetDefSettings()
	RemTorrent(strings.Repeat("ab", 20))
	settings.ReadOnly = false

	settings.Embedded = true
	called := false
	settings.EmbeddedStop = func() { called = true }
	Shutdown()
	if !called {
		t.Fatal("embedded stop")
	}
}

func TestExpiredDropsIdleTorrent(t *testing.T) {
	tor, _ := connectShow(t)
	hash := tor.Hash()
	tor.muTorrent.Lock()
	tor.expiredTime = time.Now().Add(-time.Second)
	tor.muTorrent.Unlock()
	if !tor.expired() {
		t.Fatal("expected expired")
	}
	tor.progressEvent()
	if bts.GetTorrent(hash) != nil {
		t.Fatal("expired torrent still registered")
	}
}

func TestSetSettingsReconnects(t *testing.T) {
	connectShow(t)
	sets := *settings.BTsets
	SetSettings(&sets)
	if bts == nil || bts.client == nil {
		t.Fatal("client not reconnected")
	}
	var buf bytes.Buffer
	WriteStatus(&buf)
	if buf.Len() == 0 {
		t.Fatal("empty status")
	}
}

func TestFileStatsFromData(t *testing.T) {
	if FileStatsFromData("") != nil || FileStatsFromData("{") != nil {
		t.Fatal("bad data")
	}
	got := FileStatsFromData(`{"TorrServer":{"Files":[{"id":1,"path":"a.mkv","length":3}]}}`)
	if len(got) != 1 || got[0].Path != "a.mkv" {
		t.Fatalf("%+v", got)
	}
}
