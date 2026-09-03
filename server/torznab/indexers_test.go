package torznab

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"server/settings"
)

// JacRed Torznab endpoint stored in tests (Jackett-compatible path).
const (
	JacRedSite = "https://jacred.stream"
	JacRedHost = "https://jacred.stream/api/v2.0/indexers/all/results/torznab/"
	JacRedKey  = "pp"
)

const jacRedJackettXML = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:torznab="http://torznab.com/schemas/2015/feed">
  <channel>
    <title>JacRed</title>
    <item>
      <title>The Matrix 1999</title>
      <jackettindexer>JacRed</jackettindexer>
      <size>1234567890</size>
      <enclosure url="magnet:?xt=urn:btih:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" length="1234567890" type="application/x-bittorrent"/>
    </item>
    <item>
      <title>Silo S01</title>
      <prowlarrindexer>JacRed-Prowlarr</prowlarrindexer>
      <size>222</size>
      <enclosure url="magnet:?xt=urn:btih:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" length="222" type="application/x-bittorrent"/>
    </item>
  </channel>
</rss>`

func TestNormalizeHostJacRedStream(t *testing.T) {
	tests := []struct {
		host, want string
	}{
		{JacRedHost, "https://jacred.stream/api/v2.0/indexers/all/results/torznab/api"},
		{JacRedHost + "api", "https://jacred.stream/api/v2.0/indexers/all/results/torznab/api"},
		{JacRedSite, "https://jacred.stream/api"},
	}
	for _, tt := range tests {
		u, err := normalizeHost(tt.host)
		if err != nil {
			t.Fatalf("normalizeHost(%q): %v", tt.host, err)
		}
		if got := u.String(); got != tt.want {
			t.Fatalf("normalizeHost(%q) = %q, want %q", tt.host, got, tt.want)
		}
	}
}

func TestSearchJacRedJackettAndProwlarrMock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("apikey") != JacRedKey {
			http.Error(w, "bad key", http.StatusUnauthorized)
			return
		}
		if r.URL.Query().Get("t") != "search" {
			http.Error(w, "bad t", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(jacRedJackettXML))
	}))
	t.Cleanup(srv.Close)

	prev := settings.BTsets
	settings.BTsets = &settings.BTSets{
		EnableTorznabSearch: true,
		TorznabUrls: []settings.TorznabConfig{
			{Host: srv.URL, Key: JacRedKey, Name: "JacRed", CatType: settings.CategoryAll},
		},
	}
	t.Cleanup(func() { settings.BTsets = prev })

	got := Search(context.Background(), "matrix", -1, "", 0, 50)
	if len(got) != 2 {
		t.Fatalf("len=%d want 2", len(got))
	}
	if !strings.Contains(got[0].Name, "Matrix") {
		t.Fatalf("first title %q", got[0].Name)
	}
	if got[0].Tracker != "JacRed" {
		t.Fatalf("jackett tracker = %q", got[0].Tracker)
	}
	if got[1].Tracker != "JacRed-Prowlarr" {
		t.Fatalf("prowlarr tracker = %q", got[1].Tracker)
	}
}
