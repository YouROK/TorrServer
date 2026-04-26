package tgbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v4"
	"server/dlna"
	"server/rutor"
	"server/settings"
	"server/torr"
)

func cmdSettings(c tele.Context) error {
	uid := c.Sender().ID
	if settings.BTsets == nil {
		return c.Send(tr(uid, "settings_not_loaded"))
	}
	return sendSettingsMenu(c, uid)
}

func sendSettingsMenu(c tele.Context, uid int64) error {
	return sendSettingsMenuPage(c, uid, "1")
}

func sendSettingsMenuPage(c tele.Context, uid int64, page string) error {
	msg := sendSettingsMenuText(c, uid, page)
	kbd := sendSettingsMenuKbd(uid, page)
	return c.Send(msg, kbd)
}

func sendSettingsMenuText(c tele.Context, uid int64, page string) string {
	s := settings.BTsets
	msg := "⚙️ <b>" + tr(uid, "settings_title") + "</b>"
	switch page {
	case "1":
		msg += "\n\n"
		msg += fmt.Sprintf("🔍 %s: RuTor %s · Torznab %s\n", tr(uid, "settings_section_search"), boolIcon(s.EnableRutorSearch), boolIcon(s.EnableTorznabSearch))
		msg += fmt.Sprintf("📺 %s: DLNA %s · IPv6 %s · DHT %s · PEX %s · TCP %s · UTP %s\n", tr(uid, "settings_section_network"), boolIcon(s.EnableDLNA), boolIcon(s.EnableIPv6), boolIcon(!s.DisableDHT), boolIcon(!s.DisablePEX), boolIcon(!s.DisableTCP), boolIcon(!s.DisableUTP))
		msg += fmt.Sprintf("📦 %s: CacheDrop %s · Proxy %s · UseDisk %s\n", tr(uid, "settings_section_other"), boolIcon(s.RemoveCacheOnDrop), boolIcon(s.EnableProxy), boolIcon(s.UseDisk))
	case "1a":
		msg += " — " + tr(uid, "settings_section_search")
		msg += "\n\n"
		msg += fmt.Sprintf("RuTor %s · Torznab %s", boolIcon(s.EnableRutorSearch), boolIcon(s.EnableTorznabSearch))
	case "1b":
		msg += " — " + tr(uid, "settings_section_network")
		msg += "\n\n"
		msg += fmt.Sprintf("DLNA %s · IPv6 %s · Upload %s · DHT %s · PEX %s\n", boolIcon(s.EnableDLNA), boolIcon(s.EnableIPv6), boolIcon(!s.DisableUpload), boolIcon(!s.DisableDHT), boolIcon(!s.DisablePEX))
		msg += fmt.Sprintf("TCP %s · UTP %s · UPNP %s · Encrypt %s · Debug %s", boolIcon(!s.DisableTCP), boolIcon(!s.DisableUTP), boolIcon(!s.DisableUPNP), boolIcon(s.ForceEncrypt), boolIcon(s.EnableDebug))
	case "1c":
		msg += " — " + tr(uid, "settings_section_other")
		msg += "\n\n"
		msg += fmt.Sprintf("CacheDrop %s · Responsive %s · Proxy %s · UseDisk %s · FSActive %s", boolIcon(s.RemoveCacheOnDrop), boolIcon(s.ResponsiveMode), boolIcon(s.EnableProxy), boolIcon(s.UseDisk), boolIcon(s.ShowFSActiveTorr))
	case "2":
		msg += " — " + tr(uid, "settings_page2")
		msg += "\n\n"
		msg += fmt.Sprintf("💾 %s: %d MB · Preload %d%% · ReadAhead %d%%\n", tr(uid, "settings_limits_cache"), s.CacheSize/(1024*1024), s.PreloadCache, s.ReaderReadAHead)
		msg += fmt.Sprintf("🔌 %s: %d · Port %s · Timeout %ds\n", tr(uid, "settings_limits_connections"), s.ConnectionsLimit, portStr(s.PeersListenPort), s.TorrentDisconnectTimeout)
		msg += fmt.Sprintf("⬇️ %s: Down %s · Up %s · Retr %s\n", tr(uid, "settings_limits_speed"), rateStr(s.DownloadRateLimit), rateStr(s.UploadRateLimit), retrackersStr(s.RetrackersMode))
	case "2a":
		msg += " — " + tr(uid, "settings_page2") + " · " + tr(uid, "settings_limits_cache")
		msg += "\n\n"
		msg += fmt.Sprintf("Cache %d MB · Preload %d%% · ReadAhead %d%%", s.CacheSize/(1024*1024), s.PreloadCache, s.ReaderReadAHead)
	case "2b":
		msg += " — " + tr(uid, "settings_page2") + " · " + tr(uid, "settings_limits_connections")
		msg += "\n\n"
		msg += fmt.Sprintf("Connections %d · Port %s · Timeout %ds", s.ConnectionsLimit, portStr(s.PeersListenPort), s.TorrentDisconnectTimeout)
	case "2c":
		msg += " — " + tr(uid, "settings_page2") + " · " + tr(uid, "settings_limits_speed")
		msg += "\n\n"
		msg += fmt.Sprintf("Down %s · Up %s · Retrackers %s", rateStr(s.DownloadRateLimit), rateStr(s.UploadRateLimit), retrackersStr(s.RetrackersMode))
	case "3":
		msg += " — " + tr(uid, "settings_page3")
		msg += "\n\n"
		msg += fmt.Sprintf("📺 DLNA: %s · 💾 Path: %s\n", maskStr(s.FriendlyName, 25), maskVal(s.TorrentsSavePath))
		msg += fmt.Sprintf("🔐 SSL: %s · 🔑 TMDB: %s · Torznab: %d\n", maskVal(s.SslCert), maskVal(s.TMDBSettings.APIKey), len(s.TorznabUrls))
		msg += fmt.Sprintf("🌐 Proxy: %s", maskStr(strings.Join(s.ProxyHosts, ", "), 35))
	case "4":
		msg += " — " + tr(uid, "settings_page4")
		msg += "\n\n"
		msg += fmt.Sprintf("📄 %s: %s · 📺 %s: %s\n", tr(uid, "settings_storage_settings"), storageType(s.StoreSettingsInJson), tr(uid, "settings_storage_viewed"), storageType(s.StoreViewedInJson))
		msg += fmt.Sprintf("🔑 TMDB: %s · 🖼 URL: %s", maskVal(s.TMDBSettings.APIKey), maskStr(s.TMDBSettings.ImageURL, 20))
	}
	return msg
}

