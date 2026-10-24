package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/gin-gonic/gin"

	"server/ffprobe"
	"server/rutor"
	set "server/settings"
	"server/torr"
	"server/torr/offlinetest"
	"server/torr/state"
)

func TestOfflineAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	offlinetest.Prepare(t)
	set.Embedded = true
	stopped := make(chan struct{}, 2)
	set.EmbeddedStop = func() { stopped <- struct{}{} }
	set.Args.Pprof = true
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

	var meta bytes.Buffer
	if err := show.Meta.Write(&meta); err != nil {
		t.Fatal(err)
	}
	fileSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(meta.Bytes())
	}))
	defer fileSrv.Close()

	r := gin.New()
	SetupRoute(r)
	r.GET("/na", func(c *gin.Context) {
		c.Set("auth_required", true)
		stream(c)
	})

	do := func(method, path, body string, hdr map[string]string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if body != "" && hdr["Content-Type"] == "" {
			req.Header.Set("Content-Type", "application/json")
		}
		for k, v := range hdr {
			req.Header.Set(k, v)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	add := do(http.MethodPost, "/torrents", `{"action":"add","link":"`+fileSrv.URL+`/show.torrent","title":"Show","category":"tv","save_to_db":true}`, nil)
	if add.Code != 200 {
		t.Fatalf("add %d %s", add.Code, add.Body)
	}
	var st state.TorrentStatus
	if err := json.Unmarshal(add.Body.Bytes(), &st); err != nil || len(st.FileStats) != 4 {
		t.Fatalf("status %v %s", err, add.Body)
	}
	id := st.FileStats[0].Id
	srt := id
	for _, f := range st.FileStats {
		if strings.HasSuffix(f.Path, ".srt") {
			srt = f.Id
		}
		if strings.HasSuffix(f.Path, "episode1.mkv") {
			id = f.Id
		}
	}

	if w := do(http.MethodPost, "/torrents", `{"action":"list"}`, nil); w.Code != 200 || !strings.Contains(w.Body.String(), hash) {
		t.Fatalf("list %d %s", w.Code, w.Body)
	}
	if w := do(http.MethodPost, "/torrents", `{"action":"get","hash":"`+hash+`"}`, nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if w := do(http.MethodPost, "/torrents", `{"action":"set","hash":"`+hash+`","title":"Renamed","category":"tv","poster":""}`, nil); w.Code != 200 {
		t.Fatalf("set %d %s", w.Code, w.Body)
	}

	stat := do(http.MethodGet, "/stream?link="+hash+"&stat", "", nil)
	if stat.Code != 200 {
		t.Fatal(stat.Body)
	}
	pre := do(http.MethodGet, "/stream?link="+hash+"&index="+strconv.Itoa(id)+"&preload&stat", "", nil)
	if pre.Code != 200 {
		t.Fatal(pre.Body)
	}
	m3u := do(http.MethodGet, "/stream/Show.m3u?link="+hash+"&index="+strconv.Itoa(id)+"&m3u&fromlast", "", nil)
	if m3u.Code != 200 || !strings.Contains(m3u.Body.String(), "#EXTM3U") {
		t.Fatalf("m3u %d %s", m3u.Code, m3u.Body)
	}
	play := do(http.MethodGet, "/stream?link="+hash+"&index="+strconv.Itoa(id)+"&play", "", map[string]string{"Range": "bytes=0-1023"})
	if play.Code != 200 && play.Code != 206 {
		t.Fatalf("play %d %s", play.Code, play.Body)
	}
	if play.Body.Len() == 0 {
		t.Fatal("empty range")
	}
	head := do(http.MethodHead, "/stream?link="+hash+"&index="+strconv.Itoa(id)+"&play", "", nil)
	if head.Code != 200 && head.Code != 206 {
		t.Fatalf("head %d", head.Code)
	}
	bad := do(http.MethodGet, "/stream?link="+hash+"&index=99&play", "", nil)
	if bad.Code < 400 {
		t.Fatalf("bad id %d", bad.Code)
	}
	if w := do(http.MethodGet, "/stream", "", nil); w.Code < 400 {
		t.Fatal("empty link")
	}
	if w := do(http.MethodGet, "/stream?link=ftp://x&save=1", "", nil); w.Code < 400 {
		t.Fatal("bad scheme")
	}
	if w := do(http.MethodGet, "/stream?link="+hash+"&save=1&title=Show&category=tv", "", nil); w.Code != 200 {
		t.Fatalf("save %d %s", w.Code, w.Body)
	}
	if w := do(http.MethodPost, "/torrents", `{"action":"add","link":"ftp://x"}`, nil); w.Code < 400 {
		t.Fatal("bad add")
	}
	GetGStreamerSettingsDoc()
	UpdateGStreamerSettingsDoc()
	GetGStreamerEchoDoc()
	RemoveGStreamerTaskDoc()
	GetGStreamerHeartbeatDoc()
	ProbeGStreamerSourceDoc()
	GetGStreamerMasterPlaylistDoc()
	GetGStreamerInitSegmentDoc()
	GetGStreamerMediaSegmentDoc()

	set.MaxSize = 1
	limited := do(http.MethodGet, "/play/"+hash+"/"+strconv.Itoa(id), "", map[string]string{"Range": "bytes=0-15"})
	set.MaxSize = 0
	if limited.Code != http.StatusForbidden && limited.Code < 400 {
		t.Fatalf("maxsize %d", limited.Code)
	}
	direct := do(http.MethodGet, "/play/"+hash+"/"+strconv.Itoa(id), "", map[string]string{"Range": "bytes=0-15"})
	if direct.Code != 200 && direct.Code != 206 {
		t.Fatalf("direct %d %s", direct.Code, direct.Body)
	}
	if w := do(http.MethodHead, "/play/"+hash+"/"+strconv.Itoa(id), "", nil); w.Code != 200 && w.Code != 206 {
		t.Fatalf("play head %d", w.Code)
	}
	if w := do(http.MethodGet, "/play/"+hash+"/nope", "", nil); w.Code < 400 {
		t.Fatal("bad play id")
	}

	na := do(http.MethodGet, "/na?link="+hash+"&index="+strconv.Itoa(id)+"&play", "", map[string]string{"Range": "bytes=0-15"})
	if na.Code != 200 && na.Code != 206 {
		t.Fatalf("noauth play %d %s", na.Code, na.Body)
	}
	if w := do(http.MethodGet, "/na?link="+hash, "", nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("noauth stat %d", w.Code)
	}
	if w := do(http.MethodGet, "/na", "", nil); w.Code < 400 {
		t.Fatal("noauth empty")
	}
	if w := do(http.MethodGet, "/na?link="+hash+"&m3u&fromlast", "", nil); w.Code != 200 || !strings.Contains(w.Body.String(), "#EXTM3U") {
		t.Fatalf("noauth m3u %d %s", w.Code, w.Body)
	}

	pl := do(http.MethodGet, "/playlist/Show.m3u?hash="+hash+"&fromlast=true", "", nil)
	if pl.Code != 200 || !strings.Contains(pl.Body.String(), "#EXTM3U") {
		t.Fatalf("playlist %d %s", pl.Code, pl.Body)
	}
	if w := do(http.MethodGet, "/playlist?hash=", "", nil); w.Code < 400 {
		t.Fatal("empty playlist hash")
	}
	set.BTsets.MergeAllM3U = true
	all := do(http.MethodGet, "/playlistall/all.m3u?fromlast=true&fn=Show&category=tv", "", nil)
	set.BTsets.MergeAllM3U = false
	if all.Code != 200 || !strings.Contains(all.Body.String(), "#EXTM3U") {
		t.Fatalf("all %d %s", all.Code, all.Body)
	}

	subs := do(http.MethodGet, "/subs/"+hash+"/"+strconv.Itoa(id), "", nil)
	if subs.Code != 200 {
		t.Fatalf("subs %d %s", subs.Code, subs.Body)
	}
	vtt := do(http.MethodGet, "/subs/"+hash+"/"+strconv.Itoa(id)+"?track="+strconv.Itoa(srt), "", nil)
	if vtt.Code != 200 || vtt.Body.Len() == 0 {
		t.Fatalf("vtt %d %s", vtt.Code, vtt.Body)
	}
	track := do(http.MethodGet, "/subs/"+hash+"/"+strconv.Itoa(id)+"?track="+strconv.Itoa(srt)+"&format=original&encoding=windows-1251", "", nil)
	if track.Code != 200 || track.Body.Len() == 0 {
		t.Fatalf("track %d %s", track.Code, track.Body)
	}

	if w := do(http.MethodPost, "/viewed", `{"action":"set","hash":"`+hash+`","file_index":1}`, nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if w := do(http.MethodPost, "/viewed", `{"action":"list","hash":"`+hash+`"}`, nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if w := do(http.MethodPost, "/viewed", `{"action":"rem","hash":"`+hash+`","file_index":1}`, nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if w := do(http.MethodPost, "/viewed", `{`, nil); w.Code < 400 {
		t.Fatal("bad viewed")
	}
	if w := do(http.MethodPost, "/viewed", `{"action":"set"}`, nil); w.Code < 400 {
		t.Fatal("viewed hash")
	}
	if w := do(http.MethodPost, "/torrents", `{"action":"rem"}`, nil); w.Code < 400 {
		t.Fatal("rem hash")
	}
	if w := do(http.MethodPost, "/torrents", `{"action":"drop"}`, nil); w.Code < 400 {
		t.Fatal("drop hash")
	}
	if w := do(http.MethodGet, "/playlist?hash="+strings.Repeat("ab", 20), "", nil); w.Code != http.StatusNotFound {
		t.Fatalf("missing playlist %d", w.Code)
	}
	if w := do(http.MethodGet, "/play/"+strings.Repeat("ab", 20)+"/1", "", nil); w.Code < 400 {
		t.Fatal("missing play")
	}

	if w := do(http.MethodPost, "/cache", `{"action":"get","hash":"`+hash+`"}`, nil); w.Code != 200 {
		t.Fatalf("cache %d %s", w.Code, w.Body)
	}
	if w := do(http.MethodPost, "/cache", `{"action":"get"}`, nil); w.Code < 400 {
		t.Fatal("cache hash")
	}

	if w := do(http.MethodPost, "/settings", `{"action":"get"}`, nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
	set.ReadOnly = true
	if w := do(http.MethodPost, "/settings", `{"action":"set","sets":`+do(http.MethodPost, "/settings", `{"action":"get"}`, nil).Body.String()+`}`, nil); w.Code != http.StatusForbidden {
		t.Fatalf("readonly set settings %d %s", w.Code, w.Body)
	}
	if w := do(http.MethodPost, "/settings", `{"action":"def"}`, nil); w.Code != http.StatusForbidden {
		t.Fatalf("readonly def settings %d %s", w.Code, w.Body)
	}
	set.ReadOnly = false
	if w := do(http.MethodPost, "/settings", `{"action":"set"}`, nil); w.Code != http.StatusBadRequest {
		t.Fatalf("nil sets %d %s", w.Code, w.Body)
	}
	if w := do(http.MethodPost, "/settings", `{"action":""}`, nil); w.Code < 400 {
		t.Fatal("empty settings")
	}

	if w := do(http.MethodGet, "/mediabase", "", nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if w := do(http.MethodGet, "/waf", "", nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if w := do(http.MethodGet, "/search/show", "", nil); w.Code != 200 && w.Code != 400 {
		t.Fatal(w.Code)
	}
	if w := do(http.MethodGet, "/torznab/search/show", "", nil); w.Code != 400 {
		t.Fatal(w.Body)
	}
	if w := do(http.MethodGet, "/torznab/caps", "", nil); w.Code != 400 {
		t.Fatal(w.Body)
	}
	if w := do(http.MethodPost, "/torznab/test", `{"host":"","key":""}`, nil); w.Code != 200 || !strings.Contains(w.Body.String(), "false") {
		t.Fatalf("torznab test %d %s", w.Code, w.Body)
	}
	if w := do(http.MethodGet, "/tmdb/settings", "", nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if w := do(http.MethodGet, "/gst/settings", "", nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if w := do(http.MethodPost, "/gst/settings", `{}`, nil); w.Code != http.StatusNotFound {
		t.Fatal(w.Code)
	}
	if w := do(http.MethodGet, "/ffp/status", "", nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if !ffprobe.Exists() {
		if w := do(http.MethodGet, "/ffp/"+hash+"/"+strconv.Itoa(id), "", nil); w.Code != http.StatusNotFound {
			t.Fatalf("ffp %d %s", w.Code, w.Body)
		}
	}
	if w := do(http.MethodGet, "/runtime/status", "", nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if w := do(http.MethodGet, "/download/0", "", nil); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := do(http.MethodGet, "/download/1", "", nil); w.Code != 200 || w.Body.Len() == 0 {
		t.Fatalf("download1 %d %d", w.Code, w.Body.Len())
	}
	if w := do(http.MethodGet, "/download/nope", "", nil); w.Code == 200 && w.Body.Len() > 0 {
		t.Fatal("bad size")
	}

	for _, p := range []string{
		"/debug/pprof/", "/debug/pprof/cmdline", "/debug/pprof/symbol",
		"/debug/pprof/allocs", "/debug/pprof/heap", "/debug/pprof/goroutine",
		"/debug/pprof/mutex", "/debug/pprof/block", "/debug/pprof/threadcreate",
	} {
		if w := do(http.MethodGet, p, "", nil); w.Code != 200 {
			t.Fatalf("%s %d", p, w.Code)
		}
	}

	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	part, err := mw.CreateFormFile("file", "show.torrent")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(meta.Bytes()); err != nil {
		t.Fatal(err)
	}
	_ = mw.WriteField("save", "1")
	_ = mw.WriteField("title", "Uploaded")
	_ = mw.WriteField("category", "tv")
	_ = mw.Close()
	up := do(http.MethodPost, "/torrent/upload", body.String(), map[string]string{"Content-Type": mw.FormDataContentType()})
	if up.Code != 200 {
		t.Fatalf("upload %d %s", up.Code, up.Body)
	}

	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			do(http.MethodPost, "/torrents", `{"action":"list"}`, nil)
			do(http.MethodGet, "/stream?link="+hash+"&stat", "", nil)
		}()
	}
	wg.Wait()

	if w := do(http.MethodPost, "/torrents", `{"action":"drop","hash":"`+hash+`"}`, nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if w := do(http.MethodPost, "/torrents", `{"action":"rem","hash":"`+hash+`"}`, nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if w := do(http.MethodPost, "/torrents", `{"action":"wipe"}`, nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
	// Storage writes BTsets while a live torrent's watch loop reads them.
	time.Sleep(300 * time.Millisecond)
	if w := do(http.MethodGet, "/storage/settings", "", nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
	if w := do(http.MethodPost, "/storage/settings", `{"settings":"nope"}`, nil); w.Code != 400 {
		t.Fatal(w.Body)
	}
	if w := do(http.MethodPost, "/storage/settings", `{"settings":"json"}`, nil); w.Code != 200 {
		t.Fatal(w.Body)
	}
	rutor.Stop()
	if w := do(http.MethodPost, "/storage/settings", "settings=bbolt&viewed=json", map[string]string{"Content-Type": "application/x-www-form-urlencoded"}); w.Code != 200 {
		t.Fatal(w.Body)
	}
	rutor.Stop()
	if w := do(http.MethodPost, "/storage/settings", `{}`, nil); w.Code != 400 {
		t.Fatal(w.Body)
	}
	set.ReadOnly = true
	if w := do(http.MethodPost, "/storage/settings", `{"settings":"json"}`, nil); w.Code != http.StatusForbidden {
		t.Fatal(w.Code)
	}
	set.ReadOnly = false
	if w := do(http.MethodGet, "/shutdown", "", nil); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := do(http.MethodGet, "/shutdown/tests", "", nil); w.Code != 200 {
		t.Fatal(w.Code)
	}
	for i := 0; i < 2; i++ {
		select {
		case <-stopped:
		case <-time.After(3 * time.Second):
			t.Fatal("shutdown did not finish")
		}
	}
}
