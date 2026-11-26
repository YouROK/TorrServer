package torznab

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"server/settings"
)

func TestSearchOneFormatsSizeInBinaryUnits(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<rss><channel>
<item><title>a</title><enclosure url="magnet:?xt=urn:btih:aa" length="1610612736" type="application/x-bittorrent"/></item>
<item><title>b</title><size>512</size></item>
</channel></rss>`))
	}))
	defer srv.Close()

	results := searchOne(srv.URL, "key", "", settings.CategoryAll, "q")
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Size != "1.5 GiB" {
		t.Errorf("size = %q, want %q", results[0].Size, "1.5 GiB")
	}
	if results[1].Size != "512 B" {
		t.Errorf("size = %q, want %q", results[1].Size, "512 B")
	}
}

func TestSearchOneAcceptsHostWithoutScheme(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("t") == "caps" {
			w.Write([]byte(`<caps></caps>`))
			return
		}
		w.Write([]byte(`<rss><channel><item><title>a</title><size>1</size></item></channel></rss>`))
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	if err := Test(host, "key"); err != nil {
		t.Fatalf("Test(%q) = %v", host, err)
	}
	if results := searchOne(host, "key", "", settings.CategoryAll, "q"); len(results) != 1 {
		t.Fatalf("searchOne(%q) got %d results, want 1", host, len(results))
	}
}
