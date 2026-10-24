package pages

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/gin-gonic/gin"

	"server/settings"
	"server/torr"
)

func TestMagnetsEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/magnets", nil)
	getTorrents(c)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "<div>") || strings.Contains(w.Body.String(), "magnet:") {
		t.Fatalf("%d %q", w.Code, w.Body.String())
	}
}

func TestMagnetsFromDB(t *testing.T) {
	dir := t.TempDir()
	oldPath, oldSets, oldRO, oldWA := settings.Path, settings.BTsets, settings.ReadOnly, settings.SearchWA
	settings.Path = dir
	t.Cleanup(func() {
		settings.CloseDB()
		settings.Path, settings.BTsets, settings.ReadOnly, settings.SearchWA = oldPath, oldSets, oldRO, oldWA
	})
	if err := settings.InitSets(false, false); err != nil {
		t.Fatal(err)
	}
	hash := strings.Repeat("cd", 20)
	settings.AddTorrent(&settings.TorrentDB{
		TorrentSpec: &torrent.TorrentSpec{InfoHash: metainfo.NewHashFromHex(hash), DisplayName: "Movie"},
		Title:       "Movie",
	})
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/magnets", nil)
	getTorrents(c)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "magnet:?xt=urn:btih:") {
		t.Fatalf("%d %q", w.Code, w.Body.String())
	}

	torr.InitApiHelper(torr.NewBTS())
	r := gin.New()
	SetupRoute(r)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/stat", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), "not connected") {
		t.Fatalf("stat %d %q", w.Code, w.Body.String())
	}
}
