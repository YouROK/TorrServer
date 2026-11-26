package torr

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"

	"server/settings"
	"server/torr/state"
)

// newSeasonTorrent adds a torrent whose info is known up front. Its files are stored in
// byte order (Episode 1, Episode 10, ..., Episode 2), which differs from the order Status
// sorts by, and there are more than 12 of them so sort.Slice does not fall back to
// insertion sort.
func newSeasonTorrent(t *testing.T) (*BTServer, *Torrent) {
	t.Helper()
	initTestGlobals(t)
	// with the zero value the watcher closes an idle torrent on its first tick
	prev := settings.BTsets.TorrentDisconnectTimeout
	settings.BTsets.TorrentDisconnectTimeout = 30
	t.Cleanup(func() { settings.BTsets.TorrentDisconnectTimeout = prev })

	bt := NewBTS()
	InitApiHelper(bt)
	if err := bt.Connect(); err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(bt.Disconnect)

	const episodes = 24
	info := metainfo.Info{Name: "Show", PieceLength: 1 << 10}
	for i := 1; i <= episodes; i++ {
		info.Files = append(info.Files, metainfo.FileInfo{
			Path:   []string{fmt.Sprintf("Episode %d.mkv", i)},
			Length: 1 << 10,
		})
	}
	sortFilesByPath(info.Files)
	info.Pieces = make([]byte, 20*episodes)
	mi := metainfo.MetaInfo{InfoBytes: bencode.MustMarshal(info)}
	spec := &torrent.TorrentSpec{
		InfoBytes:   mi.InfoBytes,
		InfoHash:    mi.HashInfoBytes(),
		DisplayName: info.Name,
	}
	tor, err := NewTorrent(spec, bt)
	if err != nil {
		t.Fatalf("NewTorrent: %v", err)
	}
	if !tor.GotInfo() {
		t.Fatal("GotInfo returned false for a torrent with info")
	}
	return bt, tor
}

// sortFilesByPath puts the files in plain byte order, like most torrent creators do.
func sortFilesByPath(files []metainfo.FileInfo) {
	for i := 1; i < len(files); i++ {
		for j := i; j > 0 && files[j].Path[0] < files[j-1].Path[0]; j-- {
			files[j], files[j-1] = files[j-1], files[j]
		}
	}
}

// Stream requests call GotInfo several times at once for one file while players and
// Lampa poll the status. Once the info is in, GotInfo must not take the torrent through
// TorrentGettingInfo again: other requests read Stat meanwhile.
func TestGotInfoKeepsWorkingStateUnderConcurrentRequests(t *testing.T) {
	_, tor := newSeasonTorrent(t)

	var wg sync.WaitGroup
	stop := make(chan struct{})
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					tor.GotInfo()
				}
			}
		}()
	}

	seen := map[state.TorrentStat]int{}
	for i := 0; i < 20000; i++ {
		seen[tor.Status().Stat]++
		tor.expired()
	}
	close(stop)
	wg.Wait()

	for st, n := range seen {
		if st != state.TorrentWorking {
			t.Errorf("Status saw %v %d times while GotInfo ran concurrently, want only %v", st, n, state.TorrentWorking)
		}
	}
}

// Files returns the torrent's own slice, which also backs every piece's file list.
// Status lists the files sorted by name and must not reorder that slice.
func TestStatusDoesNotReorderTorrentFiles(t *testing.T) {
	_, tor := newSeasonTorrent(t)

	before := append([]*torrent.File(nil), tor.Files()...)
	st := tor.Status()
	sorted := false
	for i, fs := range st.FileStats {
		if fs.Path != before[i].Path() {
			sorted = true
		}
	}
	if !sorted {
		t.Fatal("Status lists the files in torrent order, so this test checks nothing")
	}
	for i, f := range tor.Files() {
		if f != before[i] {
			t.Fatalf("Status reordered the torrent's files: position %d was %q, now %q", i, before[i].Path(), f.Path())
		}
	}
}

func TestGotInfoDoesNotReopenClosedTorrent(t *testing.T) {
	_, tor := newSeasonTorrent(t)

	tor.Close()
	if tor.GotInfo() {
		t.Fatal("GotInfo returned true for a closed torrent")
	}
	if st := tor.Status().Stat; st != state.TorrentClosed {
		t.Fatalf("Stat after Close and GotInfo is %v, want %v", st, state.TorrentClosed)
	}
}

// A failure inside Stream must answer the request. Returning without a write made gin
// send 200 OK with an empty body, which players take for a zero-length file.
func TestStreamAnswersUnknownFile(t *testing.T) {
	_, tor := newSeasonTorrent(t)

	req := httptest.NewRequest(http.MethodGet, "/stream?play&index=99", nil)
	req.Header.Set("Range", "bytes=0-")
	w := httptest.NewRecorder()
	if err := tor.Stream(99, req, w); err == nil {
		t.Fatal("Stream returned no error for a missing file")
	}
	if w.Code != http.StatusNotFound {
		t.Fatalf("status %d, want %d", w.Code, http.StatusNotFound)
	}
}