func storageType(useJSON bool) string {
	if useJSON {
		return "json"
	}
	return "bbolt"
}

func rateStr(kb int) string {
	if kb == 0 {
		return "∞"
	}
	return fmt.Sprintf("%d", kb)
}

func portStr(port int) string {
	if port == 0 {
		return "auto"
	}
	return fmt.Sprintf("%d", port)
}

func maskStr(s string, maxLen int) string {
	if s == "" {
		return "—"
	}
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

func maskVal(s string) string {
	if s == "" {
		return "—"
	}
	return "***"
}

func retrackersStr(mode int) string {
	switch mode {
	case 0:
		return "off"
	case 1:
		return "add"
	case 2:
		return "remove"
	case 3:
		return "replace"
	default:
		return "?"
	}
}

func sendSettingsMenuKbd(uid int64, page string) *tele.ReplyMarkup {
	s := settings.BTsets
	var btns [][]tele.InlineButton

	switch page {
	case "1":
		btns = [][]tele.InlineButton{
			{
				{Text: "🔍 " + tr(uid, "settings_section_search"), Unique: "fset", Data: "page|1a"},
				{Text: "📺 " + tr(uid, "settings_section_network"), Unique: "fset", Data: "page|1b"},
				{Text: "📦 " + tr(uid, "settings_section_other"), Unique: "fset", Data: "page|1c"},
			},
			{
				{Text: "📥 " + tr(uid, "settings_export"), Unique: "fset", Data: "export"},
				{Text: "📊 " + tr(uid, "settings_nav_cache"), Unique: "fset", Data: "page|2"},
				{Text: "✏️ " + tr(uid, "settings_nav_paths"), Unique: "fset", Data: "page|3"},
				{Text: "💾 " + tr(uid, "settings_nav_storage"), Unique: "fset", Data: "page|4"},
			},
		}
	case "1a":
		btns = [][]tele.InlineButton{
			{
				{Text: "◀️ " + tr(uid, "settings_back"), Unique: "fset", Data: "page|1"},
			},
			{
				{Text: toggleBtn("RuTor", s.EnableRutorSearch), Unique: "fset", Data: "rutor|1a"},
				{Text: toggleBtn("Torznab", s.EnableTorznabSearch), Unique: "fset", Data: "torznab|1a"},
			},
		}
	case "1b":
		btns = [][]tele.InlineButton{
			{
				{Text: "◀️ " + tr(uid, "settings_back"), Unique: "fset", Data: "page|1"},
			},
			{
				{Text: toggleBtn("DLNA", s.EnableDLNA), Unique: "fset", Data: "dlna|1b"},
				{Text: toggleBtn("IPv6", s.EnableIPv6), Unique: "fset", Data: "ipv6|1b"},
				{Text: toggleBtn("Upload", !s.DisableUpload), Unique: "fset", Data: "upload|1b"},
			},
			{
				{Text: toggleBtn("DHT", !s.DisableDHT), Unique: "fset", Data: "dht|1b"},
				{Text: toggleBtn("PEX", !s.DisablePEX), Unique: "fset", Data: "pex|1b"},
				{Text: toggleBtn("TCP", !s.DisableTCP), Unique: "fset", Data: "tcp|1b"},
				{Text: toggleBtn("UTP", !s.DisableUTP), Unique: "fset", Data: "utp|1b"},
			},
			{
				{Text: toggleBtn("UPNP", !s.DisableUPNP), Unique: "fset", Data: "upnp|1b"},
				{Text: toggleBtn("Encrypt", s.ForceEncrypt), Unique: "fset", Data: "encrypt|1b"},
				{Text: toggleBtn("Debug", s.EnableDebug), Unique: "fset", Data: "debug|1b"},
			},
		}
	case "1c":
		btns = [][]tele.InlineButton{
			{
				{Text: "◀️ " + tr(uid, "settings_back"), Unique: "fset", Data: "page|1"},
			},
			{
				{Text: toggleBtn("CacheDrop", s.RemoveCacheOnDrop), Unique: "fset", Data: "cachedrop|1c"},
				{Text: toggleBtn("Responsive", s.ResponsiveMode), Unique: "fset", Data: "responsive|1c"},
				{Text: toggleBtn("Proxy", s.EnableProxy), Unique: "fset", Data: "proxy|1c"},
			},
			{
				{Text: toggleBtn("UseDisk", s.UseDisk), Unique: "fset", Data: "usedisk|1c"},
				{Text: toggleBtn("FSActive", s.ShowFSActiveTorr), Unique: "fset", Data: "fsactive|1c"},
			},
		}
	case "2":
		btns = [][]tele.InlineButton{
			{
				{Text: "💾 " + tr(uid, "settings_limits_cache"), Unique: "fset", Data: "page|2a"},
				{Text: "🔌 " + tr(uid, "settings_limits_connections"), Unique: "fset", Data: "page|2b"},
				{Text: "⬇️ " + tr(uid, "settings_limits_speed"), Unique: "fset", Data: "page|2c"},
			},
			{
				{Text: "◀️ " + tr(uid, "settings_back"), Unique: "fset", Data: "page|1"},
				{Text: "✏️ " + tr(uid, "settings_nav_paths"), Unique: "fset", Data: "page|3"},
				{Text: "💾 " + tr(uid, "settings_nav_storage"), Unique: "fset", Data: "page|4"},
			},
		}
	case "2a":
		cacheMB := int(s.CacheSize / (1024 * 1024))
		btns = [][]tele.InlineButton{
			{
				{Text: "◀️ " + tr(uid, "settings_back"), Unique: "fset", Data: "page|2"},
			},
			{
				{Text: "💾 " + optBtn("64", cacheMB == 64), Unique: "fset", Data: "cache|64|2a"},
				{Text: optBtn("128", cacheMB == 128), Unique: "fset", Data: "cache|128|2a"},
				{Text: optBtn("256", cacheMB == 256), Unique: "fset", Data: "cache|256|2a"},
				{Text: optBtn("512", cacheMB == 512), Unique: "fset", Data: "cache|512|2a"},
			},
			{
				{Text: "📥 " + optBtn("25%", s.PreloadCache == 25), Unique: "fset", Data: "preload|25|2a"},
				{Text: optBtn("50%", s.PreloadCache == 50), Unique: "fset", Data: "preload|50|2a"},
				{Text: optBtn("75%", s.PreloadCache == 75), Unique: "fset", Data: "preload|75|2a"},
				{Text: optBtn("95%", s.PreloadCache == 95), Unique: "fset", Data: "preload|95|2a"},
			},
			{
				{Text: "📖 " + optBtn("50%", s.ReaderReadAHead == 50), Unique: "fset", Data: "readahead|50|2a"},
				{Text: optBtn("75%", s.ReaderReadAHead == 75), Unique: "fset", Data: "readahead|75|2a"},
				{Text: optBtn("95%", s.ReaderReadAHead == 95), Unique: "fset", Data: "readahead|95|2a"},
				{Text: optBtn("100%", s.ReaderReadAHead == 100), Unique: "fset", Data: "readahead|100|2a"},
			},
		}
	case "2b":
		btns = [][]tele.InlineButton{
			{
				{Text: "◀️ " + tr(uid, "settings_back"), Unique: "fset", Data: "page|2"},
			},
			{
				{Text: "🔌 " + optBtn("25", s.ConnectionsLimit == 25), Unique: "fset", Data: "conn|25|2b"},
				{Text: optBtn("50", s.ConnectionsLimit == 50), Unique: "fset", Data: "conn|50|2b"},
				{Text: optBtn("100", s.ConnectionsLimit == 100), Unique: "fset", Data: "conn|100|2b"},
			},
			{
				{Text: "⏱ " + optBtn("15s", s.TorrentDisconnectTimeout == 15), Unique: "fset", Data: "timeout|15|2b"},
				{Text: optBtn("30s", s.TorrentDisconnectTimeout == 30), Unique: "fset", Data: "timeout|30|2b"},
				{Text: optBtn("60s", s.TorrentDisconnectTimeout == 60), Unique: "fset", Data: "timeout|60|2b"},
				{Text: optBtn("120s", s.TorrentDisconnectTimeout == 120), Unique: "fset", Data: "timeout|120|2b"},
			},
			{
				{Text: "🔌 " + optBtn("auto", s.PeersListenPort == 0), Unique: "fset", Data: "port|0|2b"},
				{Text: optBtn("6881", s.PeersListenPort == 6881), Unique: "fset", Data: "port|6881|2b"},
				{Text: optBtn("51413", s.PeersListenPort == 51413), Unique: "fset", Data: "port|51413|2b"},
			},
		}
	case "2c":
		btns = [][]tele.InlineButton{
			{
				{Text: "◀️ " + tr(uid, "settings_back"), Unique: "fset", Data: "page|2"},
			},
			{
				{Text: "⬇️ " + optBtn("∞", s.DownloadRateLimit == 0), Unique: "fset", Data: "down|0|2c"},
				{Text: optBtn("1M", s.DownloadRateLimit == 1024), Unique: "fset", Data: "down|1024|2c"},
				{Text: optBtn("5M", s.DownloadRateLimit == 5120), Unique: "fset", Data: "down|5120|2c"},
				{Text: optBtn("10M", s.DownloadRateLimit == 10240), Unique: "fset", Data: "down|10240|2c"},
			},
			{
				{Text: "⬆️ " + optBtn("∞", s.UploadRateLimit == 0), Unique: "fset", Data: "up|0|2c"},
				{Text: optBtn("1M", s.UploadRateLimit == 1024), Unique: "fset", Data: "up|1024|2c"},
				{Text: optBtn("5M", s.UploadRateLimit == 5120), Unique: "fset", Data: "up|5120|2c"},
				{Text: optBtn("10M", s.UploadRateLimit == 10240), Unique: "fset", Data: "up|10240|2c"},
			},
			{
				{Text: "🔄 " + optBtn("off", s.RetrackersMode == 0), Unique: "fset", Data: "retr|0|2c"},
				{Text: optBtn("add", s.RetrackersMode == 1), Unique: "fset", Data: "retr|1|2c"},
				{Text: optBtn("rem", s.RetrackersMode == 2), Unique: "fset", Data: "retr|2|2c"},
				{Text: optBtn("repl", s.RetrackersMode == 3), Unique: "fset", Data: "retr|3|2c"},
			},
		}
	case "3":
		btns = [][]tele.InlineButton{
			{
				{Text: "◀️ " + tr(uid, "settings_back"), Unique: "fset", Data: "page|1"},
				{Text: "📊 " + tr(uid, "settings_nav_cache"), Unique: "fset", Data: "page|2"},
				{Text: "💾 " + tr(uid, "settings_nav_storage"), Unique: "fset", Data: "page|4"},
			},
			{
				{Text: "✏️ " + tr(uid, "settings_set_friendlyname"), Unique: "fset", Data: "ask|friendlyname"},
			},
			{
				{Text: "✏️ " + tr(uid, "settings_set_path"), Unique: "fset", Data: "ask|torrentssavepath"},
			},
			{
				{Text: "🔐 " + tr(uid, "settings_set_sslcert"), Unique: "fset", Data: "ask|sslcert"},
				{Text: "🔑 " + tr(uid, "settings_set_sslkey"), Unique: "fset", Data: "ask|sslkey"},
			},
			{
				{Text: "🎬 " + tr(uid, "settings_set_tmdbkey"), Unique: "fset", Data: "ask|tmdbkey"},
			},
			{
				{Text: "🔍 " + tr(uid, "settings_torznab_test"), Unique: "fset", Data: "ask|torznab_test"},
				{Text: "➕ " + tr(uid, "settings_add_torznab"), Unique: "fset", Data: "ask|torznab_add"},
				{Text: "🗑 " + tr(uid, "settings_clear_torznab"), Unique: "fset", Data: "torznab_clear"},
			},
			{
				{Text: "✏️ " + tr(uid, "settings_set_proxyhosts"), Unique: "fset", Data: "ask|proxyhosts"},
			},
		}
	case "4":
		btns = [][]tele.InlineButton{
			{
				{Text: "◀️ " + tr(uid, "settings_back"), Unique: "fset", Data: "page|1"},
				{Text: "📊 " + tr(uid, "settings_nav_cache"), Unique: "fset", Data: "page|2"},
				{Text: "✏️ " + tr(uid, "settings_nav_paths"), Unique: "fset", Data: "page|3"},
			},
			{
				{Text: "📄 " + optBtn("json", s.StoreSettingsInJson), Unique: "fset", Data: "storage_set|json"},
				{Text: optBtn("bbolt", !s.StoreSettingsInJson), Unique: "fset", Data: "storage_set|bbolt"},
			},
			{
				{Text: "📺 " + optBtn("json", s.StoreViewedInJson), Unique: "fset", Data: "storage_view|json"},
				{Text: optBtn("bbolt", !s.StoreViewedInJson), Unique: "fset", Data: "storage_view|bbolt"},
			},
			{
				{Text: "🔄 " + tr(uid, "settings_reset"), Unique: "fset", Data: "reset_confirm"},
			},
		}
	}
	return &tele.ReplyMarkup{InlineKeyboard: btns}
}

func boolIcon(v bool) string {
	if v {
		return "✅"
	}
	return "❌"
}

func toggleBtn(label string, on bool) string {
	if on {
		return label + " ✅"
	}
	return label + " ❌"
}

func optBtn(label string, isCurrent bool) string {
	if isCurrent {
		return label + " ✓"
	}
	return label
}

func settingsCallback(c tele.Context, action string) error {
	uid := c.Sender().ID
	if !isAdmin(uid) {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "admin_only")})
	}
	if settings.BTsets == nil {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "settings_not_loaded")})
	}

	if action == "export" {
		buf, err := json.MarshalIndent(settings.BTsets, "", "  ")
		if err != nil {
			return c.Respond(&tele.CallbackResponse{Text: fmt.Sprintf(tr(uid, "settings_error"), err.Error())})
		}
		doc := &tele.Document{}
		doc.FileName = "torrserver_settings.json"
		doc.FileReader = bytes.NewReader(buf)
		doc.Caption = "⚙️ " + tr(uid, "settings_export_caption")
		_ = c.Send(doc)
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "settings_exported")})
	}

	if action == "input_cancel" {
		return cancelSettingsInput(c)
	}

	if action == "reset_confirm" {
		btnYes := tele.InlineButton{Text: tr(uid, "btn_yes"), Unique: "fset", Data: "reset_def|1"}
		btnNo := tele.InlineButton{Text: tr(uid, "btn_no"), Unique: "fset", Data: "reset_def|0"}
		kbd := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{btnYes, btnNo}}}
		msg := sendSettingsMenuText(c, uid, "4") + "\n\n⚠️ " + tr(uid, "settings_reset_confirm")
		if _, err := c.Bot().Edit(c.Callback().Message, msg, kbd, tele.ModeHTML); err != nil {
			_ = c.Send(tr(uid, "settings_reset_confirm"), kbd)
		}
		return c.Respond(&tele.CallbackResponse{})
	}

	if len(action) > 9 && action[:9] == "reset_def|" {
		if action[9:] != "1" {
			msg := sendSettingsMenuText(c, uid, "4")
			kbd := sendSettingsMenuKbd(uid, "4")
			if _, err := c.Bot().Edit(c.Callback().Message, msg, kbd, tele.ModeHTML); err != nil {
				_ = sendSettingsMenuPage(c, uid, "4")
			}
			return c.Respond(&tele.CallbackResponse{Text: tr(uid, "canceled")})
		}
		if settings.ReadOnly {
			return c.Respond(&tele.CallbackResponse{Text: tr(uid, "settings_readonly")})
		}
		torr.SetDefSettings()
		dlna.Stop()
		rutor.Stop()
		rutor.Start()
		msg := sendSettingsMenuText(c, uid, "4")
		kbd := sendSettingsMenuKbd(uid, "4")
		if _, err := c.Bot().Edit(c.Callback().Message, msg, kbd, tele.ModeHTML); err != nil {
			_ = sendSettingsMenuPage(c, uid, "4")
		}
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "settings_reset_done")})
	}

	if len(action) > 12 && action[:12] == "storage_set|" {
		if settings.ReadOnly {
			return c.Respond(&tele.CallbackResponse{Text: tr(uid, "settings_readonly")})
		}
		val := action[12:]
		prefs := map[string]interface{}{"settings": val}
		if err := settings.SetStoragePreferences(prefs); err != nil {
			return c.Respond(&tele.CallbackResponse{Text: fmt.Sprintf(tr(uid, "settings_error"), err.Error())})
		}
		page := "4"
		msg := sendSettingsMenuText(c, uid, page)
		kbd := sendSettingsMenuKbd(uid, page)
		if _, err := c.Bot().Edit(c.Callback().Message, msg, kbd, tele.ModeHTML); err != nil {
			_ = sendSettingsMenuPage(c, uid, page)
		}
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "settings_saved")})
	}

	if len(action) > 12 && action[:12] == "storage_view|" {
		if settings.ReadOnly {
			return c.Respond(&tele.CallbackResponse{Text: tr(uid, "settings_readonly")})
		}
		val := action[12:]
		prefs := map[string]interface{}{"viewed": val}
		if err := settings.SetStoragePreferences(prefs); err != nil {
			return c.Respond(&tele.CallbackResponse{Text: fmt.Sprintf(tr(uid, "settings_error"), err.Error())})
		}
		page := "4"
		msg := sendSettingsMenuText(c, uid, page)
		kbd := sendSettingsMenuKbd(uid, page)
		if _, err := c.Bot().Edit(c.Callback().Message, msg, kbd, tele.ModeHTML); err != nil {
			_ = sendSettingsMenuPage(c, uid, page)
		}
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "settings_saved")})
	}

	if len(action) > 4 && action[:4] == "ask|" {
		setting := action[4:]
		var hint string
		switch setting {
		case "friendlyname":
			hint = tr(uid, "settings_hint_friendlyname")
		case "torrentssavepath":
			hint = tr(uid, "settings_hint_path")
		case "sslcert":
			hint = tr(uid, "settings_hint_sslcert")
		case "sslkey":
			hint = tr(uid, "settings_hint_sslkey")
		case "tmdbkey":
			hint = tr(uid, "settings_hint_tmdbkey")
		case "proxyhosts":
			hint = tr(uid, "settings_hint_proxyhosts")
		case "torznab_add":
			hint = tr(uid, "settings_hint_torznab")
		case "torznab_test":
			hint = tr(uid, "settings_hint_torznab_test")
		default:
			return c.Respond(&tele.CallbackResponse{Text: tr(uid, "callback_unknown")})
		}
		return sendSettingsInputPrompt(c, uid, setting, hint)
	}

	if len(action) > 5 && action[:5] == "page|" {
		page := action[5:]
		msg := sendSettingsMenuText(c, uid, page)
		kbd := sendSettingsMenuKbd(uid, page)
		if _, err := c.Bot().Edit(c.Callback().Message, msg, kbd, tele.ModeHTML); err != nil {
			_ = sendSettingsMenuPage(c, uid, page)
		}
		return c.Respond(&tele.CallbackResponse{})
	}

	if settings.ReadOnly {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "settings_readonly")})
	}

	sets := new(settings.BTSets)
	*sets = *settings.BTsets
	page := "1"

	// Extract return page from action (e.g. "rutor|1a" -> action "rutor", page "1a")
	if idx := strings.Index(action, "|"); idx >= 0 {
		suffix := action[idx+1:]
		if suffix == "1a" || suffix == "1b" || suffix == "1c" {
			page = suffix
			action = action[:idx]
		}
	}

	switch action {
	case "rutor":
		sets.EnableRutorSearch = !sets.EnableRutorSearch
	case "torznab":
		sets.EnableTorznabSearch = !sets.EnableTorznabSearch
	case "dlna":
		sets.EnableDLNA = !sets.EnableDLNA
	case "ipv6":
		sets.EnableIPv6 = !sets.EnableIPv6
	case "upload":
		sets.DisableUpload = !sets.DisableUpload
	case "dht":
		sets.DisableDHT = !sets.DisableDHT
	case "pex":
		sets.DisablePEX = !sets.DisablePEX
	case "tcp":
		sets.DisableTCP = !sets.DisableTCP
	case "utp":
		sets.DisableUTP = !sets.DisableUTP
	case "upnp":
		sets.DisableUPNP = !sets.DisableUPNP
	case "encrypt":
		sets.ForceEncrypt = !sets.ForceEncrypt
	case "debug":
		sets.EnableDebug = !sets.EnableDebug
	case "cachedrop":
		sets.RemoveCacheOnDrop = !sets.RemoveCacheOnDrop
	case "responsive":
		sets.ResponsiveMode = !sets.ResponsiveMode
	case "proxy":
		sets.EnableProxy = !sets.EnableProxy
	case "usedisk":
		sets.UseDisk = !sets.UseDisk
	case "fsactive":
		sets.ShowFSActiveTorr = !sets.ShowFSActiveTorr
	case "storejson":
		sets.StoreSettingsInJson = !sets.StoreSettingsInJson
	case "viewedjson":
		sets.StoreViewedInJson = !sets.StoreViewedInJson
	case "torznab_clear":
		if settings.ReadOnly {
			return c.Respond(&tele.CallbackResponse{Text: tr(uid, "settings_readonly")})
		}
		sets.TorznabUrls = nil
		page = "3"
		torr.SetSettings(sets)
		rutor.Stop()
		rutor.Start()
		msg := sendSettingsMenuText(c, uid, page)
		kbd := sendSettingsMenuKbd(uid, page)
		if _, err := c.Bot().Edit(c.Callback().Message, msg, kbd, tele.ModeHTML); err != nil {
			_ = sendSettingsMenuPage(c, uid, page)
		}
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "settings_saved")})
	default:
		if parts := splitAction(action); len(parts) == 2 {
			key, value := parts[0], parts[1]
			page = "2"
			if idx := strings.Index(value, "|"); idx >= 0 {
				if ret := value[idx+1:]; ret == "2a" || ret == "2b" || ret == "2c" {
					page = ret
				}
				value = value[:idx]
			}
			switch key {
			case "cache":
				if v := parseInt(value); v > 0 {
					sets.CacheSize = int64(v) * 1024 * 1024
				}
			case "preload":
				if v := parseInt(value); v >= 0 && v <= 100 {
					sets.PreloadCache = v
				}
			case "readahead":
				if v := parseInt(value); v >= 5 && v <= 100 {
					sets.ReaderReadAHead = v
				}
			case "conn":
				if v := parseInt(value); v > 0 {
					sets.ConnectionsLimit = v
				}
			case "timeout":
				if v := parseInt(value); v > 0 {
					sets.TorrentDisconnectTimeout = v
				}
			case "port":
				v := parseInt(value)
				if v >= 0 && (v == 0 || (v >= 1024 && v <= 65535)) {
					sets.PeersListenPort = v
				}
			case "down":
				sets.DownloadRateLimit = parseInt(value)
			case "up":
				sets.UploadRateLimit = parseInt(value)
			case "retr":
				if v := parseInt(value); v >= 0 && v <= 3 {
					sets.RetrackersMode = v
				}
			default:
				return c.Respond(&tele.CallbackResponse{Text: tr(uid, "callback_unknown")})
			}
		} else {
			return c.Respond(&tele.CallbackResponse{Text: tr(uid, "callback_unknown")})
		}
	}

	torr.SetSettings(sets)
	dlna.Stop()
	if sets.EnableDLNA {
		dlna.Start()
	}
	rutor.Stop()
	rutor.Start()

	msg := sendSettingsMenuText(c, uid, page)
	kbd := sendSettingsMenuKbd(uid, page)
	if _, err := c.Bot().Edit(c.Callback().Message, msg, kbd, tele.ModeHTML); err != nil {
		_ = sendSettingsMenuPage(c, uid, page)
	}
	return c.Respond(&tele.CallbackResponse{Text: tr(uid, "settings_saved")})
}

func splitAction(action string) []string {
	for i := 0; i < len(action); i++ {
		if action[i] == '|' {
			return []string{action[:i], action[i+1:]}
		}
	}
	return nil
}

func parseInt(s string) int {
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}
