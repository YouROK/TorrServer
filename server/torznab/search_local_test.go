package torznab

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"server/settings"
)

func useIndexer(t *testing.T, h http.HandlerFunc) string {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	prev := httpClient
	httpClient = srv.Client()
	t.Cleanup(func() { httpClient = prev })
	return srv.URL
}

const searchFeed = `<?xml version="1.0"?><rss version="2.0"><channel>
<item>
  <title>With Enclosure</title>
  <link>http://x/page</link>
  <pubDate>Mon, 02 Jan 2006 15:04:05 +0000</pubDate>
  <enclosure url="http://x/file.torrent" length="2048" type="application/x-bittorrent"/>
  <attr name="seeders" value="7"/>
  <attr name="peers" value="9"/>
  <attr name="infohash" value="aabbccddeeff00112233445566778899aabbccdd"/>
</item>
<item>
  <title>Magnet Link</title>
  <link>magnet:?xt=urn:btih:00112233445566778899aabbccddeeff00112233&amp;dn=x</link>
  <size>1024</size>
  <attr name="leechers" value="3"/>
  <attr name="peers" value="9"/>
</item>
<item>
  <title>Magnet Attr</title>
  <attr name="magneturl" value="magnet:?xt=urn:btih:ffeeddccbbaa99887766554433221100ffeeddcc"/>
</item>
</channel></rss>`

func TestSearchOneParsesItemsAndQuery(t *testing.T) {
	var gotQuery map[string]string
	url := useIndexer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = map[string]string{}
		for k, v := range r.URL.Query() {
			gotQuery[k] = v[0]
		}
		if r.URL.Path != "/api" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(searchFeed))
	})

	res := searchOne(context.Background(), url, "k", "dune", "label", "2000", 5, 10)
	if len(res) != 3 {
		t.Fatalf("results = %d", len(res))
	}
	for k, want := range map[string]string{"apikey": "k", "t": "search", "q": "dune", "cat": "2000", "offset": "5", "limit": "10"} {
		if gotQuery[k] != want {
			t.Errorf("query %s = %q, want %q", k, gotQuery[k], want)
		}
	}

	a := res[0]
	if a.Link != "http://x/file.torrent" || a.Seed != 7 || a.Peer != 9 || a.Tracker != "label" ||
		a.Hash != "aabbccddeeff00112233445566778899aabbccdd" || a.Size == "" || a.CreateDate.IsZero() {
		t.Errorf("enclosure item = %+v", a)
	}
	b := res[1]
	if b.Magnet != b.Link || !strings.HasPrefix(b.Magnet, "magnet:") || b.Peer != 3 || b.Hash == "" {
		t.Errorf("magnet link item = %+v", b)
	}
	c := res[2]
	if !strings.HasPrefix(c.Magnet, "magnet:") || c.Hash == "" {
		t.Errorf("magnet attr item = %+v", c)
	}
}

func TestSearchOneOmitsOptionalParams(t *testing.T) {
	var raw string
	url := useIndexer(t, func(w http.ResponseWriter, r *http.Request) {
		raw = r.URL.RawQuery
		_, _ = w.Write([]byte(`<rss><channel></channel></rss>`))
	})
	res := searchOne(context.Background(), url, "k", "q", "l", "", 0, 0)
	if len(res) != 0 {
		t.Fatalf("results = %v", res)
	}
	for _, p := range []string{"cat=", "offset=", "limit="} {
		if strings.Contains(raw, p) {
			t.Errorf("query %q contains %q", raw, p)
		}
	}
}

func TestSearchOneFailures(t *testing.T) {
	cases := map[string]http.HandlerFunc{
		"non-200": func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "boom", http.StatusInternalServerError)
		},
		"api error": func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`<error code="100" description="Incorrect user credentials"/>`))
		},
		"bad xml": func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`<rss><channel><item>`))
		},
		"empty body": func(w http.ResponseWriter, r *http.Request) {},
	}
	for name, h := range cases {
		t.Run(name, func(t *testing.T) {
			url := useIndexer(t, h)
			if res := searchOne(context.Background(), url, "k", "q", "l", "", 0, 0); res != nil {
				t.Fatalf("results = %v, want nil", res)
			}
		})
	}
}

func TestSearchOneBadHostAndConnectError(t *testing.T) {
	if res := searchOne(context.Background(), "://bad", "k", "q", "l", "", 0, 0); res != nil {
		t.Fatalf("bad host results = %v", res)
	}

	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()
	if res := searchOne(context.Background(), url, "k", "q", "l", "", 0, 0); res != nil {
		t.Fatalf("closed server results = %v", res)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	live := useIndexer(t, func(w http.ResponseWriter, r *http.Request) {})
	if res := searchOne(ctx, live, "k", "q", "l", "", 0, 0); res != nil {
		t.Fatalf("canceled ctx results = %v", res)
	}
}

func TestSearchMergesIndexers(t *testing.T) {
	url := useIndexer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(searchFeed))
	})
	prev := settings.BTsets
	settings.BTsets = &settings.BTSets{
		EnableTorznabSearch: true,
		TorznabUrls: []settings.TorznabConfig{
			{Host: url, Key: "k", Name: "one"},
			{Host: "", Key: "k"},
			{Host: url, Key: "k", Name: "two"},
		},
	}
	t.Cleanup(func() { settings.BTsets = prev })

	if got := Search(context.TODO(), "q", -1, "", 0, 0); len(got) != 6 {
		t.Fatalf("merged = %d", len(got))
	}
	if got := Search(context.Background(), "q", 2, "", 0, 0); len(got) != 3 || got[0].Tracker != "two" {
		t.Fatalf("indexer 2 = %v", got)
	}
	if got := Search(context.Background(), "q", 1, "", 0, 0); got != nil {
		t.Fatalf("incomplete indexer = %v", got)
	}
}
