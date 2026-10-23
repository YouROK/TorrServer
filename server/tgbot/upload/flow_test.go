package upload

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anacrolix/torrent"
	tele "gopkg.in/telebot.v4"

	"server/tgbot/config"
	"server/torr"
	"server/torr/offlinetest"
	"server/torr/state"
)

type apiMock struct {
	mu   sync.Mutex
	docs int
	id   int
}

func (a *apiMock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = io.Copy(io.Discard, r.Body)
	method := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	w.Header().Set("Content-Type", "application/json")
	if method == "sendMessage" || method == "editMessageText" || method == "sendDocument" {
		if method == "sendDocument" {
			a.mu.Lock()
			a.docs++
			a.mu.Unlock()
		}
		a.mu.Lock()
		a.id++
		id := a.id
		a.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": map[string]any{
			"message_id": id,
			"chat":       map[string]any{"id": 7, "type": "private"},
		}})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": true})
}

func testCtx(t *testing.T, a *apiMock) tele.Context {
	t.Helper()
	srv := httptest.NewServer(a)
	t.Cleanup(srv.Close)
	b, err := tele.NewBot(tele.Settings{URL: srv.URL, Token: "123:test", Offline: true, Client: srv.Client()})
	if err != nil {
		t.Fatal(err)
	}
	return b.NewContext(tele.Update{Message: &tele.Message{
		ID: 1, Sender: &tele.User{ID: 7, FirstName: "T"}, Chat: &tele.Chat{ID: 7, Type: tele.ChatPrivate},
	}})
}

func TestQueueFullShowAndCancel(t *testing.T) {
	a := &apiMock{}
	c := testCtx(t, a)
	m := &Manager{queue: make([]*Worker, 51)}
	m.AddRange(c, strings.Repeat("ab", 20), 1, 1)
	if len(m.queue) != 51 {
		t.Fatal("queue accepted past the cap")
	}

	manager.queueLock.Lock()
	manager.queue = []*Worker{{torrentHash: strings.Repeat("ab", 20), c: c, msg: &tele.Message{ID: 1}}}
	manager.working = map[int]*Worker{1: {torrentHash: strings.Repeat("cd", 20), c: c}}
	manager.queueLock.Unlock()
	if err := ShowQueue(c); err != nil {
		t.Fatal(err)
	}
	manager.queueLock.Lock()
	manager.queue, manager.working = nil, nil
	manager.queueLock.Unlock()

	w := &Worker{id: 3, c: c, msg: &tele.Message{ID: 1, Chat: &tele.Chat{ID: 7, Type: tele.ChatPrivate}}, torrentHash: "h"}
	m = &Manager{queue: []*Worker{w}, working: map[int]*Worker{4: {id: 4}}}
	m.Cancel(3)
	m.Cancel(4)
	if len(m.queue) != 0 || !w.isCancelled || !m.working[4].isCancelled {
		t.Fatal("cancel")
	}
}

func TestUploadOneFile(t *testing.T) {
	offlinetest.Prepare(t)
	config.Cfg = &config.Config{HostTG: "http://127.0.0.1"}
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

	a := &apiMock{}
	c := testCtx(t, a)
	Start()
	AddRange(c, strings.Repeat("ab", 20), 1, 1)
	Cancel(1)

	TrFunc = func(_ int64, key string) string { return key }
	EscapeFunc = func(s string) string { return s }
	t.Cleanup(func() { TrFunc, EscapeFunc = nil, nil })
	m := &Manager{}
	m.Start()
	t.Cleanup(m.Stop)
	t.Cleanup(Stop)
	m.AddRange(c, tor.Hash().HexString(), 1, 2)
	m.AddRange(c, tor.Hash().HexString(), 3, 3)
	queued := &Manager{}
	for i := 0; i < 40; i++ {
		queued.queue = append(queued.queue, &Worker{
			c: c, msg: &tele.Message{ID: 1, Chat: &tele.Chat{ID: 7, Type: tele.ChatPrivate}},
			torrentHash: strings.Repeat("ab", 20),
		})
	}
	queued.sendQueueStatus()
	_ = tr(1, "k")
	_ = escapeHtml("a<b")
	loading(&Worker{
		isCancelled: true, c: c, from: 1, to: 1, torrentHash: tor.Hash().HexString(),
		msg: &tele.Message{ID: 3, Chat: &tele.Chat{ID: 7, Type: tele.ChatPrivate}},
	})
	_ = uploadFile(&Worker{
		c: c, torrentHash: tor.Hash().HexString(), msg: &tele.Message{ID: 2, Chat: &tele.Chat{ID: 7, Type: tele.ChatPrivate}},
	}, &state.TorrentFileStat{Id: 99, Path: "missing.bin", Length: 8}, 1, 1)

	deadline := time.Now().Add(25 * time.Second)
	for time.Now().Before(deadline) {
		a.mu.Lock()
		n := a.docs
		a.mu.Unlock()
		if n > 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("document was not sent")
}
