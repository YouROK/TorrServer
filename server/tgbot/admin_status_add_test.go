package tgbot

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/anacrolix/torrent"
	tele "gopkg.in/telebot.v4"

	"server/settings"
	"server/tgbot/config"
	"server/torr"
	"server/torr/offlinetest"
)

// callRecorder is a Telegram API stub that records sent/edited texts,
// callback answers and deletes, and serves one file for getFile.
type callRecorder struct {
	mu       sync.Mutex
	sent     []string
	edited   []string
	answers  []string
	deleted  int
	id       int
	file     []byte
	failEdit atomic.Bool
}

func (r *callRecorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body, _ := io.ReadAll(req.Body)
	if strings.Contains(req.URL.Path, "/file/bot") {
		_, _ = w.Write(r.file)
		return
	}
	method := req.URL.Path[strings.LastIndex(req.URL.Path, "/")+1:]
	var in struct {
		Text string `json:"text"`
	}
	_ = json.Unmarshal(body, &in)
	message := func(text string) map[string]any {
		r.mu.Lock()
		r.id++
		id := r.id
		r.mu.Unlock()
		return map[string]any{
			"message_id": id, "text": text,
			"chat": map[string]any{"id": 42, "type": "private"},
			"from": map[string]any{"id": 1, "is_bot": true, "first_name": "T"},
		}
	}
	switch method {
	case "sendMessage":
		r.mu.Lock()
		r.sent = append(r.sent, in.Text)
		r.mu.Unlock()
		writeJSON(w, map[string]any{"ok": true, "result": message(in.Text)})
	case "editMessageText":
		if r.failEdit.Load() {
			writeJSON(w, map[string]any{"ok": false, "error_code": 400, "description": "Bad Request: message can't be edited"})
			return
		}
		r.mu.Lock()
		r.edited = append(r.edited, in.Text)
		r.mu.Unlock()
		writeJSON(w, map[string]any{"ok": true, "result": message(in.Text)})
	case "answerCallbackQuery":
		r.mu.Lock()
		r.answers = append(r.answers, in.Text)
		r.mu.Unlock()
		writeJSON(w, map[string]any{"ok": true, "result": true})
	case "deleteMessage":
		r.mu.Lock()
		r.deleted++
		r.mu.Unlock()
		writeJSON(w, map[string]any{"ok": true, "result": true})
	case "getFile":
		writeJSON(w, map[string]any{"ok": true, "result": map[string]any{"file_id": "f1", "file_path": "docs/show.torrent"}})
	default:
		writeJSON(w, map[string]any{"ok": true, "result": true})
	}
}

func (r *callRecorder) lastAnswer() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.answers) == 0 {
		return ""
	}
	return r.answers[len(r.answers)-1]
}

func (r *callRecorder) lastSent() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.sent) == 0 {
		return ""
	}
	return r.sent[len(r.sent)-1]
}

func (r *callRecorder) counts() (sent, edited, deleted int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.sent), len(r.edited), r.deleted
}

func (r *callRecorder) lastID() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.id
}

