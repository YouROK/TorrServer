package torr

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Players read an empty 200 as a zero-length file (AGENTS.md): every failure
// after the handler accepted the request must carry its own status and body.
func TestStreamFailuresAnswerWithStatusAndBody(t *testing.T) {
	tor, _ := connectShow(t)
	id := tor.Status().FileStats[0].Id

	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	req.RemoteAddr = "127.0.0.1:9"

	w := httptest.NewRecorder()
	if err := tor.Stream(9999, req, w); err == nil || w.Code != http.StatusNotFound || w.Body.Len() == 0 {
		t.Errorf("unknown file: err=%v code=%d body=%q", err, w.Code, w.Body.String())
	}

	// drop() nils the cache while Lampac still has a request in flight: there is
	// no reader to hand out, and the client must be told.
	cache := tor.cache
	tor.muTorrent.Lock()
	tor.cache = nil
	tor.muTorrent.Unlock()
	t.Cleanup(func() {
		tor.muTorrent.Lock()
		tor.cache = cache
		tor.muTorrent.Unlock()
	})
	w = httptest.NewRecorder()
	err := tor.Stream(id, req, w)
	if err == nil || w.Code != http.StatusServiceUnavailable || !strings.Contains(w.Body.String(), "reader") {
		t.Errorf("closed torrent: err=%v code=%d body=%q", err, w.Code, w.Body.String())
	}
	if n := cache.Readers(); n != 0 {
		t.Errorf("failed stream left %d readers registered", n)
	}
}

func TestStreamReleasesReaderOnEveryPath(t *testing.T) {
	tor, _ := connectShow(t)
	id := tor.Status().FileStats[0].Id

	for _, rng := range []string{"", "bytes=0-99", "bytes=999999999-", "bytes=abc"} {
		req := httptest.NewRequest(http.MethodGet, "/stream", nil)
		req.RemoteAddr = "127.0.0.1:9"
		if rng != "" {
			req.Header.Set("Range", rng)
		}
		_ = tor.Stream(id, req, httptest.NewRecorder())
		if n := tor.cache.Readers(); n != 0 {
			t.Fatalf("Range %q left %d readers behind", rng, n)
		}
	}
}
