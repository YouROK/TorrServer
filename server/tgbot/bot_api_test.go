package tgbot

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anacrolix/torrent"

	"server/settings"
	"server/tgbot/config"
	"server/torr"
	"server/torr/offlinetest"
)

type tgAPI struct {
	mu        sync.Mutex
	q         []json.RawMessage
	want      int
	delivered int
	idle      chan struct{}
	docs      int
	file      []byte
	msgID     int
	updID     int
}

func (a *tgAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = io.Copy(io.Discard, r.Body)
	path := r.URL.Path
	if strings.Contains(path, "/file/bot") {
		_, _ = w.Write(a.file)
		return
	}
	method := path[strings.LastIndex(path, "/")+1:]
	switch method {
	case "getMe":
		writeJSON(w, map[string]any{"ok": true, "result": map[string]any{"id": 1, "is_bot": true, "username": "testbot", "first_name": "T"}})
	case "getUpdates":
		a.mu.Lock()
		var batch []json.RawMessage
		if len(a.q) > 0 {
			batch = a.q[:1]
			a.q = a.q[1:]
		}
		a.delivered += len(batch)
		ready := a.delivered >= a.want && len(batch) == 0
		a.mu.Unlock()
		if len(batch) == 0 {
			if ready {
				select {
				case a.idle <- struct{}{}:
				default:
				}
			}
			timer := time.NewTimer(15 * time.Millisecond)
			select {
			case <-r.Context().Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
		writeJSON(w, map[string]any{"ok": true, "result": batch})
	case "getFile":
		writeJSON(w, map[string]any{"ok": true, "result": map[string]any{"file_id": "f1", "file_path": "docs/show.torrent"}})
	case "sendMessage", "editMessageText", "sendDocument", "copyMessage":
		if method == "sendDocument" {
			a.mu.Lock()
			a.docs++
			a.mu.Unlock()
		}
		a.mu.Lock()
		a.msgID++
		id := a.msgID
		a.mu.Unlock()
		writeJSON(w, map[string]any{"ok": true, "result": map[string]any{
			"message_id": id,
			"text":       "ok",
			"chat":       map[string]any{"id": 42, "type": "private"},
			"from":       map[string]any{"id": 1, "is_bot": true, "first_name": "T"},
		}})
	default:
		writeJSON(w, map[string]any{"ok": true, "result": true})
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (a *tgAPI) push(v any) {
	buf, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	a.mu.Lock()
	a.want++
	a.q = append(a.q, buf)
	a.mu.Unlock()
}

func (a *tgAPI) wait(t *testing.T) {
	t.Helper()
	deadline := time.After(20 * time.Second)
	for {
		select {
		case <-a.idle:
			a.mu.Lock()
			ok := a.delivered >= a.want && len(a.q) == 0
			a.mu.Unlock()
			if ok {
				return
			}
		case <-deadline:
			a.mu.Lock()
			t.Fatalf("updates %d/%d queued %d", a.delivered, a.want, len(a.q))
			a.mu.Unlock()
		}
	}
}

func (a *tgAPI) text(body string, doc bool) {
	a.mu.Lock()
	a.updID++
	id := a.updID
	a.mu.Unlock()
	msg := map[string]any{
		"message_id": id,
		"from":       map[string]any{"id": 42, "is_bot": false, "first_name": "P"},
		"chat":       map[string]any{"id": 42, "type": "private"},
		"text":       body,
	}
	if strings.HasPrefix(body, "/") {
		n := len(body)
		if i := strings.IndexByte(body, ' '); i > 0 {
			n = i
		}
		msg["entities"] = []any{map[string]any{"type": "bot_command", "offset": 0, "length": n}}
	}
	if doc {
		delete(msg, "text")
		msg["document"] = map[string]any{"file_id": "f1", "file_unique_id": "u1", "file_name": "show.torrent", "mime_type": "application/x-bittorrent"}
	}
	a.push(map[string]any{"update_id": id, "message": msg})
}

func (a *tgAPI) callback(data string) {
	a.mu.Lock()
	a.updID++
	id := a.updID
	a.mu.Unlock()
	a.push(map[string]any{
		"update_id": id,
		"callback_query": map[string]any{
			"id":   "cb" + strings.ReplaceAll(data, "|", "_"),
			"from": map[string]any{"id": 42, "is_bot": false, "first_name": "P"},
			"data": data,
			"message": map[string]any{
				"message_id": 9,
				"chat":       map[string]any{"id": 42, "type": "private"},
				"from":       map[string]any{"id": 1, "is_bot": true, "first_name": "T"},
			},
		},
	})
}

func TestBotCommandsAndUpload(t *testing.T) {
	offlinetest.Prepare(t)
	settings.Embedded = true
	settings.EmbeddedStop = func() {}
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

	var raw bytes.Buffer
	if err := show.Meta.Write(&raw); err != nil {
		t.Fatal(err)
	}

	api := &tgAPI{idle: make(chan struct{}, 1), file: raw.Bytes()}
	srv := httptest.NewServer(api)
	t.Cleanup(srv.Close)

	cfg, _ := json.Marshal(&config.Config{HostTG: srv.URL, WhiteIds: []int64{42}})
	if err := os.WriteFile(filepath.Join(settings.Path, "tg.cfg"), cfg, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Start("123:test"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		Stop()
		for id := 1; id < 200; id++ {
			stopStatusRefresh(id)
			stopSnakeRefresh(id)
		}
	})
	api.wait(t)

	step := func(v any) {
		api.push(v)
		api.wait(t)
		time.Sleep(300 * time.Millisecond)
	}
	for _, c := range []string{
		"/start", "/help", "/id", "/more", "/cancel", "/list", "/clear", "/add",
		"/status " + hash, "/server", "/link " + hash + " 1", "/play " + hash + " 1",
		"/cache", "/m3u " + hash, "/m3uall", "/search", "/search show", "/rutor show", "/torznab show",
		"/preload " + hash + " 1", "/queue",
		"/hash", "/export", "/import", "/categories", "/echo hi", "/db", "/viewed",
		"/ffp", "/speedtest 1", "/lang", "/stats", "/stat", "/snake " + hash,
		"/shutdown", "/settings", "/preset",
	} {
		api.text(c, false)
		api.wait(t)
		time.Sleep(300 * time.Millisecond)
	}
	api.text("hello", false)
	api.wait(t)
	time.Sleep(300 * time.Millisecond)
	step(map[string]any{
		"update_id": 9000,
		"inline_query": map[string]any{
			"id": "q1", "from": map[string]any{"id": 42, "is_bot": false, "first_name": "P"},
			"query": "list", "offset": "",
		},
	})

	for _, data := range []string{
		"\ffiles|" + hash,
		"\fflist|0",
		"\ffrefresh|0",
		"\fftpick|" + hash + "|0",
		"\ffstatus|" + hash,
		"\ffstatusstop|" + hash,
		"\ffm3u|" + hash,
		"\fflink|" + hash + "|1",
		"\ffmenu|hub|lib",
		"\ffmenu|help",
		"\ffset|export",
		"\ffset|more",
		"\ffset|reset_confirm",
		"\ffset|reset_def|0",
		"\ffclear|0",
		"\ffshutdown|0",
		"\ffpreset|0",
		"\ffnop",
		"nope",
		"\fcancel|99",
	} {
		api.callback(data)
		api.wait(t)
		time.Sleep(300 * time.Millisecond)
	}

	api.mu.Lock()
	before := api.docs
	api.mu.Unlock()
	api.callback("\fupload|" + hash + "|1")
	api.wait(t)
	deadline := time.Now().Add(25 * time.Second)
	for time.Now().Before(deadline) {
		api.mu.Lock()
		n := api.docs
		api.mu.Unlock()
		if n > before {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("upload did not send a document")
}
