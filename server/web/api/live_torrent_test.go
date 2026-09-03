package api_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// Live torrent HTTP checks against a running TorrServer (data/ settings, real hash).
//
//	LIVE_TORRENT=1 TORRSERVER_URL=http://127.0.0.1:18091 TS_USER=ts TS_PASS=ts \
//	  go test ./web/api -run TestLiveTorrent -count=1 -timeout 3m
func TestLiveTorrentStreamStatPlayFFP(t *testing.T) {
	base := strings.TrimRight(os.Getenv("TORRSERVER_URL"), "/")
	if os.Getenv("LIVE_TORRENT") != "1" || base == "" {
		t.Skip("set LIVE_TORRENT=1 TORRSERVER_URL=http://127.0.0.1:PORT")
	}
	hash := os.Getenv("LIVE_TORRENT_HASH")
	if hash == "" {
		hash = "e5a5bdb8ff6152657a1e051024b618dd37d76957"
	}
	user := getenv("TS_USER", "ts")
	pass := getenv("TS_PASS", "ts")

	client := &http.Client{Timeout: 60 * time.Second}
	do := func(method, url, body string, extra map[string]string) *http.Response {
		t.Helper()
		var rdr io.Reader
		if body != "" {
			rdr = strings.NewReader(body)
		}
		req, err := http.NewRequest(method, url, rdr)
		if err != nil {
			t.Fatal(err)
		}
		req.SetBasicAuth(user, pass)
		req.Header.Set("Accept", "application/json")
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		for k, v := range extra {
			req.Header.Set(k, v)
		}
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		return res
	}

	readClose := func(res *http.Response) []byte {
		t.Helper()
		defer func() { _ = res.Body.Close() }()
		raw, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}

	echo := do(http.MethodGet, base+"/echo", "", nil)
	b := readClose(echo)
	if echo.StatusCode != 200 || !strings.Contains(string(b), "MatriX") {
		t.Fatalf("echo %d %s", echo.StatusCode, b)
	}

	add := do(http.MethodPost, base+"/torrents", `{"action":"add","link":"`+hash+`","save_to_db":true}`, nil)
	_ = readClose(add)

	var stat map[string]any
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		res := do(http.MethodGet, base+"/stream?link="+hash+"&stat", "", nil)
		raw := readClose(res)
		if res.StatusCode == 200 && json.Unmarshal(raw, &stat) == nil {
			if files, _ := stat["file_stats"].([]any); len(files) > 0 {
				break
			}
		}
		stat = nil
		time.Sleep(2 * time.Second)
	}
	if stat == nil {
		t.Fatal("stream?stat never returned file_stats")
	}
	files, _ := stat["file_stats"].([]any)
	fileID := "1"
	if len(files) > 0 {
		if m, ok := files[0].(map[string]any); ok {
			if v, ok := m["id"].(float64); ok {
				fileID = fmt.Sprintf("%d", int(v))
			}
		}
	}

	play := do(http.MethodGet, base+"/stream?link="+hash+"&index="+fileID+"&play", "", map[string]string{
		"Range": "bytes=0-1023",
	})
	chunk := readClose(play)
	if play.StatusCode != 200 && play.StatusCode != 206 {
		n := min(len(chunk), 200)
		t.Fatalf("play %d %s", play.StatusCode, chunk[:n])
	}
	if len(chunk) == 0 {
		t.Fatal("empty play body")
	}

	ffp := do(http.MethodGet, base+"/ffp/"+hash+"/"+fileID, "", nil)
	ffpBody := readClose(ffp)
	if ffp.StatusCode == 200 {
		var probe map[string]any
		if err := json.Unmarshal(ffpBody, &probe); err != nil {
			n := min(len(ffpBody), 200)
			t.Fatalf("ffp json: %v %s", err, ffpBody[:n])
		}
		if probe["format"] == nil && probe["streams"] == nil {
			n := min(len(ffpBody), 200)
			t.Fatalf("ffp missing format/streams: %s", ffpBody[:n])
		}
	} else {
		n := min(len(ffpBody), 200)
		t.Logf("ffp non-200 (%d) — ffprobe may be missing: %s", ffp.StatusCode, ffpBody[:n])
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
