package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anacrolix/torrent"
	"github.com/gin-gonic/gin"

	set "server/settings"
	"server/torr"
	"server/torr/offlinetest"
	"server/torr/state"
)

// TestHandlerErrorPaths walks the refusal and error branches of the torrent
// handlers on one offline client: every failure after the handler accepted the
// request must carry a status, never an empty 200.
func TestHandlerErrorPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	offlinetest.Prepare(t)
	set.Embedded = true
	set.EmbeddedStop = func() {}
	bt := torr.NewBTS()
	torr.InitApiHelper(bt)
	if err := bt.Connect(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(bt.Disconnect)

	show := offlinetest.WriteShow(t)
	seed := offlinetest.Seed(t, show)
	tor, err := torr.AddTorrent(show.Spec, "Show", "", "", "tv")
	if err != nil || !tor.GotInfo() {
		t.Fatal(err)
	}
	tor.AddPeers([]torrent.Peer{seed.Peer()})
	hash := tor.Hash().HexString()
	unknown := strings.Repeat("ab", 20)

	var meta bytes.Buffer
	if err := show.Meta.Write(&meta); err != nil {
		t.Fatal(err)
	}
	fileSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(meta.Bytes())
	}))
	defer fileSrv.Close()

	r := gin.New()
	// A request with X-Anon behaves like an unauthenticated media client.
	r.Use(func(c *gin.Context) {
		if c.GetHeader("X-Anon") != "" {
			c.Set("auth_required", true)
		}
	})
	r.POST("/torrents", torrents)
	r.GET("/stream", stream)
	r.GET("/stream/*fname", stream)
	r.GET("/noauth", streamNoAuth)
	r.GET("/play/:hash/:id", play)
	r.GET("/subs/:hash/:id", subsTracks)

	do := func(method, path, body string, anon bool) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		if anon {
			req.Header.Set("X-Anon", "1")
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}
	want := func(name string, w *httptest.ResponseRecorder, code int) {
		t.Helper()
		if w.Code != code {
			t.Errorf("%s: %d, want %d (%s)", name, w.Code, code, w.Body)
		}
	}

	t.Run("torrents actions", func(t *testing.T) {
		want("bad json", do(http.MethodPost, "/torrents", `{nope`, false), http.StatusBadRequest)
		want("unknown action", do(http.MethodPost, "/torrents", `{"action":"explode"}`, false), http.StatusBadRequest)
		want("add without link", do(http.MethodPost, "/torrents", `{"action":"add"}`, false), http.StatusBadRequest)
		want("add bad link", do(http.MethodPost, "/torrents", `{"action":"add","link":"ftp://x"}`, false), http.StatusBadRequest)
		want("add bad torrs", do(http.MethodPost, "/torrents", `{"action":"add","link":"torrs://not-a-token"}`, false), http.StatusBadRequest)
		for _, act := range []string{"get", "set", "rem", "drop"} {
			want(act+" without hash", do(http.MethodPost, "/torrents", `{"action":"`+act+`"}`, false), http.StatusBadRequest)
		}
		want("get unknown", do(http.MethodPost, "/torrents", `{"action":"get","hash":"`+unknown+`"}`, false), http.StatusNotFound)
	})

	t.Run("add keeps the imported timestamp", func(t *testing.T) {
		w := do(http.MethodPost, "/torrents", `{"action":"add","link":"`+fileSrv.URL+`/show.torrent","timestamp":12345}`, false)
		want("add", w, http.StatusOK)
		var st state.TorrentStatus
		if err := json.Unmarshal(w.Body.Bytes(), &st); err != nil || st.Timestamp != 12345 {
			t.Fatalf("timestamp %d err %v body %s", st.Timestamp, err, w.Body)
		}
	})

	t.Run("stream without credentials", func(t *testing.T) {
		want("bad link", do(http.MethodGet, "/stream?link=ftp://x&play", "", true), http.StatusBadRequest)
		w := do(http.MethodGet, "/stream?link="+hash, "", true)
		want("no play or m3u", w, http.StatusUnauthorized)
		if w.Header().Get("WWW-Authenticate") == "" {
			t.Error("401 without WWW-Authenticate")
		}
		want("play unknown hash", do(http.MethodGet, "/stream?link="+unknown+"&index=1&play", "", true), http.StatusUnauthorized)
		m := do(http.MethodGet, "/stream/list.m3u?link="+hash+"&index=1&m3u", "", true)
		want("m3u known hash", m, http.StatusOK)
		if !strings.HasPrefix(m.Body.String(), "#EXTM3U") {
			t.Errorf("m3u body %q", m.Body)
		}
		want("play without index", do(http.MethodGet, "/stream?link="+hash+"&play", "", true), http.StatusBadRequest)
		want("play known hash", do(http.MethodGet, "/stream?link="+hash+"&index=1&play", "", true), http.StatusOK)
	})

	t.Run("streamNoAuth direct", func(t *testing.T) {
		want("no link", do(http.MethodGet, "/noauth", "", false), http.StatusBadRequest)
		want("bad torrs", do(http.MethodGet, "/noauth?link=torrs://nope&play", "", false), http.StatusBadRequest)
		want("bad link", do(http.MethodGet, "/noauth?link=ftp://x&play", "", false), http.StatusBadRequest)
		want("preload only", do(http.MethodGet, "/noauth?link="+hash+"&index=1&preload", "", false), http.StatusUnauthorized)
		want("title overrides", do(http.MethodGet, "/noauth?link="+hash+"&index=1&play&title=T&poster=P&category=tv", "", false), http.StatusOK)
	})

	t.Run("play", func(t *testing.T) {
		want("unknown hash", do(http.MethodGet, "/play/"+unknown+"/1", "", false), http.StatusInternalServerError)
		want("unknown hash anonymous", do(http.MethodGet, "/play/"+unknown+"/1", "", true), http.StatusUnauthorized)
		want("bad hash", do(http.MethodGet, "/play/zzz/1", "", false), http.StatusInternalServerError)
		want("bad index", do(http.MethodGet, "/play/"+hash+"/x", "", false), http.StatusBadRequest)
	})

	t.Run("subs", func(t *testing.T) {
		want("unknown hash", do(http.MethodGet, "/subs/"+unknown+"/1", "", false), http.StatusNotFound)
		want("unknown hash anonymous", do(http.MethodGet, "/subs/"+unknown+"/1", "", true), http.StatusUnauthorized)
		want("bad hash", do(http.MethodGet, "/subs/zzz/1", "", false), http.StatusInternalServerError)
		want("bad index", do(http.MethodGet, "/subs/"+hash+"/x", "", false), http.StatusBadRequest)
	})

	t.Run("wipe removes everything", func(t *testing.T) {
		want("wipe", do(http.MethodPost, "/torrents", `{"action":"wipe"}`, false), http.StatusOK)
		if left := torr.ListTorrent(); len(left) != 0 {
			t.Fatalf("%d torrents left after wipe", len(left))
		}
		want("get after wipe", do(http.MethodPost, "/torrents", `{"action":"get","hash":"`+hash+`"}`, false), http.StatusNotFound)
	})
}

// restartServices applies settings to DLNA, Bonjour and rutor. With all three
// off it only stops them; it must not bind a port or leave a goroutine.
func TestRestartServicesWithEverythingOff(t *testing.T) {
	offlinetest.Prepare(t)
	set.BTsets.EnableDLNA = false
	set.BTsets.EnableBonjour = false
	set.BTsets.EnableRutorSearch = false
	restartServices()
}
