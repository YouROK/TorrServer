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
	"server/tgbot/config"
	"server/torr"
)

func isAdmin(userID int64) bool {
	if len(config.Cfg.WhiteIds) == 0 {
		return false
	}
	for _, id := range config.Cfg.WhiteIds {
		if id == userID {
			return true
		}
	}
	return false
}

func cmdShutdown(c tele.Context) error {
	uid := c.Sender().ID
	btnYes := tele.InlineButton{Text: tr(uid, "btn_yes"), Unique: "fshutdown", Data: "1"}
	btnNo := tele.InlineButton{Text: tr(uid, "btn_no"), Unique: "fshutdown", Data: "0"}
	kbd := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{btnYes, btnNo}}}
	return c.Send(tr(uid, "shutdown_confirm"), kbd)
}

func shutdownConfirm(c tele.Context, confirm string) error {
	uid := c.Sender().ID
	if !isAdmin(uid) {
		return c.Respond(&tele.CallbackResponse{Text: tr(uid, "admin_only")})
	}
	if confirm != "1" {
		_ = c.Respond(&tele.CallbackResponse{Text: tr(uid, "canceled")})
		return c.Bot().Delete(c.Callback().Message)
	}
	_ = c.Respond(&tele.CallbackResponse{Text: tr(c.Sender().ID, "server_stopped")})
	_ = c.Bot().Delete(c.Callback().Message)
	_ = c.Send(tr(c.Sender().ID, "server_stopped"))
	go func() {
		torr.Shutdown()
	}()
	return nil
}

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
	case "2":
		msg += " — " + tr(uid, "settings_page2")
	case "3":
		msg += " — " + tr(uid, "settings_page3")
	}
	msg += "\n\n"

	switch page {
	case "1":
		msg += "<b>" + tr(uid, "settings_section_search") + "</b>\n"
		msg += fmt.Sprintf("🔍 RuTor %s | Torznab %s\n\n", boolIcon(s.EnableRutorSearch), boolIcon(s.EnableTorznabSearch))
		msg += "<b>" + tr(uid, "settings_section_network") + "</b>\n"
		msg += fmt.Sprintf("📺 DLNA %s | IPv6 %s\n", boolIcon(s.EnableDLNA), boolIcon(s.EnableIPv6))
		msg += fmt.Sprintf("⬇️ Upload %s | DHT %s | PEX %s\n", boolIcon(!s.DisableUpload), boolIcon(!s.DisableDHT), boolIcon(!s.DisablePEX))
		msg += fmt.Sprintf("🔗 TCP %s | UTP %s | UPNP %s\n", boolIcon(!s.DisableTCP), boolIcon(!s.DisableUTP), boolIcon(!s.DisableUPNP))
		msg += fmt.Sprintf("🔒 Encrypt %s | Debug %s\n\n", boolIcon(s.ForceEncrypt), boolIcon(s.EnableDebug))
		msg += "<b>" + tr(uid, "settings_section_other") + "</b>\n"
		msg += fmt.Sprintf("📦 CacheDrop %s | Responsive %s | Proxy %s\n", boolIcon(s.RemoveCacheOnDrop), boolIcon(s.ResponsiveMode), boolIcon(s.EnableProxy))
		msg += fmt.Sprintf("💾 UseDisk %s | FSActive %s\n", boolIcon(s.UseDisk), boolIcon(s.ShowFSActiveTorr))
		msg += fmt.Sprintf("📄 StoreJSON %s | ViewedJSON %s\n", boolIcon(s.StoreSettingsInJson), boolIcon(s.StoreViewedInJson))
	case "2":
		msg += "<b>" + tr(uid, "settings_section_limits") + "</b>\n\n"
		msg += fmt.Sprintf("💾 <b>Cache:</b> %d MB | <b>Preload:</b> %d%% | <b>ReadAhead:</b> %d%%\n\n", s.CacheSize/(1024*1024), s.PreloadCache, s.ReaderReadAHead)
		msg += fmt.Sprintf("🔌 <b>Connections:</b> %d | <b>Port:</b> %s | <b>Timeout:</b> %ds\n\n", s.ConnectionsLimit, portStr(s.PeersListenPort), s.TorrentDisconnectTimeout)
		msg += fmt.Sprintf("⬇️ <b>Down:</b> %s | ⬆️ <b>Up:</b> %s KB/s\n\n", rateStr(s.DownloadRateLimit), rateStr(s.UploadRateLimit))
		msg += fmt.Sprintf("🔄 <b>Retrackers:</b> %s\n", retrackersStr(s.RetrackersMode))
	case "3":
		msg += "<b>" + tr(uid, "settings_section_paths") + "</b>\n\n"
		msg += fmt.Sprintf("📺 <b>DLNA:</b> %s\n", maskStr(s.FriendlyName, 30))
		msg += fmt.Sprintf("💾 <b>Path:</b> %s\n", maskVal(s.TorrentsSavePath))
		msg += fmt.Sprintf("🔐 <b>SSL:</b> %s\n", maskVal(s.SslCert))
		msg += fmt.Sprintf("🔑 <b>TMDB:</b> %s | <b>Torznab:</b> %d\n", maskVal(s.TMDBSettings.APIKey), len(s.TorznabUrls))
		msg += fmt.Sprintf("🌐 <b>Proxy:</b> %s\n", maskStr(strings.Join(s.ProxyHosts, ", "), 40))
	}
	return msg
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
				{Text: "📥 " + tr(uid, "settings_export"), Unique: "fset", Data: "export"},
				{Text: "▶️ " + tr(uid, "settings_more"), Unique: "fset", Data: "page|2"},
				{Text: "✏️ " + tr(uid, "settings_page3"), Unique: "fset", Data: "page|3"},
			},
			{
				{Text: toggleBtn("RuTor", s.EnableRutorSearch), Unique: "fset", Data: "rutor"},
				{Text: toggleBtn("Torznab", s.EnableTorznabSearch), Unique: "fset", Data: "torznab"},
			},
			{
				{Text: toggleBtn("DLNA", s.EnableDLNA), Unique: "fset", Data: "dlna"},
				{Text: toggleBtn("IPv6", s.EnableIPv6), Unique: "fset", Data: "ipv6"},
			},
			{
				{Text: toggleBtn("Upload", !s.DisableUpload), Unique: "fset", Data: "upload"},
				{Text: toggleBtn("DHT", !s.DisableDHT), Unique: "fset", Data: "dht"},
				{Text: toggleBtn("PEX", !s.DisablePEX), Unique: "fset", Data: "pex"},
			},
			{
				{Text: toggleBtn("TCP", !s.DisableTCP), Unique: "fset", Data: "tcp"},
				{Text: toggleBtn("UTP", !s.DisableUTP), Unique: "fset", Data: "utp"},
				{Text: toggleBtn("UPNP", !s.DisableUPNP), Unique: "fset", Data: "upnp"},
			},
			{
				{Text: toggleBtn("Encrypt", s.ForceEncrypt), Unique: "fset", Data: "encrypt"},
				{Text: toggleBtn("Debug", s.EnableDebug), Unique: "fset", Data: "debug"},
			},
			{
				{Text: toggleBtn("CacheDrop", s.RemoveCacheOnDrop), Unique: "fset", Data: "cachedrop"},
				{Text: toggleBtn("Responsive", s.ResponsiveMode), Unique: "fset", Data: "responsive"},
				{Text: toggleBtn("Proxy", s.EnableProxy), Unique: "fset", Data: "proxy"},
			},
			{
				{Text: toggleBtn("UseDisk", s.UseDisk), Unique: "fset", Data: "usedisk"},
				{Text: toggleBtn("FSActive", s.ShowFSActiveTorr), Unique: "fset", Data: "fsactive"},
			},
			{
				{Text: toggleBtn("StoreJSON", s.StoreSettingsInJson), Unique: "fset", Data: "storejson"},
				{Text: toggleBtn("ViewedJSON", s.StoreViewedInJson), Unique: "fset", Data: "viewedjson"},
			},
		}
	case "2":
		cacheMB := int(s.CacheSize / (1024 * 1024))
		btns = [][]tele.InlineButton{
			{
				{Text: "◀️ " + tr(uid, "settings_back"), Unique: "fset", Data: "page|1"},
				{Text: "▶️ " + tr(uid, "settings_page3"), Unique: "fset", Data: "page|3"},
			},
			{
				{Text: "💾 " + optBtn("64", cacheMB == 64), Unique: "fset", Data: "cache|64"},
				{Text: optBtn("128", cacheMB == 128), Unique: "fset", Data: "cache|128"},
				{Text: optBtn("256", cacheMB == 256), Unique: "fset", Data: "cache|256"},
				{Text: optBtn("512 MB", cacheMB == 512), Unique: "fset", Data: "cache|512"},
			},
			{
				{Text: "📥 " + optBtn("25%", s.PreloadCache == 25), Unique: "fset", Data: "preload|25"},
				{Text: optBtn("50%", s.PreloadCache == 50), Unique: "fset", Data: "preload|50"},
				{Text: optBtn("75%", s.PreloadCache == 75), Unique: "fset", Data: "preload|75"},
				{Text: optBtn("95%", s.PreloadCache == 95), Unique: "fset", Data: "preload|95"},
			},
			{
				{Text: "📖 " + optBtn("50%", s.ReaderReadAHead == 50), Unique: "fset", Data: "readahead|50"},
				{Text: optBtn("75%", s.ReaderReadAHead == 75), Unique: "fset", Data: "readahead|75"},
				{Text: optBtn("95%", s.ReaderReadAHead == 95), Unique: "fset", Data: "readahead|95"},
				{Text: optBtn("100%", s.ReaderReadAHead == 100), Unique: "fset", Data: "readahead|100"},
			},
			{
				{Text: "🔌 " + optBtn("25", s.ConnectionsLimit == 25), Unique: "fset", Data: "conn|25"},
				{Text: optBtn("50", s.ConnectionsLimit == 50), Unique: "fset", Data: "conn|50"},
				{Text: optBtn("100", s.ConnectionsLimit == 100), Unique: "fset", Data: "conn|100"},
			},
			{
				{Text: "⏱ " + optBtn("15s", s.TorrentDisconnectTimeout == 15), Unique: "fset", Data: "timeout|15"},
				{Text: optBtn("30s", s.TorrentDisconnectTimeout == 30), Unique: "fset", Data: "timeout|30"},
				{Text: optBtn("60s", s.TorrentDisconnectTimeout == 60), Unique: "fset", Data: "timeout|60"},
				{Text: optBtn("120s", s.TorrentDisconnectTimeout == 120), Unique: "fset", Data: "timeout|120"},
			},
			{
				{Text: "🔌 " + optBtn("auto", s.PeersListenPort == 0), Unique: "fset", Data: "port|0"},
				{Text: optBtn("6881", s.PeersListenPort == 6881), Unique: "fset", Data: "port|6881"},
				{Text: optBtn("51413", s.PeersListenPort == 51413), Unique: "fset", Data: "port|51413"},
			},
			{
				{Text: "⬇️ " + optBtn("∞", s.DownloadRateLimit == 0), Unique: "fset", Data: "down|0"},
				{Text: optBtn("1M", s.DownloadRateLimit == 1024), Unique: "fset", Data: "down|1024"},
				{Text: optBtn("5M", s.DownloadRateLimit == 5120), Unique: "fset", Data: "down|5120"},
				{Text: optBtn("10M", s.DownloadRateLimit == 10240), Unique: "fset", Data: "down|10240"},
			},
			{
				{Text: "⬆️ " + optBtn("∞", s.UploadRateLimit == 0), Unique: "fset", Data: "up|0"},
				{Text: optBtn("1M", s.UploadRateLimit == 1024), Unique: "fset", Data: "up|1024"},
				{Text: optBtn("5M", s.UploadRateLimit == 5120), Unique: "fset", Data: "up|5120"},
				{Text: optBtn("10M", s.UploadRateLimit == 10240), Unique: "fset", Data: "up|10240"},
			},
			{
				{Text: "🔄 " + optBtn("off", s.RetrackersMode == 0), Unique: "fset", Data: "retr|0"},
				{Text: optBtn("add", s.RetrackersMode == 1), Unique: "fset", Data: "retr|1"},
				{Text: optBtn("rem", s.RetrackersMode == 2), Unique: "fset", Data: "retr|2"},
				{Text: optBtn("repl", s.RetrackersMode == 3), Unique: "fset", Data: "retr|3"},
			},
		}
	case "3":
		btns = [][]tele.InlineButton{
			{
				{Text: "◀️ " + tr(uid, "settings_back"), Unique: "fset", Data: "page|1"},
				{Text: "📊 " + tr(uid, "settings_to_page2"), Unique: "fset", Data: "page|2"},
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
				{Text: "➕ " + tr(uid, "settings_add_torznab"), Unique: "fset", Data: "ask|torznab_add"},
				{Text: "🗑 " + tr(uid, "settings_clear_torznab"), Unique: "fset", Data: "torznab_clear"},
			},
			{
				{Text: "✏️ " + tr(uid, "settings_set_proxyhosts"), Unique: "fset", Data: "ask|proxyhosts"},
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
		masked := maskSensitiveSettings(settings.BTsets)
		buf, err := json.MarshalIndent(masked, "", "  ")
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
			page = "2"
			switch parts[0] {
			case "cache":
				if v := parseInt(parts[1]); v > 0 {
					sets.CacheSize = int64(v) * 1024 * 1024
				}
			case "preload":
				if v := parseInt(parts[1]); v >= 0 && v <= 100 {
					sets.PreloadCache = v
				}
			case "readahead":
				if v := parseInt(parts[1]); v >= 5 && v <= 100 {
					sets.ReaderReadAHead = v
				}
			case "conn":
				if v := parseInt(parts[1]); v > 0 {
					sets.ConnectionsLimit = v
				}
			case "timeout":
				if v := parseInt(parts[1]); v > 0 {
					sets.TorrentDisconnectTimeout = v
				}
			case "port":
				v := parseInt(parts[1])
				if v >= 0 && (v == 0 || (v >= 1024 && v <= 65535)) {
					sets.PeersListenPort = v
				}
			case "down":
				sets.DownloadRateLimit = parseInt(parts[1])
			case "up":
				sets.UploadRateLimit = parseInt(parts[1])
			case "retr":
				if v := parseInt(parts[1]); v >= 0 && v <= 3 {
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

func maskSensitiveSettings(sets *settings.BTSets) map[string]interface{} {
	buf, _ := json.Marshal(sets)
	var m map[string]interface{}
	_ = json.Unmarshal(buf, &m)
	if m == nil {
		return m
	}
	for _, k := range []string{"SslCert", "SslKey", "TorrentsSavePath"} {
		if v, ok := m[k].(string); ok && v != "" {
			m[k] = "***"
		}
	}
	if t, ok := m["TMDBSettings"].(map[string]interface{}); ok {
		if v, ok := t["APIKey"].(string); ok && v != "" {
			t["APIKey"] = "***"
		}
	}
	if urls, ok := m["TorznabUrls"].([]interface{}); ok {
		for _, cfg := range urls {
			if c, ok := cfg.(map[string]interface{}); ok {
				if v, ok := c["Key"].(string); ok && v != "" {
					c["Key"] = "***"
				}
			}
		}
	}
	return m
}