func TestAdminStatusAddCallbacks(t *testing.T) {
	offlinetest.Prepare(t)
	settings.Embedded = true
	settings.EmbeddedStop = func() {}
	bt := torr.NewBTS()
	torr.InitApiHelper(bt)
	if err := bt.Connect(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(bt.Disconnect)
	oldRutor := startRutor
	startRutor = func() {}
	t.Cleanup(func() { startRutor = oldRutor })

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

	rec := &callRecorder{file: raw.Bytes()}
	srv := httptest.NewServer(rec)
	t.Cleanup(srv.Close)
	oldCfg := config.Cfg
	config.Cfg = &config.Config{HostTG: srv.URL, WhiteIds: []int64{42}}
	t.Cleanup(func() { config.Cfg = oldCfg })

	b, err := tele.NewBot(tele.Settings{URL: srv.URL, Token: "123:test", Offline: true, Client: srv.Client()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for id := 1; id < 300; id++ {
			stopStatusRefresh(id)
		}
	})
	const admin, stranger = int64(42), int64(7)
	chat := &tele.Chat{ID: 42, Type: tele.ChatPrivate}
	msg := func(uid int64, text, payload string) tele.Context {
		return b.NewContext(tele.Update{Message: &tele.Message{Text: text, Payload: payload, Sender: &tele.User{ID: uid}, Chat: chat}})
	}
	cb := func(uid int64, data string) tele.Context {
		return b.NewContext(tele.Update{Callback: &tele.Callback{
			ID: "1", Data: data, Sender: &tele.User{ID: uid},
			Message: &tele.Message{ID: rec.lastID(), Chat: chat, Sender: &tele.User{ID: 1, IsBot: true}},
		}})
	}
	noErr := func(t *testing.T, err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("status refresh and wait", func(t *testing.T) {
		noErr(t, callbackStatusRefresh(cb(admin, ""), ""))
		if rec.lastAnswer() != tr(admin, "callback_unknown") {
			t.Fatalf("empty hash: %q", rec.lastAnswer())
		}
		noErr(t, callbackStatusRefresh(cb(admin, ""), strings.Repeat("0", 40)))
		if rec.lastAnswer() != tr(admin, "torrent_not_found") {
			t.Fatalf("unknown hash: %q", rec.lastAnswer())
		}
		_, _, del := rec.counts()
		noErr(t, callbackStatusRefresh(cb(admin, ""), hash))
		if _, _, d := rec.counts(); d != del+1 {
			t.Fatalf("old status message not deleted: %d -> %d", del, d)
		}
		if !strings.Contains(rec.lastSent(), hash) {
			t.Fatalf("status text lacks hash: %q", rec.lastSent())
		}
		haltStatusLoop(t, rec.lastID())

		// waitForInfoAndUpdateStatus: gone torrent, then a ready one.
		m := &tele.Message{ID: 500, Chat: chat}
		waitForInfoAndUpdateStatus(b, m, strings.Repeat("1", 40), admin)
		rec.mu.Lock()
		gone := rec.edited[len(rec.edited)-1]
		rec.mu.Unlock()
		if !strings.Contains(gone, tr(admin, "torrent_not_found")) || !strings.Contains(gone, strings.Repeat("1", 40)) {
			t.Fatalf("gone: %q", gone)
		}
		waitForInfoAndUpdateStatus(b, m, hash, admin)
		rec.mu.Lock()
		ready := rec.edited[len(rec.edited)-1]
		rec.mu.Unlock()
		if !strings.Contains(ready, hash) {
			t.Fatalf("ready: %q", ready)
		}
		haltStatusLoop(t, m.ID)
	})

	t.Run("edit failure", func(t *testing.T) {
		rec.failEdit.Store(true)
		defer rec.failEdit.Store(false)
		waitForInfoAndUpdateStatus(b, &tele.Message{ID: 501, Chat: chat}, hash, admin)
	})

	t.Run("add from document", func(t *testing.T) {
		c := msg(admin, "", "")
		if _, err := addTorrentFromDocument(c, nil); err == nil {
			t.Fatal("nil doc accepted")
		}
		if _, err := addTorrentFromDocument(c, &tele.Document{}); err == nil {
			t.Fatal("empty file id accepted")
		}
		doc := &tele.Document{File: tele.File{FileID: "f1"}, FileName: "show.torrent"}
		got, err := addTorrentFromDocument(c, doc)
		noErr(t, err)
		if got == nil || got.Hash().HexString() != hash {
			t.Fatalf("document added %v", got)
		}
		if want := strings.Replace(tr(admin, "add_success"), "%s", "show.torrent", 1); rec.lastEdited() != want {
			t.Fatalf("success text %q, want %q", rec.lastEdited(), want)
		}
		doc.FileName = ""
		_, err = addTorrentFromDocument(c, doc)
		noErr(t, err)

		rec.mu.Lock()
		rec.file = []byte("not a torrent")
		rec.mu.Unlock()
		if _, err := addTorrentFromDocument(c, doc); err == nil {
			t.Fatal("garbage file parsed")
		}
		rec.mu.Lock()
		rec.file = raw.Bytes()
		rec.mu.Unlock()
	})

	t.Run("cmdAdd", func(t *testing.T) {
		noErr(t, cmdAdd(msg(admin, "/add", "")))
		if rec.lastSent() == "" {
			t.Fatal("no add prompt")
		}
		if _, err := addTorrent(msg(admin, "", ""), "torrs://!!!"); err == nil {
			t.Fatal("bad torrs link accepted")
		}
		if _, err := addTorrent(msg(admin, "", ""), "not a link"); err == nil {
			t.Fatal("bad link accepted")
		}
		if err := cmdAdd(msg(admin, "/add nope", "nope")); err == nil {
			t.Fatal("bad link returned no error")
		}
		magnet := "magnet:?xt=urn:btih:" + hash
		noErr(t, cmdAdd(msg(admin, "/add "+magnet, magnet)))
		if !strings.Contains(rec.lastSent(), hash) && !strings.Contains(rec.lastSent(), "Show") {
			t.Logf("card text %q", rec.lastSent())
		}
		// Category set: card; category empty: picker.
		tor.Category = ""
		noErr(t, afterAdd(msg(admin, "", ""), tor))
		tor.Category = "tv"
		noErr(t, afterAdd(msg(admin, "", ""), tor))
		noErr(t, afterAdd(msg(admin, "", ""), nil))
	})

	t.Run("preset", func(t *testing.T) {
		before := *settings.BTsets
		if err := cmdPreset(msg(stranger, "/preset low", "low")); err != nil || rec.lastSent() != tr(stranger, "admin_only") {
			t.Fatalf("non-admin: %v %q", err, rec.lastSent())
		}
		noErr(t, cmdPreset(msg(admin, "/preset", "")))
		if rec.lastSent() != tr(admin, "preset_usage") {
			t.Fatalf("usage: %q", rec.lastSent())
		}
		noErr(t, cmdPreset(msg(admin, "/preset cache zero", "")))
		if rec.lastSent() != tr(admin, "preset_usage") {
			t.Fatalf("bad kv: %q", rec.lastSent())
		}
		settings.ReadOnly = true
		noErr(t, cmdPreset(msg(admin, "/preset low", "low")))
		settings.ReadOnly = false
		if rec.lastSent() != tr(admin, "settings_readonly") {
			t.Fatalf("read-only: %q", rec.lastSent())
		}

		pendingFor := func() pendingPreset {
			pendingPresetMu.Lock()
			defer pendingPresetMu.Unlock()
			return pendingPresets[chatMsgKey(chat.ID, rec.lastID())]
		}
		// Confirmation: non-admin leaves the pending entry alone.
		noErr(t, cmdPreset(msg(admin, "/preset cache 8 preload 20 conn 30", "")))
		p := pendingFor()
		if p.Sets == nil || p.Sets.CacheSize != 8<<20 || p.Sets.PreloadCache != 20 || p.Sets.ConnectionsLimit != 30 || p.UserID != admin {
			t.Fatalf("pending preset %+v", p)
		}
		if settings.BTsets.CacheSize != before.CacheSize {
			t.Fatal("settings changed before confirmation")
		}
		noErr(t, presetConfirm(cb(stranger, ""), "1"))
		if rec.lastAnswer() != tr(stranger, "admin_only") || pendingFor().Sets == nil {
			t.Fatalf("non-admin confirm: %q", rec.lastAnswer())
		}
		// Someone else's entry is dropped and canceled.
		pendingPresetMu.Lock()
		pendingPresets[chatMsgKey(chat.ID, rec.lastID())] = pendingPreset{Sets: p.Sets, UserID: stranger}
		pendingPresetMu.Unlock()
		noErr(t, presetConfirm(cb(admin, ""), "1"))
		if rec.lastAnswer() != tr(admin, "canceled") || settings.BTsets.CacheSize != before.CacheSize {
			t.Fatalf("foreign entry: %q", rec.lastAnswer())
		}
		// Unknown message.
		noErr(t, presetConfirm(cb(admin, ""), "1"))
		if rec.lastAnswer() != tr(admin, "canceled") {
			t.Fatalf("unknown: %q", rec.lastAnswer())
		}
		// Decline.
		noErr(t, cmdPreset(msg(admin, "/preset low", "low")))
		noErr(t, presetConfirm(cb(admin, ""), "0"))
		if rec.lastAnswer() != tr(admin, "canceled") || settings.BTsets.ConnectionsLimit != before.ConnectionsLimit {
			t.Fatalf("decline: %q", rec.lastAnswer())
		}
		// Read-only at confirm time.
		noErr(t, cmdPreset(msg(admin, "/preset low", "low")))
		settings.ReadOnly = true
		noErr(t, presetConfirm(cb(admin, ""), "1"))
		settings.ReadOnly = false
		if rec.lastAnswer() != tr(admin, "settings_readonly") {
			t.Fatalf("confirm read-only: %q", rec.lastAnswer())
		}
		// Nil sets on a non-default entry.
		noErr(t, cmdPreset(msg(admin, "/preset low", "low")))
		pendingPresetMu.Lock()
		k := chatMsgKey(chat.ID, rec.lastID())
		e := pendingPresets[k]
		e.Sets = nil
		pendingPresets[k] = e
		pendingPresetMu.Unlock()
		noErr(t, presetConfirm(cb(admin, ""), "1"))
		if rec.lastAnswer() != tr(admin, "callback_unknown") {
			t.Fatalf("nil sets: %q", rec.lastAnswer())
		}
		// Apply.
		noErr(t, cmdPreset(msg(admin, "/preset cache 8 preload 20 conn 30", "")))
		noErr(t, presetConfirm(cb(admin, ""), "1"))
		if got := rec.lastSent(); got != tr(admin, "preset_applied")+"cache 8 preload 20 conn 30" {
			t.Fatalf("applied: %q", got)
		}
		if s := settings.BTsets; s.CacheSize != 8<<20 || s.PreloadCache != 20 || s.ConnectionsLimit != 30 {
			t.Fatalf("settings after preset: cache=%d preload=%d conn=%d", s.CacheSize, s.PreloadCache, s.ConnectionsLimit)
		}
	})

	t.Run("settings guards and pages", func(t *testing.T) {
		noErr(t, settingsCallback(cb(stranger, ""), "dht"))
		if rec.lastAnswer() != tr(stranger, "admin_only") {
			t.Fatalf("non-admin: %q", rec.lastAnswer())
		}
		old := settings.BTsets
		settings.BTsets = nil
		noErr(t, settingsCallback(cb(admin, ""), "dht"))
		settings.BTsets = old
		if rec.lastAnswer() != tr(admin, "settings_not_loaded") {
			t.Fatalf("not loaded: %q", rec.lastAnswer())
		}
		settings.ReadOnly = true
		noErr(t, settingsCallback(cb(admin, ""), "debug"))
		noErr(t, settingsCallback(cb(admin, ""), "torznab_clear"))
		noErr(t, settingsCallback(cb(admin, ""), "storage_set|json"))
		settings.ReadOnly = false
		if settings.BTsets.EnableDebug {
			t.Fatal("read-only toggle changed settings")
		}
		for _, bad := range []string{"nope", "nope|1", "bogus|5", "ask|nope"} {
			noErr(t, settingsCallback(cb(admin, ""), bad))
			if rec.lastAnswer() != tr(admin, "callback_unknown") {
				t.Fatalf("%q: %q", bad, rec.lastAnswer())
			}
		}

		// Export sends a document and confirms.
		noErr(t, settingsCallback(cb(admin, ""), "export"))
		if rec.lastAnswer() != tr(admin, "settings_exported") {
			t.Fatalf("export: %q", rec.lastAnswer())
		}

		// Prompts for every text setting.
		for _, s := range []struct{ name, hint string }{
			{"friendlyname", "settings_hint_friendlyname"}, {"torrentssavepath", "settings_hint_path"},
			{"sslcert", "settings_hint_sslcert"}, {"sslkey", "settings_hint_sslkey"},
			{"tmdbkey", "settings_hint_tmdbkey"}, {"torznab_add", "settings_hint_torznab"},
			{"torznab_test", "settings_hint_torznab_test"},
		} {
			noErr(t, settingsCallback(cb(admin, ""), "ask|"+s.name))
			if got := rec.lastSent(); !strings.Contains(got, tr(admin, s.hint)) {
				t.Fatalf("prompt %s: %q", s.name, got)
			}
		}
		noErr(t, settingsCallback(cb(admin, ""), "input_cancel"))

		// Reset confirmation, with and without a working edit.
		noErr(t, settingsCallback(cb(admin, ""), "reset_confirm"))
		rec.failEdit.Store(true)
		noErr(t, settingsCallback(cb(admin, ""), "reset_confirm"))
		if !strings.Contains(rec.lastSent(), tr(admin, "settings_reset_confirm")) {
			t.Fatalf("reset confirm fallback: %q", rec.lastSent())
		}
		for _, p := range []string{"1", "1a", "2c", "3", "4"} {
			noErr(t, settingsCallback(cb(admin, ""), "page|"+p))
		}
		noErr(t, settingsCallback(cb(admin, ""), "storage_set|json"))
		if rec.lastAnswer() != tr(admin, "settings_saved") {
			t.Fatalf("storage: %q", rec.lastAnswer())
		}
		rec.failEdit.Store(false)
		noErr(t, settingsCallback(cb(admin, ""), "page|2"))
		noErr(t, settingsCallback(cb(admin, ""), "more"))
	})

	t.Run("settings toggles and values", func(t *testing.T) {
		// Each applied change restarts the BT client (seconds), so keep this short.
		was := settings.BTsets.EnableDebug
		noErr(t, settingsCallback(cb(admin, ""), "debug|1b"))
		if settings.BTsets.EnableDebug == was || rec.lastAnswer() != tr(admin, "settings_saved") {
			t.Fatalf("debug stayed %v, answer %q", was, rec.lastAnswer())
		}
		settings.BTsets.EnableDebug = false

		noErr(t, settingsCallback(cb(admin, ""), "cache|128|2a"))
		if settings.BTsets.CacheSize != 128<<20 {
			t.Fatalf("cache = %d", settings.BTsets.CacheSize)
		}

		// Clearing Torznab indexers.
		settings.BTsets.TorznabUrls = []settings.TorznabConfig{{Host: "https://idx.example", Key: "k"}}
		noErr(t, settingsCallback(cb(admin, ""), "torznab_clear"))
		if len(settings.BTsets.TorznabUrls) != 0 || rec.lastAnswer() != tr(admin, "settings_saved") {
			t.Fatalf("torznab_clear: %+v", settings.BTsets.TorznabUrls)
		}
	})
}

// haltStatusLoop stops the refresh goroutine for msgID once it has registered,
// so it cannot read settings while a later subtest rewrites them.
func haltStatusLoop(t *testing.T, msgID int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		statusStopChansMu.Lock()
		_, ok := statusStopChans[msgID]
		statusStopChansMu.Unlock()
		if ok {
			stopStatusRefresh(msgID)
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("status loop for message %d never started", msgID)
}

func (r *callRecorder) lastEdited() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.edited) == 0 {
		return ""
	}
	return r.edited[len(r.edited)-1]
}

func TestApplyPresetKV(t *testing.T) {
	const uid = int64(42)
	usage := tr(uid, "preset_usage")
	cases := []struct {
		name  string
		args  []string
		ok    bool
		check func(*settings.BTSets) bool
	}{
		{"too short", []string{"cache"}, false, nil},
		{"unknown key", []string{"nope", "1"}, false, nil},
		{"zero cache", []string{"cache", "0"}, false, nil},
		{"cache", []string{"cache", "256"}, true, func(s *settings.BTSets) bool { return s.CacheSize == 256<<20 }},
		{"preload bounds", []string{"preload", "101"}, false, nil},
		{"preload", []string{"preload", "100"}, true, func(s *settings.BTSets) bool { return s.PreloadCache == 100 }},
		{"readahead low", []string{"readahead", "4"}, false, nil},
		{"readahead", []string{"readahead", "5"}, true, func(s *settings.BTSets) bool { return s.ReaderReadAHead == 5 }},
		{"conn", []string{"connections", "77"}, true, func(s *settings.BTSets) bool { return s.ConnectionsLimit == 77 }},
		{"timeout", []string{"timeout", "90"}, true, func(s *settings.BTSets) bool { return s.TorrentDisconnectTimeout == 90 }},
		{"port auto", []string{"port", "auto"}, true, func(s *settings.BTSets) bool { return s.PeersListenPort == 0 }},
		{"port privileged", []string{"port", "80"}, false, nil},
		{"port", []string{"port", "6881"}, true, func(s *settings.BTSets) bool { return s.PeersListenPort == 6881 }},
		{"down inf", []string{"down", "inf"}, true, func(s *settings.BTSets) bool { return s.DownloadRateLimit == 0 }},
		{"download", []string{"download", "2000"}, true, func(s *settings.BTSets) bool { return s.DownloadRateLimit == 2000 }},
		{"up inf", []string{"up", "∞"}, true, func(s *settings.BTSets) bool { return s.UploadRateLimit == 0 }},
		{"upload", []string{"upload", "300"}, true, func(s *settings.BTSets) bool { return s.UploadRateLimit == 300 }},
		{"retr off", []string{"retr", "off"}, true, func(s *settings.BTSets) bool { return s.RetrackersMode == 0 }},
		{"retr add", []string{"retr", "add"}, true, func(s *settings.BTSets) bool { return s.RetrackersMode == 1 }},
		{"retr rem", []string{"retrackers", "remove"}, true, func(s *settings.BTSets) bool { return s.RetrackersMode == 2 }},
		{"retr repl", []string{"retr", "replace"}, true, func(s *settings.BTSets) bool { return s.RetrackersMode == 3 }},
		{"retr number", []string{"retr", "2"}, true, func(s *settings.BTSets) bool { return s.RetrackersMode == 2 }},
		{"retr out of range", []string{"retr", "9"}, false, nil},
		{"responsive on", []string{"responsive", "ДА"}, true, func(s *settings.BTSets) bool { return s.ResponsiveMode }},
		{"responsive off", []string{"responsive", "off"}, true, func(s *settings.BTSets) bool { return !s.ResponsiveMode }},
		{"cachedrop", []string{"cachedrop", "yes"}, true, func(s *settings.BTSets) bool { return s.RemoveCacheOnDrop }},
		{"multiple, one valid", []string{"cache", "0", "conn", "12"}, true, func(s *settings.BTSets) bool { return s.ConnectionsLimit == 12 && s.CacheSize == 1 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &settings.BTSets{CacheSize: 1, ResponsiveMode: tc.name == "responsive off"}
			ok, errMsg := applyPresetKV(s, tc.args, uid)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if !ok && errMsg != usage {
				t.Fatalf("error %q, want usage", errMsg)
			}
			if ok && (errMsg != "" || !tc.check(s)) {
				t.Fatalf("settings %+v err %q", s, errMsg)
			}
		})
	}
}

func TestApplyNamedPreset(t *testing.T) {
	const uid = int64(42)
	for name, want := range map[string]int64{"perf": 512, "store": 64, "stream": 256, "minimal": 64} {
		s := &settings.BTSets{}
		ok, msg := applyNamedPreset(s, name, uid)
		if !ok || s.CacheSize != want<<20 || !strings.HasPrefix(msg, tr(uid, "preset_applied")) {
			t.Fatalf("%s: ok=%v cache=%d msg=%q", name, ok, s.CacheSize, msg)
		}
	}
	for _, name := range []string{"default", "nope"} {
		if ok, _ := applyNamedPreset(&settings.BTSets{}, name, uid); ok {
			t.Fatalf("%s applied", name)
		}
	}
}

func TestPendingPresetCleanupStops(t *testing.T) {
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() { pendingPresetCleanup(stop); close(done) }()
	close(stop)
	<-done
}
