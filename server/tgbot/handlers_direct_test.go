package tgbot

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/anacrolix/torrent"
	tele "gopkg.in/telebot.v4"
	ffp "gopkg.in/vansante/go-ffprobe.v2"

	"server/rutor/models"
	"server/settings"
	"server/tgbot/config"
	"server/torr"
	"server/torr/offlinetest"
)

func TestHandlersDirect(t *testing.T) {
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

	fileSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = show.Meta.Write(w)
	}))
	t.Cleanup(fileSrv.Close)

	api := &tgAPI{idle: make(chan struct{}, 1)}
	srv := httptest.NewServer(api)
	t.Cleanup(srv.Close)
	config.Cfg = &config.Config{HostTG: srv.URL, WhiteIds: []int64{42}, BlackIds: []int64{}}

	b, err := tele.NewBot(tele.Settings{URL: srv.URL, Token: "123:test", Offline: true, Client: srv.Client()})
	if err != nil {
		t.Fatal(err)
	}
	user := &tele.User{ID: 42, FirstName: "P"}
	chat := &tele.Chat{ID: 42, Type: tele.ChatPrivate}
	msg := func(text, payload string) tele.Context {
		return b.NewContext(tele.Update{Message: &tele.Message{Text: text, Payload: payload, Sender: user, Chat: chat}})
	}
	cb := func(data string) tele.Context {
		api.mu.Lock()
		id := api.msgID
		api.mu.Unlock()
		return b.NewContext(tele.Update{Callback: &tele.Callback{
			ID: "1", Data: data, Sender: user,
			Message: &tele.Message{ID: id, Chat: chat, Sender: &tele.User{ID: 1, IsBot: true, FirstName: "T"}},
		}})
	}
	call := func(err error) {
		t.Helper()
		if err != nil {
			t.Log(err)
		}
	}

	t.Cleanup(func() {
		for id := 1; id < 300; id++ {
			stopStatusRefresh(id)
			stopSnakeRefresh(id)
		}
	})

	call(list(msg("/list", "")))
	call(callbackListPage(cb("\fflist|0"), "0"))
	call(callbackListRefresh(cb("\ffrefresh|0"), "0"))
	call(callbackListCat(cb("\fflistcat|tv"), "tv"))
	call(callbackTorrentPick(cb("\fftpick|"+hash+"|0"), hash, "0"))
	call(callbackBackList(cb("\ffbacklist|0"), "0"))
	call(files(cb("\ffiles|" + hash)))
	call(callbackFileListPage(cb("\ffitems|0|"+hash), "0", hash))
	call(callbackFileListRefresh(cb("\ffifresh|0|"+hash), "0", hash))
	call(cmdStatus(msg("/status "+hash, hash)))
	call(callbackStatus(cb("\ffstatus|"+hash), hash))
	call(callbackStatusStop(cb("\ffstatusstop|"+hash), hash))
	call(callbackStatusAllPage(cb("\ffstatusall|0"), "0"))
	call(callbackStatusAllRefresh(cb("\ffstatusallrefresh|0"), "0"))
	call(cmdHash(msg("/hash", "")))
	call(cmdHash(msg("/hash 1", "1")))
	call(callbackHashPage(cb("\ffhash|0"), "0"))
	call(callbackHashRefresh(cb("\ffhashrefresh|0"), "0"))
	call(cmdCache(msg("/cache "+hash, hash)))
	call(cmdM3u(msg("/m3u "+hash, hash)))
	call(cmdM3uAll(msg("/m3uall", "")))
	call(cmdLink(msg("/link "+hash+" 1", hash+" 1")))
	call(callbackLink(cb("\fflink|"+hash+"|1"), hash+"|1"))
	call(cmdViewed(msg("/viewed", "")))
	call(cmdViewed(msg("/viewed set "+hash+" 1", "set "+hash+" 1")))
	call(cmdViewed(msg("/viewed rem "+hash+" 1", "rem "+hash+" 1")))
	call(cmdNext(msg("/next Show", "Show")))
	call(callbackNextMark(cb("\ffnextmark|"+hash+"|1"), hash, "1"))
	call(cmdSet(msg("/set "+hash+" Renamed", hash+" Renamed")))
	call(cmdSetCat(msg("/setcat "+hash, hash)))
	call(callbackSetCat(cb("\ffsetcat|"+hash+"|movie"), hash, "movie"))
	call(cmdAdd(msg("/add", "")))
	if tor, err := addTorrent(msg("/add "+fileSrv.URL, fileSrv.URL), fileSrv.URL+"/a.torrent"); err != nil {
		t.Log(err)
	} else if tor != nil {
		call(afterAdd(msg("", ""), tor))
	}
	call(cmdExport(msg("/export", "")))
	call(callbackExportPage(cb("\ffexport|0"), "0"))
	call(callbackExportRefresh(cb("\ffexportrefresh|0"), "0"))
	call(cmdDb(msg("/db", "")))
	call(callbackDbPage(cb("\ffdb|0"), "0"))
	call(callbackDbRefresh(cb("\ffdbrefresh|0"), "0"))
	call(cmdImport(msg("/import", "")))
	call(cmdEcho(msg("/echo hi", "hi")))
	call(cmdServer(msg("/server", "")))
	call(cmdStats(msg("/stats", "")))
	call(cmdStat(msg("/stat", "")))
	call(cmdCategories(msg("/categories", "")))
	call(cmdLang(msg("/lang en", "en")))
	call(cmdSnake(msg("/snake "+hash, hash)))
	call(callbackSnakeRefresh(cb("\ffsnakerefresh|"+hash), hash))
	call(callbackSnakeStop(cb("\ffsnakestop|"+hash), hash))
	for _, act := range []string{"clear", "hash", "categories", "export", "import", "snake", "preload", "next", "queue", "ffp", "speedtest", "echo", "db", "viewed", "server", "stats", "stat", "m3uall", "cache", "help", "lang", "settings", "preset", "shutdown", "nope"} {
		call(callbackMenuAct(cb("\ffmenu|act|"+act), act))
	}
	call(callbackMenu(cb("\ffmenu|hub|lib"), []string{"hub", "lib"}))
	call(callbackMenu(cb("\ffmenu|root"), []string{"root"}))
	call(handleInlineQuery(b.NewContext(tele.Update{Query: &tele.Query{ID: "q", Text: "list", Sender: user}})))
	call(handleInlineQuery(b.NewContext(tele.Update{Query: &tele.Query{ID: "q2", Text: "", Sender: user}})))
	call(handleInlineChosen(b.NewContext(tele.Update{InlineResult: &tele.InlineResult{ResultID: "nope", Sender: user}})))

	for _, page := range []string{"1", "1a", "1b", "1c", "2", "2a", "2b", "2c", "3", "4"} {
		call(settingsCallback(cb("\ffset|page|"+page), "page|"+page))
	}
	call(settingsCallback(cb("\ffset|ask|friendlyname"), "ask|friendlyname"))
	call(settingsCallback(cb("\ffset|ask|nope"), "ask|nope"))
	call(settingsCallback(cb("\ffset|export"), "export"))
	call(settingsCallback(cb("\ffset|more"), "more"))
	call(settingsCallback(cb("\ffset|reset_confirm"), "reset_confirm"))
	call(settingsCallback(cb("\ffset|reset_def|0"), "reset_def|0"))
	call(settingsCallback(cb("\ffset|storage_set|json"), "storage_set|json"))
	call(settingsCallback(cb("\ffset|storage_view|bbolt"), "storage_view|bbolt"))
	call(settingsCallback(cb("\ffset|input_cancel"), "input_cancel"))

	call(cmdPreset(msg("/preset", "")))
	call(cmdPreset(msg("/preset performance", "performance")))
	call(presetConfirm(cb("\ffpreset|0"), "0"))
	call(cmdPreset(msg("/preset storage", "storage")))
	call(presetConfirm(cb("\ffpreset|0"), "0"))
	call(cmdPreset(msg("/preset streaming", "streaming")))
	call(presetConfirm(cb("\ffpreset|0"), "0"))
	call(cmdPreset(msg("/preset low", "low")))
	call(presetConfirm(cb("\ffpreset|0"), "0"))
	call(cmdPreset(msg("/preset cache 8 preload 20 conn 30", "cache 8 preload 20 conn 30")))
	call(presetConfirm(cb("\ffpreset|0"), "0"))
	call(cmdPreset(msg("/preset default", "default")))
	call(presetConfirm(cb("\ffpreset|0"), "0"))

	_, _, _ = ParseRange(42, "1-2")
	_, _, _ = ParseRange(42, "bad")
	_ = helpAllText(42)
	_ = helpCompactText(42)
	if formatFfpHuman(&ffp.ProbeData{
		Format: &ffp.Format{FormatLongName: "matroska", DurationSeconds: 3661, Size: "1024", BitRate: "128000"},
		Streams: []*ffp.Stream{
			{CodecType: "video", CodecLongName: "h264", Profile: "High", Width: 16, Height: 16, DisplayAspectRatio: "1:1", PixFmt: "yuv420p", RFrameRate: "24/1", BitRate: "1000", ColorSpace: "bt709", ColorTransfer: "bt709", ColorPrimaries: "bt709", TagList: ffp.Tags{"title": "v", "language": "en"}},
			{CodecType: "audio", CodecLongName: "aac", Profile: "LC", SampleRate: "48000", Channels: 2, ChannelLayout: "stereo", BitRate: "128000", TagList: ffp.Tags{"title": "a", "language": "ru"}},
			{CodecType: "subtitle", CodecLongName: "subrip", TagList: ffp.Tags{"title": "s", "language-eng": "en"}},
			{CodecType: "data", CodecLongName: "bin"},
		},
	}, 42) == "" {
		t.Fatal("ffp format")
	}
	_ = getTag(nil, "title")
	_ = getTag(ffp.Tags{"title": 1, "language-en": "en"}, "language")

	for _, data := range []string{
		"\ffexport|0", "\ffexportrefresh|0", "\ffhash|0", "\ffhashrefresh|0",
		"\ffstatusall|0", "\ffstatusallrefresh|0", "\ffdb|0", "\ffdbrefresh|0",
		"\ffadd|0", "\ffmore|10", "\fflistcat|tv", "\ffsetcat|" + hash + "|tv",
		"\ffnextmark|" + hash + "|1", "\ffitems|0|" + hash, "\ffifresh|0|" + hash,
		"\ffbacklist|0", "\ffnop",
		"\ffsnakerefresh|" + hash, "\ffsnakestop|" + hash,
		"\ffclear|0", "\ffshutdown|0", "\ffpreset|0", "\ffset|page|2",
	} {
		started := time.Now()
		call(handleCallback(cb(data)))
		if d := time.Since(started); d > 500*time.Millisecond {
			t.Logf("slow callback %q %s", data, d)
		}
	}
	call(handleMenuButton(msg("library", ""), "library"))
	_ = listFilterKeyboard(42, "tv")
	_ = humanizeSpeedBits(42, 1024)
	_ = torrentsByHashPrefix(hash[:4])
	_ = miniAppURL()
	call(sendSettingsMenuPage(msg("/settings", ""), 42, "2"))
	call(startSnakeForHash(msg("/snake "+hash, hash), hash))
	stopSnakeRefresh(api.msgID)

	origR, origT := rutorSearch, torznabSearch
	t.Cleanup(func() { rutorSearch, torznabSearch = origR, origT })
	rutorSearch = func(q string) []*models.TorrentDetails {
		return []*models.TorrentDetails{{Title: q, Magnet: "magnet:?xt=urn:btih:" + strings.Repeat("a", 40), Size: "1", Seed: 2, Peer: 1}}
	}
	torznabSearch = func(context.Context, string, int, string, int, int) []*models.TorrentDetails {
		return []*models.TorrentDetails{{Title: "tz", Link: "http://127.0.0.1/t", Size: "2", Seed: 1}}
	}
	settings.BTsets.EnableRutorSearch = true
	settings.BTsets.EnableTorznabSearch = true
	hits := []*models.TorrentDetails{{Title: "nolink"}, {Title: "hit", Magnet: "magnet:?xt=urn:btih:" + strings.Repeat("c", 40), Size: "3", Seed: 1, Peer: 1}}
	for i := 0; i < 12; i++ {
		hits = append(hits, &models.TorrentDetails{Title: "p", Magnet: "magnet:?xt=urn:btih:" + strings.Repeat("d", 40)})
	}
	call(sendSearchResultsToRecipient(b, user, 42, 0, hits, "RuTor"))
	call(callbackSearchMore(cb("\ffmore|10"), "10"))
	call(callbackSearchAdd(cb("\ffadd|0"), "0"))
	call(callbackSearchAdd(cb("\ffadd|x"), "x"))
	call(runSearchQuery(msg("/search show", "show"), "show"))
	call(cmdSearchRutor(msg("/rutor show", "show")))
	call(cmdTorznab(msg("/torznab show 0", "show 0")))
	call(handleInlineQuery(b.NewContext(tele.Update{Query: &tele.Query{ID: "qs", Text: "show", ChatType: "sender", Sender: user}})))
	time.Sleep(200 * time.Millisecond)
	settings.BTsets.EnableRutorSearch = false
	settings.BTsets.EnableTorznabSearch = false

	for id := 1; id < 300; id++ {
		stopStatusRefresh(id)
		stopSnakeRefresh(id)
	}
	origStart := startRutor
	startRutor = func() {}
	t.Cleanup(func() { startRutor = origStart })
	applySettingsInput(msg("n", ""), "friendlyname", "Box")
	applySettingsInput(msg("n", ""), "sslcert", "")
	applySettingsInput(msg("n", ""), "tmdbkey", strings.Repeat("k", 60))
	applySettingsInput(msg("n", ""), "torznab_add", "http://127.0.0.1/tz|key|name|1000|movies")
	applySettingsInput(msg("n", ""), "torznab_test", "")
	applySettingsInput(msg("n", ""), "nope", "x")

	call(cmdDrop(msg("/drop "+hash, hash)))
	call(callbackDrop(cb("\ffdrop|"+hash), hash))
	call(cmdRemove(msg("/remove", "")))
	call(cmdRemove(msg("/remove "+hash, hash)))
	deleteTorrent(cb("\fdelete|" + hash))
	call(clearConfirm(cb("\ffclear|0"), "0"))
	call(shutdownConfirm(cb("\ffshutdown|0"), "0"))
}
