package version

import (
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/anacrolix/torrent"
)

func TestGetTorrentVersion(t *testing.T) {
	got := GetTorrentVersion()
	if got == "" {
		t.Fatal("expected the replaced anacrolix/torrent module version")
	}
	SetRequest(nil)
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	SetRequest(req)
	if req.Header.Get("User-Agent") != UserAgent() {
		t.Fatal(req.Header.Get("User-Agent"))
	}
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	client := srv.Client()
	client.Transport = WithUserAgent(nil)
	client.Transport = WithUserAgent(client.Transport)
	res, err := client.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if gotUA != UserAgent() {
		t.Fatalf("ua %q", gotUA)
	}
}
