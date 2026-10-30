package tgbot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tele "gopkg.in/telebot.v4"

	"server/settings"
	"server/tgbot/config"
)

const testHash = "2f6da5bb5f1ea1994be6814ae3cbcc8bdf911eaf"

func offlineCtx(t *testing.T, msg *tele.Message) tele.Context {
	t.Helper()
	b, err := tele.NewBot(tele.Settings{Token: "123:test", Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	return b.NewContext(tele.Update{Message: msg})
}

func replyWith(data ...string) *tele.Message {
	var row []tele.InlineButton
	for _, d := range data {
		row = append(row, tele.InlineButton{Data: d})
	}
	return &tele.Message{ReplyTo: &tele.Message{ReplyMarkup: &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{row}}}}
}

func TestExtractHashFromReply(t *testing.T) {
	tests := []struct {
		name string
		msg  *tele.Message
		want string
	}{
		{"no message", nil, ""},
		{"not a reply", &tele.Message{}, ""},
		{"reply without markup", &tele.Message{ReplyTo: &tele.Message{}}, ""},
		{"empty keyboard", &tele.Message{ReplyTo: &tele.Message{ReplyMarkup: &tele.ReplyMarkup{}}}, ""},
		{"plain hash button", replyWith("", "noise", testHash), testHash},
		{"all| prefix", replyWith("files|all|" + testHash), testHash},
		{"all| with suffix", replyWith("x|all|" + testHash + "|2"), testHash},
		{"all| without a hash", replyWith("x|all|nothash"), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractHashFromReply(offlineCtx(t, tt.msg)); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLogHashOrTruncate(t *testing.T) {
	long := strings.Repeat("a", 80)
	tests := []struct{ name, in, want string }{
		{"hash", testHash, testHash},
		{"magnet", "magnet:?xt=urn:btih:" + testHash + "&dn=x", testHash},
		{"torrs", "torrs://" + testHash + "rest", testHash},
		{"short link", "http://x/y.torrent", "http://x/y.torrent"},
		{"long link", "http://" + long, ("http://" + long)[:64] + "..."},
		{"magnet with bad hash", "magnet:?xt=urn:btih:" + strings.Repeat("z", 40), "magnet:?xt=urn:btih:" + strings.Repeat("z", 40)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := logHashOrTruncate(tt.in); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRetrackersStr(t *testing.T) {
	for mode, want := range map[int]string{0: "off", 1: "add", 2: "remove", 3: "replace", 4: "?", -1: "?"} {
		if got := retrackersStr(mode); got != want {
			t.Errorf("retrackersStr(%d) = %q, want %q", mode, got, want)
		}
	}
}

func TestCategoryLabel(t *testing.T) {
	const uid = 42
	for _, cat := range []string{"movie", "tv", "music", "other", ""} {
		if got := categoryLabel(uid, cat); got == "" {
			t.Errorf("categoryLabel(%q) is empty", cat)
		}
	}
	if got := categoryLabel(uid, "audiobooks"); got != "audiobooks" {
		t.Errorf("custom category = %q, want it passed through", got)
	}
	if categoryLabel(uid, "") == categoryLabel(uid, "movie") {
		t.Error("uncategorized and movie share a label")
	}
}

func TestLoadUserLangs(t *testing.T) {
	oldPath := settings.Path
	userLangMu.Lock()
	old := userLang
	userLang = map[int64]string{}
	userLangMu.Unlock()
	t.Cleanup(func() {
		settings.Path = oldPath
		userLangMu.Lock()
		userLang = old
		userLangMu.Unlock()
	})
	settings.Path = t.TempDir()
	fn := filepath.Join(settings.Path, "tg_langs.json")

	loadUserLangs() // missing file: no panic, nothing loaded
	if len(userLang) != 0 {
		t.Fatalf("loaded from nothing: %v", userLang)
	}

	if err := os.WriteFile(fn, []byte(`{not json`), 0o600); err != nil {
		t.Fatal(err)
	}
	loadUserLangs()
	if len(userLang) != 0 {
		t.Fatalf("loaded corrupt file: %v", userLang)
	}

	body := `{"10":"` + LangRU + `","11":"` + LangEN + `","12":"xx","abc":"` + LangRU + `"}`
	if err := os.WriteFile(fn, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	loadUserLangs()
	userLangMu.RLock()
	defer userLangMu.RUnlock()
	if len(userLang) != 2 || userLang[10] != LangRU || userLang[11] != LangEN {
		t.Fatalf("userLang = %v, want only valid ids and languages", userLang)
	}
}

func TestNewTelegramHTTPClient(t *testing.T) {
	old := config.Cfg
	t.Cleanup(func() { config.Cfg = old })
	for name, socks := range map[string]string{
		"direct":          "",
		"blank":           "   ",
		"socks5":          "127.0.0.1:1080",
		"socks5 with url": "socks5://user:pw@127.0.0.1:1080",
		"wrong scheme":    "http://127.0.0.1:3128",
		"empty host":      "socks5://",
		"unparseable":     "socks5://[::1",
	} {
		t.Run(name, func(t *testing.T) {
			config.Cfg = &config.Config{Socks5: socks}
			c := newTelegramHTTPClient()
			if c == nil || c.Timeout.Minutes() != 5 {
				t.Fatalf("client %+v", c)
			}
			if _, ok := c.Transport.(copyTextFixTransport); !ok {
				t.Fatalf("transport %T is not wrapped", c.Transport)
			}
		})
	}
}
