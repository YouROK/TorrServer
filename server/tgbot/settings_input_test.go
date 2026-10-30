package tgbot

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	tele "gopkg.in/telebot.v4"

	"server/settings"
	"server/tgbot/config"
	"server/torr"
	"server/torr/offlinetest"
)

// sendRecorder is a Telegram API stub that keeps the text of every send.
type sendRecorder struct {
	mu    sync.Mutex
	texts []string
	id    int
}

func (s *sendRecorder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	method := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	switch method {
	case "sendMessage", "editMessageText":
		var req struct {
			Text string `json:"text"`
		}
		_ = json.Unmarshal(body, &req)
		s.mu.Lock()
		s.texts = append(s.texts, req.Text)
		s.id++
		id := s.id
		s.mu.Unlock()
		writeJSON(w, map[string]any{"ok": true, "result": map[string]any{
			"message_id": id, "text": req.Text,
			"chat": map[string]any{"id": 42, "type": "private"},
			"from": map[string]any{"id": 1, "is_bot": true, "first_name": "T"},
		}})
	default:
		writeJSON(w, map[string]any{"ok": true, "result": true})
	}
}

func (s *sendRecorder) last() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.texts) == 0 {
		return ""
	}
	return s.texts[len(s.texts)-1]
}

func TestSettingsInputReply(t *testing.T) {
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

	rec := &sendRecorder{}
	srv := httptest.NewServer(rec)
	t.Cleanup(srv.Close)
	oldCfg := config.Cfg
	config.Cfg = &config.Config{HostTG: srv.URL, WhiteIds: []int64{42}}
	t.Cleanup(func() { config.Cfg = oldCfg })

	b, err := tele.NewBot(tele.Settings{URL: srv.URL, Token: "123:test", Offline: true, Client: srv.Client()})
	if err != nil {
		t.Fatal(err)
	}
	const admin, stranger = int64(42), int64(7)
	chat := &tele.Chat{ID: 42, Type: tele.ChatPrivate}
	promptID := 100
	// reply builds a text reply to the prompt message, sent by uid.
	reply := func(uid int64, text string) tele.Context {
		return b.NewContext(tele.Update{Message: &tele.Message{
			Text: text, Sender: &tele.User{ID: uid}, Chat: chat,
			ReplyTo: &tele.Message{ID: promptID, Chat: chat},
		}})
	}
	// expect registers a pending prompt for setting and feeds it a reply.
	expect := func(uid int64, setting, text string) bool {
		promptID++
		pendingInputMu.Lock()
		pendingInputs[chatMsgKey(chat.ID, promptID)] = pendingInput{Setting: setting, UserID: admin, CreatedAt: time.Now()}
		pendingInputMu.Unlock()
		return handleSettingsInputReply(reply(uid, text))
	}

	t.Run("not a reply", func(t *testing.T) {
		c := b.NewContext(tele.Update{Message: &tele.Message{Text: "x", Sender: &tele.User{ID: admin}, Chat: chat}})
		if handleSettingsInputReply(c) {
			t.Fatal("plain message handled")
		}
	})
	t.Run("unknown prompt", func(t *testing.T) {
		if handleSettingsInputReply(reply(admin, "x")) {
			t.Fatal("reply to an unknown message handled")
		}
	})
	t.Run("someone else's prompt", func(t *testing.T) {
		if expect(stranger, "friendlyname", "x") {
			t.Fatal("handled a reply from another user")
		}
	})
	t.Run("expired", func(t *testing.T) {
		promptID++
		pendingInputMu.Lock()
		pendingInputs[chatMsgKey(chat.ID, promptID)] = pendingInput{Setting: "friendlyname", UserID: admin, CreatedAt: time.Now().Add(-2 * pendingInputTTL)}
		pendingInputMu.Unlock()
		if !handleSettingsInputReply(reply(admin, "late")) || rec.last() != tr(admin, "canceled") {
			t.Fatalf("expired: last=%q", rec.last())
		}
	})
	t.Run("pending is consumed once", func(t *testing.T) {
		promptID++
		pendingInputMu.Lock()
		pendingInputs[chatMsgKey(chat.ID, promptID)] = pendingInput{Setting: "nope", UserID: admin, CreatedAt: time.Now()}
		pendingInputMu.Unlock()
		if !handleSettingsInputReply(reply(admin, "x")) {
			t.Fatal("first reply not handled")
		}
		if handleSettingsInputReply(reply(admin, "x")) {
			t.Fatal("second reply to the same prompt handled")
		}
	})
	t.Run("guards", func(t *testing.T) {
		c := reply(stranger, "x")
		applySettingsInput(c, "friendlyname", "x")
		if rec.last() != tr(stranger, "admin_only") {
			t.Fatalf("non-admin: %q", rec.last())
		}
		settings.ReadOnly = true
		applySettingsInput(reply(admin, "x"), "friendlyname", "x")
		settings.ReadOnly = false
		if rec.last() != tr(admin, "settings_readonly") {
			t.Fatalf("read-only: %q", rec.last())
		}
		old := settings.BTsets
		settings.BTsets = nil
		applySettingsInput(reply(admin, "x"), "friendlyname", "x")
		settings.BTsets = old
		if rec.last() != tr(admin, "settings_not_loaded") {
			t.Fatalf("not loaded: %q", rec.last())
		}
	})
	t.Run("unknown setting", func(t *testing.T) {
		applySettingsInput(reply(admin, "x"), "nope", "x")
		if rec.last() != tr(admin, "callback_unknown") {
			t.Fatalf("unknown: %q", rec.last())
		}
	})
	t.Run("torznab usage and test", func(t *testing.T) {
		applySettingsInput(reply(admin, ""), "torznab_add", "")
		if rec.last() != tr(admin, "settings_input_torznab_usage") {
			t.Fatalf("add usage: %q", rec.last())
		}
		applySettingsInput(reply(admin, ""), "torznab_test", "")
		if rec.last() != tr(admin, "settings_input_torznab_usage") {
			t.Fatalf("test usage: %q", rec.last())
		}
		idx := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`<caps><limits max="1" default="1"/></caps>`))
		}))
		defer idx.Close()
		applySettingsInput(reply(admin, ""), "torznab_test", idx.URL+"|key")
		if rec.last() != tr(admin, "settings_torznab_test_ok") {
			t.Fatalf("test ok: %q", rec.last())
		}
		applySettingsInput(reply(admin, ""), "torznab_test", "http://127.0.0.1:1|key")
		if prefix := strings.TrimSuffix(tr(admin, "settings_torznab_test_fail"), "%s"); !strings.HasPrefix(rec.last(), prefix) || rec.last() == prefix {
			t.Fatalf("test fail: %q", rec.last())
		}
	})
	t.Run("applies and clears settings", func(t *testing.T) {
		if !expect(admin, "friendlyname", "  My Server ") {
			t.Fatal("friendlyname not handled")
		}
		if got := settings.BTsets.FriendlyName; got != "My Server" {
			t.Fatalf("FriendlyName = %q", got)
		}
		if !expect(admin, "friendlyname", "-") || settings.BTsets.FriendlyName != "" {
			t.Fatalf("clear: %q", settings.BTsets.FriendlyName)
		}
		expect(admin, "torznab_add", "idx.example|k1|Idx|2000,5000")
		urls := settings.BTsets.TorznabUrls
		if len(urls) != 1 || urls[0].Host != "https://idx.example" || urls[0].Key != "k1" ||
			urls[0].Name != "Idx" || urls[0].Categories != "2000,5000" || urls[0].CatType != settings.CategoryDefault {
			t.Fatalf("TorznabUrls = %+v", urls)
		}
		expect(admin, "tmdbkey", "abc")
		if settings.BTsets.TMDBSettings.APIKey != "abc" {
			t.Fatalf("tmdb key %q", settings.BTsets.TMDBSettings.APIKey)
		}
	})
}

func TestValueOrClear(t *testing.T) {
	if valueOrClear("") != "(cleared)" || valueOrClear("short") != "short" {
		t.Fatal("short values")
	}
	long := strings.Repeat("x", 60)
	if got := valueOrClear(long); got != long[:47]+"..." {
		t.Fatalf("long value %q", got)
	}
}
