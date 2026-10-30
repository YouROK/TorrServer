package tgbot

import (
	"net/http/httptest"
	"testing"

	tele "gopkg.in/telebot.v4"

	"server/settings"
	"server/tgbot/config"
	"server/torr"
	"server/torr/offlinetest"
)

// The reset and storage_view buttons used to be dead: the prefix test sliced one
// character too few, so the action never matched and nothing happened.
func TestSettingsCallbackResetAndStorageView(t *testing.T) {
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
	chat := &tele.Chat{ID: 42, Type: tele.ChatPrivate}
	cb := func(data string) tele.Context {
		return b.NewContext(tele.Update{Callback: &tele.Callback{
			ID: "1", Data: data, Sender: &tele.User{ID: 42},
			Message: &tele.Message{ID: 7, Chat: chat, Sender: &tele.User{ID: 1, IsBot: true}},
		}})
	}
	edits := func() int {
		rec.mu.Lock()
		defer rec.mu.Unlock()
		return len(rec.texts)
	}

	t.Run("reset no keeps settings and redraws the menu", func(t *testing.T) {
		settings.BTsets.FriendlyName = "keep me"
		before := edits()
		if err := settingsCallback(cb("\ffset|reset_def|0"), "reset_def|0"); err != nil {
			t.Fatal(err)
		}
		if settings.BTsets.FriendlyName != "keep me" {
			t.Fatalf("settings changed on No: %q", settings.BTsets.FriendlyName)
		}
		if edits() == before {
			t.Fatal("the menu was not redrawn: the callback did not match")
		}
	})

	t.Run("reset yes read-only changes nothing", func(t *testing.T) {
		settings.BTsets.FriendlyName = "keep me"
		settings.ReadOnly = true
		defer func() { settings.ReadOnly = false }()
		_ = settingsCallback(cb("\ffset|reset_def|1"), "reset_def|1")
		if settings.BTsets.FriendlyName != "keep me" {
			t.Fatalf("read-only reset changed settings: %q", settings.BTsets.FriendlyName)
		}
	})

	t.Run("reset yes restores defaults", func(t *testing.T) {
		settings.BTsets.FriendlyName = "custom"
		settings.BTsets.ConnectionsLimit = 99
		if err := settingsCallback(cb("\ffset|reset_def|1"), "reset_def|1"); err != nil {
			t.Fatal(err)
		}
		if settings.BTsets.FriendlyName != "" || settings.BTsets.ConnectionsLimit != 25 {
			t.Fatalf("not reset: FriendlyName=%q ConnectionsLimit=%d", settings.BTsets.FriendlyName, settings.BTsets.ConnectionsLimit)
		}
	})

	t.Run("storage_view switches the viewed backend", func(t *testing.T) {
		for _, pref := range []string{"json", "bbolt", "json"} {
			before := edits()
			if err := settingsCallback(cb("\ffset|storage_view|"+pref), "storage_view|"+pref); err != nil {
				t.Fatal(err)
			}
			if got := settings.BTsets.StoreViewedInJson; got != (pref == "json") {
				t.Fatalf("after %s: StoreViewedInJson=%v", pref, got)
			}
			if edits() == before {
				t.Fatalf("%s: the menu was not redrawn", pref)
			}
		}
	})

	t.Run("storage_view read-only is refused", func(t *testing.T) {
		want := settings.BTsets.StoreViewedInJson
		settings.ReadOnly = true
		defer func() { settings.ReadOnly = false }()
		other := "json"
		if want {
			other = "bbolt"
		}
		_ = settingsCallback(cb("\ffset|storage_view|"+other), "storage_view|"+other)
		if settings.BTsets.StoreViewedInJson != want {
			t.Fatal("read-only storage_view changed the backend")
		}
	})
}
