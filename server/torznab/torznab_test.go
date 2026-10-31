package torznab

import (
	"net/http"
	"net/http/httptest"
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
