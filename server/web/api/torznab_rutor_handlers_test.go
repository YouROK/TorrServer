package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"

	"server/rutor/models"
	sets "server/settings"
)

const fakeFeed = `<?xml version="1.0"?><rss version="2.0"><channel>
<item><title>Show S01</title><link>magnet:?xt=urn:btih:2f6da5bb5f1ea1994be6814ae3cbcc8bdf911eaf</link></item>
</channel></rss>`

const fakeCaps = `<?xml version="1.0"?><caps>
<limits max="250" default="50"/>
<categories><category id="2000" name="Movies"><subcat id="2010" name="HD"/></category></categories>
</caps>`

// withBTsets swaps the global BTsets for one test.
func withBTsets(t *testing.T, mutate func(*sets.BTSets)) {
	t.Helper()
	old := sets.BTsets
	cp := &sets.BTSets{}
	if old != nil {
		*cp = *old
	}
	mutate(cp)
	sets.BTsets = cp
	t.Cleanup(func() { sets.BTsets = old })
}

func hit(r *gin.Engine, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

func torznabRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/torznab/search", torznabSearch)
	r.GET("/torznab/caps", torznabCaps)
	r.POST("/torznab/test", torznabTest)
	r.GET("/search", rutorSearch)
	return r
}

func fakeIndexer(t *testing.T) (*httptest.Server, func() []string) {
	t.Helper()
	var mu sync.Mutex
	var queries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		queries = append(queries, r.URL.RawQuery)
		mu.Unlock()
		if r.URL.Query().Get("t") == "caps" {
			_, _ = w.Write([]byte(fakeCaps))
			return
		}
		_, _ = w.Write([]byte(fakeFeed))
	}))
	t.Cleanup(srv.Close)
	return srv, func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), queries...)
	}
}

func TestTorznabHandlersDisabled(t *testing.T) {
	withBTsets(t, func(s *sets.BTSets) { s.EnableTorznabSearch = false })
	r := torznabRouter()
	for _, p := range []string{"/torznab/search?query=x", "/torznab/caps?index=0"} {
		if w := hit(r, p); w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "torznab disabled") {
			t.Errorf("%s: %d %s", p, w.Code, w.Body)
		}
	}
}

func TestTorznabSearchClampsLimitAndReturnsResults(t *testing.T) {
	srv, queries := fakeIndexer(t)
	withBTsets(t, func(s *sets.BTSets) {
		s.EnableTorznabSearch = true
		s.TorznabUrls = []sets.TorznabConfig{{Host: srv.URL, Key: "k", Name: "fake"}}
	})
	r := torznabRouter()

	w := hit(r, "/torznab/search?query=show%2520x&index=0&offset=5&limit=9999&cat=5000")
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var list []*models.TorrentDetails
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil || len(list) != 1 || list[0].Title != "Show S01" {
		t.Fatalf("list %v err %v", list, err)
	}
	qs := queries()
	if len(qs) != 1 {
		t.Fatalf("indexer queries %v", qs)
	}
	for _, want := range []string{"limit=250", "offset=5", "cat=5000", "apikey=k", "t=search"} {
		if !strings.Contains(qs[0], want) {
			t.Errorf("query %q lacks %s", qs[0], want)
		}
	}

	// A bad limit/offset falls back to defaults; index=-1 searches every indexer from offset 0.
	if w := hit(r, "/torznab/search?query=a&limit=-3&offset=-1&index=zz"); w.Code != 200 {
		t.Fatalf("defaults: %d %s", w.Code, w.Body)
	}
	last := queries()[len(queries())-1]
	if !strings.Contains(last, "limit=100") || strings.Contains(last, "offset=") {
		t.Errorf("default query %q", last)
	}
}

func TestTorznabSearchNoIndexerIsEmptyArray(t *testing.T) {
	withBTsets(t, func(s *sets.BTSets) {
		s.EnableTorznabSearch = true
		s.TorznabUrls = nil
	})
	w := hit(torznabRouter(), "/torznab/search?query=x")
	if w.Code != 200 || strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatalf("%d %q", w.Code, w.Body)
	}
}

func TestTorznabCaps(t *testing.T) {
	srv, _ := fakeIndexer(t)
	withBTsets(t, func(s *sets.BTSets) {
		s.EnableTorznabSearch = true
		s.TorznabUrls = []sets.TorznabConfig{
			{Host: srv.URL, Key: "k"},
			{Host: "", Key: ""},
			{Host: "http://127.0.0.1:1", Key: "k"}, // refused
		}
	})
	r := torznabRouter()

	w := hit(r, "/torznab/caps?index=0")
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var resp torznabCapsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Limits.Max != 250 || resp.Limits.Default != 50 || len(resp.Categories) != 1 ||
		resp.Categories[0].ID != "2000" || len(resp.Categories[0].Subcategories) != 1 ||
		resp.Categories[0].Subcategories[0].Name != "HD" {
		t.Fatalf("caps %+v", resp)
	}

	for path, want := range map[string]int{
		"/torznab/caps":          http.StatusBadRequest, // index required
		"/torznab/caps?index=x":  http.StatusBadRequest,
		"/torznab/caps?index=-1": http.StatusBadRequest,
		"/torznab/caps?index=9":  http.StatusBadRequest,
		"/torznab/caps?index=1":  http.StatusBadRequest, // not configured
		"/torznab/caps?index=2":  http.StatusInternalServerError,
	} {
		if w := hit(r, path); w.Code != want {
			t.Errorf("%s: %d, want %d (%s)", path, w.Code, want, w.Body)
		}
	}
}

func TestTorznabTestEndpoint(t *testing.T) {
	srv, _ := fakeIndexer(t)
	r := torznabRouter()
	post := func(body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/torznab/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		return w
	}
	if w := post(`{"host":"` + srv.URL + `","key":"k"}`); w.Code != 200 || !strings.Contains(w.Body.String(), `"success":true`) {
		t.Fatalf("ok: %d %s", w.Code, w.Body)
	}
	if w := post(`{"host":"","key":"k"}`); w.Code != 200 || !strings.Contains(w.Body.String(), `"success":false`) {
		t.Fatalf("empty host: %d %s", w.Code, w.Body)
	}
	if w := post(`{not json`); w.Code != http.StatusBadRequest {
		t.Fatalf("bad json: %d %s", w.Code, w.Body)
	}
}

func TestRutorSearchHandler(t *testing.T) {
	r := torznabRouter()

	withBTsets(t, func(s *sets.BTSets) { s.EnableRutorSearch = false })
	if w := hit(r, "/search?query=x"); w.Code != http.StatusBadRequest || strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatalf("disabled: %d %q", w.Code, w.Body)
	}

	// Enabled but the index is not loaded: an empty array, not null.
	withBTsets(t, func(s *sets.BTSets) { s.EnableRutorSearch = true })
	if w := hit(r, "/search?query=matrix%2520x"); w.Code != 200 || strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatalf("enabled: %d %q", w.Code, w.Body)
	}
}
