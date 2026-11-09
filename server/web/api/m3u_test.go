package api

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/gin-contrib/location/v2"
	"github.com/gin-gonic/gin"

	sets "server/settings"
	"server/torr"
	"server/torr/state"
	"server/web/sslcerts"
)

func TestMediaBaseURL(t *testing.T) {
	oldPath, oldPort, oldSets := sets.Path, sets.Port, sets.BTsets
	t.Cleanup(func() { sets.Path, sets.Port, sets.BTsets = oldPath, oldPort, oldSets })
	sets.Path = t.TempDir()
	sets.Port = "8090"

	genCert, genKey, err := sslcerts.MakeCertKeyFiles(nil)
	if err != nil {
		t.Fatal(err)
	}
	userDir := t.TempDir()
	userCert, userKey := filepath.Join(userDir, "c.pem"), filepath.Join(userDir, "k.pem")
	for src, dst := range map[string]string{genCert: userCert, genKey: userKey} {
		b, _ := os.ReadFile(src)
		os.WriteFile(dst, b, 0o600)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(location.Default())
	r.GET("/playlist", func(c *gin.Context) { c.String(http.StatusOK, mediaBaseURL(c)) })

	oldSsl, oldArgs := sets.Ssl, sets.Args
	t.Cleanup(func() { sets.Ssl, sets.Args = oldSsl, oldArgs })
	sets.Ssl = true

	tests := []struct {
		name       string
		cert, key  string
		forceHTTPS bool
		httpMedia  bool
		host       string
		tls        bool
		want       string
	}{
		{"self-signed, http serves media -> http port", genCert, genKey, false, false, "192.168.1.2:8091", true, "http://192.168.1.2:8090"},
		{"ipv6 host", genCert, genKey, false, false, "[fe80::1]:8091", true, "http://[fe80::1]:8090"},
		{"self-signed with --force-https --http-media -> http port", genCert, genKey, true, true, "192.168.1.2:8091", true, "http://192.168.1.2:8090"},
		{"self-signed with --force-https only stays https", genCert, genKey, true, false, "192.168.1.2:8091", true, "https://192.168.1.2:8091"},
		{"user cert stays https", userCert, userKey, false, false, "tv.example.com:8091", true, "https://tv.example.com:8091"},
		{"plain http unchanged", genCert, genKey, false, false, "192.168.1.2:8090", false, "http://192.168.1.2:8090"},
	}
	for _, tt := range tests {
		sets.BTsets = &sets.BTSets{SslCert: tt.cert, SslKey: tt.key}
		sets.Args = &sets.ExecArgs{ForceHTTPS: tt.forceHTTPS, HTTPMedia: tt.httpMedia}
		req := httptest.NewRequest(http.MethodGet, "/playlist", nil)
		req.Host = tt.host
		if tt.tls {
			req.TLS = &tls.ConnectionState{}
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if got := w.Body.String(); got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestMediaBaseEndpoint(t *testing.T) {
	oldPath, oldPort, oldSets, oldSsl, oldArgs, oldAuth := sets.Path, sets.Port, sets.BTsets, sets.Ssl, sets.Args, sets.HttpAuth
	t.Cleanup(func() {
		sets.Path, sets.Port, sets.BTsets, sets.Ssl, sets.Args, sets.HttpAuth = oldPath, oldPort, oldSets, oldSsl, oldArgs, oldAuth
	})
	sets.Path = t.TempDir()
	sets.Port = "8090"
	sets.Ssl = true
	cert, key, err := sslcerts.MakeCertKeyFiles(nil)
	if err != nil {
		t.Fatal(err)
	}
	sets.BTsets = &sets.BTSets{SslCert: cert, SslKey: key}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(location.Default())
	SetupRoute(r)

	get := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/mediabase", nil)
		req.Host = "192.168.1.2:8091"
		req.TLS = &tls.ConnectionState{}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	tests := []struct {
		name              string
		forceHTTPS, media bool
		want              string
	}{
		{"self-signed, http serves media", false, false, `{"base":"http://192.168.1.2:8090"}`},
		{"strict --force-https keeps https", true, false, `{"base":"https://192.168.1.2:8091"}`},
		{"--force-https --http-media", true, true, `{"base":"http://192.168.1.2:8090"}`},
	}
	for _, tt := range tests {
		sets.Args = &sets.ExecArgs{ForceHTTPS: tt.forceHTTPS, HTTPMedia: tt.media}
		w := get()
		if w.Code != http.StatusOK || w.Body.String() != tt.want {
			t.Errorf("%s: %d %s, want %s", tt.name, w.Code, w.Body.String(), tt.want)
		}
		if cc := w.Header().Get("Cache-Control"); cc != "no-store" {
			t.Errorf("%s: Cache-Control = %q, want no-store", tt.name, cc)
		}
	}

	// the settings change at runtime: a trusted cert takes effect without restart
	userDir := t.TempDir()
	userCert, userKey := filepath.Join(userDir, "c.pem"), filepath.Join(userDir, "k.pem")
	b, _ := os.ReadFile(cert)
	os.WriteFile(userCert, b, 0o644)
	b, _ = os.ReadFile(key)
	os.WriteFile(userKey, b, 0o600)
	sets.BTsets = &sets.BTSets{SslCert: userCert, SslKey: userKey}
	sets.Args = &sets.ExecArgs{}
	if w := get(); w.Body.String() != `{"base":"https://192.168.1.2:8091"}` {
		t.Errorf("trusted cert: %s, want https base", w.Body.String())
	}

	// requires authentication when --httpauth is on
	sets.HttpAuth = true
	if w := get(); w.Code != http.StatusUnauthorized {
		t.Errorf("without credentials: %d, want 401", w.Code)
	}
}

func TestStatusFromSpecMatchesTorrentFileIDs(t *testing.T) {
	info := metainfo.Info{
		Name:        "Show",
		PieceLength: 16 << 10,
		Files: []metainfo.FileInfo{
			{Path: []string{"Episode 10.mp4"}, Length: 1},
			{Path: []string{"Episode 2.mp4"}, Length: 1},
			{Path: []string{"Episode 3.mp4"}, Length: 1},
		},
	}
	infoBytes, err := bencode.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	tr := &torr.Torrent{TorrentSpec: &torrent.TorrentSpec{InfoBytes: infoBytes}}

	// Status numbers files by their library path sorted with utils.CompareStrings,
	// and stream links resolve index= against those ids.
	want := []string{"Show/Episode 2.mp4", "Show/Episode 3.mp4", "Show/Episode 10.mp4"}
	st := statusFromSpec(tr)
	if st == nil || len(st.FileStats) != len(want) {
		t.Fatalf("statusFromSpec = %+v", st)
	}
	for i, f := range st.FileStats {
		if f.Id != i+1 || f.Path != want[i] {
			t.Errorf("file %d = {Id: %d, Path: %q}, want {Id: %d, Path: %q}", i, f.Id, f.Path, i+1, want[i])
		}
	}
}

func TestFindFileNamesakesNeedsTheWholeName(t *testing.T) {
	paths := []string{
		"Show/Show S01E1.mkv",
		"Show/Show S01E10.mkv",
		"Show/Show S01E11.mkv",
		"Show/Show S01E1.rus.mka",
		"Show/Sound/Show S01E1 [Studio].mka",
		"Show/Subs/Show S01E1/English.srt",
		"Show/Фильм.mkv",
		"Show/Фильмы.mkv",
		"Show/Фильм.ru.srt",
	}
	var files []*state.TorrentFileStat
	for i, p := range paths {
		files = append(files, &state.TorrentFileStat{Id: i + 1, Path: p})
	}
	names := func(video int) []string {
		var got []string
		for _, f := range findFileNamesakes(files, files[video]) {
			got = append(got, f.Path)
		}
		return got
	}

	want := []string{"Show/Show S01E1.rus.mka", "Show/Sound/Show S01E1 [Studio].mka", "Show/Subs/Show S01E1/English.srt"}
	if got := names(0); !slices.Equal(got, want) {
		t.Errorf("S01E1 namesakes = %q, want %q", got, want)
	}
	if got := names(1); len(got) != 0 {
		t.Errorf("S01E10 namesakes = %q, want none", got)
	}
	if got := names(6); !slices.Equal(got, []string{"Show/Фильм.ru.srt"}) {
		t.Errorf("Фильм namesakes = %q", got)
	}
}

func TestSendM3UETagFollowsBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// gin buffers the status, so read it from the context writer.
	send := func(m3u, ifNoneMatch string) (int, http.Header, string) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/playlist", nil)
		if ifNoneMatch != "" {
			c.Request.Header.Set("If-None-Match", ifNoneMatch)
		}
		sendM3U(c, "movie.m3u", "08ada5a7", m3u)
		return c.Writer.Status(), w.Header(), w.Body.String()
	}

	_, header, _ := send("#EXTM3U\nhttp://h/stream/a.mkv\n", "")
	tag := header.Get("ETag")
	if tag == "" {
		t.Fatal("no ETag")
	}
	if status, _, _ := send("#EXTM3U\nhttp://h/stream/a.mkv\n", tag); status != http.StatusNotModified {
		t.Fatalf("same playlist: status = %d, want 304", status)
	}
	status, _, body := send("#EXTM3U\nhttp://h/stream/b.mkv\n", tag)
	if status != http.StatusOK || !strings.Contains(body, "b.mkv") {
		t.Fatalf("changed playlist with the old ETag: status = %d, body %q, want 200 with the new list", status, body)
	}
}
