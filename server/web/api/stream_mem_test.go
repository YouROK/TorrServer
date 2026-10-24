package api

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/gin-gonic/gin"

	"server/leakcheck"
	"server/torr"
	"server/torr/offlinetest"
)

func TestStreamDisconnectReleasesReaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	offlinetest.Prepare(t)
	bt := torr.NewBTS()
	torr.InitApiHelper(bt)
	if err := bt.Connect(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(bt.Disconnect)

	show := bulkShow(t, 1<<20)
	seed := offlinetest.Seed(t, show)
	tor, err := torr.AddTorrent(show.Spec, "Movie", "", "", "movie")
	if err != nil || !tor.GotInfo() {
		t.Fatal(err)
	}
	tor.AddPeers([]torrent.Peer{seed.Peer()})
	hash := tor.Hash().HexString()

	r := gin.New()
	SetupRoute(r)
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}, Timeout: 20 * time.Second}
	t.Cleanup(client.CloseIdleConnections)

	one := func() error {
		req, err := http.NewRequest(http.MethodGet, srv.URL+"/stream?link="+hash+"&index=1&play", nil)
		if err != nil {
			return err
		}
		req.Header.Set("Range", "bytes=0-65535")
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer func() { _ = res.Body.Close() }()
		if res.StatusCode != http.StatusPartialContent && res.StatusCode != http.StatusOK {
			buf, _ := io.ReadAll(io.LimitReader(res.Body, 200))
			return &statusErr{res.StatusCode, buf}
		}
		buf := make([]byte, 4096)
		if _, err := io.ReadFull(res.Body, buf); err != nil {
			return err
		}
		return nil
	}
	if err := one(); err != nil {
		t.Fatal(err)
	}

	baseG := runtime.NumGoroutine()
	warm := leakcheck.Heap()

	var wg sync.WaitGroup
	errc := make(chan error, 100)
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/stream?link="+hash+"&index=1&play", nil)
			if err != nil {
				errc <- err
				return
			}
			req.Header.Set("Range", "bytes=0-1048575")
			res, err := client.Do(req)
			if err != nil {
				errc <- err
				return
			}
			buf := make([]byte, 4096)
			_, _ = io.ReadFull(res.Body, buf)
			_ = res.Body.Close()
		}()
	}
	wg.Wait()
	close(errc)
	for err := range errc {
		t.Error(err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for torr.GetActiveStreams() != 0 || (tor.GetCache() != nil && tor.GetCache().Readers() != 0) {
		if time.Now().After(deadline) {
			t.Fatalf("after disconnect: streams %d readers %d", torr.GetActiveStreams(), readersOf(tor))
		}
		time.Sleep(10 * time.Millisecond)
	}

	after := leakcheck.Heap()
	const slack = uint64(16 << 20)
	if after.HeapAlloc > warm.HeapAlloc+slack || after.HeapObjects > warm.HeapObjects+20000 {
		t.Fatalf("heap alloc %d->%d objects %d->%d", warm.HeapAlloc, after.HeapAlloc, warm.HeapObjects, after.HeapObjects)
	}
	leakcheck.AssertNoGoroutineLeak(t, baseG)
}

func readersOf(tor *torr.Torrent) int {
	if tor.GetCache() == nil {
		return 0
	}
	return tor.GetCache().Readers()
}

type statusErr struct {
	code int
	body []byte
}

func (e *statusErr) Error() string {
	return http.StatusText(e.code) + ": " + string(e.body)
}

func bulkShow(t testing.TB, n int) *offlinetest.Show {
	t.Helper()
	root := filepath.Join(t.TempDir(), "Movie")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	data := bytes.Repeat([]byte("Z"), n)
	if err := os.WriteFile(filepath.Join(root, "movie.mkv"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	info := metainfo.Info{PieceLength: 256 * 1024}
	if err := info.BuildFromFilePath(root); err != nil {
		t.Fatal(err)
	}
	var raw bytes.Buffer
	for _, fi := range info.UpvertedFiles() {
		b, err := os.ReadFile(filepath.Join(root, filepath.Join(fi.Path...)))
		if err != nil {
			t.Fatal(err)
		}
		raw.Write(b)
	}
	mi := &metainfo.MetaInfo{InfoBytes: bencode.MustMarshal(info)}
	return &offlinetest.Show{Meta: mi, Spec: torrent.TorrentSpecFromMetaInfo(mi), Data: raw.Bytes()}
}
