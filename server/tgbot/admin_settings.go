package tgbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"server/bonjour"
	"server/dlna"
	"server/rutor"
	"server/settings"
	"server/torr"

	tele "gopkg.in/telebot.v4"
)

func cmdSettings(c tele.Context) error {
	uid := c.Sender().ID
	if settings.BTsets == nil {
		return c.Send(tr(uid, "settings_not_loaded"))
	}
	return showSettings(c, uid, "1", c.Callback() != nil)
}

func sendSettingsMenuPage(c tele.Context, uid int64, page string) error {
	return showSettings(c, uid, page, false)
}

func showSettings(c tele.Context, uid int64, page string, edit bool) error {
	msg := sendSettingsMenuText(c, uid, page)
	kbd := sendSettingsMenuKbd(uid, page)
	if edit && c.Callback() != nil && c.Callback().Message != nil {
		_, err := c.Bot().Edit(c.Callback().Message, msg, kbd, tele.ModeHTML)
		if err == nil {
			return nil
		}
	}
	return c.Send(msg, kbd, tele.ModeHTML)
}

func sendSettingsMenuText(c tele.Context, uid int64, page string) string {
	s := settings.BTsets
	if s == nil {
		s = &settings.BTSets{}
	}
	var b strings.Builder
	b.WriteString("⚙️ <b>" + tr(uid, "settings_title") + "</b>")
	switch page {
	case "1":
		b.WriteString(sectionHead(tr(uid, "settings_section_search")))
		b.WriteString(boolKV(tr(uid, "settings_tgl_rutor"), s.EnableRutorSearch))
		b.WriteString(boolKV(tr(uid, "settings_tgl_torznab"), s.EnableTorznabSearch))
		b.WriteString(sectionHead(tr(uid, "settings_section_network")))
		b.WriteString(boolKV(tr(uid, "settings_tgl_dlna"), s.EnableDLNA))
		b.WriteString(boolKV(tr(uid, "settings_tgl_dht"), !s.DisableDHT))
		b.WriteString(boolKV(tr(uid, "settings_tgl_ipv6"), s.EnableIPv6))
		b.WriteString(sectionHead(tr(uid, "settings_limits_cache")))
		fmt.Fprintf(&b, "%s: <b>%d MB</b>\n", tr(uid, "settings_tgl_cache"), s.CacheSize/(1024*1024))
		fmt.Fprintf(&b, "%s: <b>%d%%</b>\n", tr(uid, "settings_tgl_preload"), s.PreloadCache)
	case "1a":
		b.WriteString(" — " + tr(uid, "settings_section_search") + "\n\n")
		b.WriteString(boolKV(tr(uid, "settings_tgl_rutor"), s.EnableRutorSearch))
		b.WriteString(boolKV(tr(uid, "settings_tgl_torznab"), s.EnableTorznabSearch))
	case "1b":
		b.WriteString(" — " + tr(uid, "settings_section_network") + "\n\n")
		b.WriteString(boolKV(tr(uid, "settings_tgl_dlna"), s.EnableDLNA))
		b.WriteString(boolKV(tr(uid, "settings_tgl_bonjour"), s.EnableBonjour))
		b.WriteString(boolKV(tr(uid, "settings_tgl_ipv6"), s.EnableIPv6))
		b.WriteString(boolKV(tr(uid, "settings_tgl_upload"), !s.DisableUpload))
		b.WriteString(boolKV(tr(uid, "settings_tgl_dht"), !s.DisableDHT))
		b.WriteString(boolKV(tr(uid, "settings_tgl_pex"), !s.DisablePEX))
		b.WriteString(boolKV(tr(uid, "settings_tgl_tcp"), !s.DisableTCP))
		b.WriteString(boolKV(tr(uid, "settings_tgl_utp"), !s.DisableUTP))
		b.WriteString(boolKV(tr(uid, "settings_tgl_upnp"), !s.DisableUPNP))
		b.WriteString(boolKV(tr(uid, "settings_tgl_encrypt"), s.ForceEncrypt))
		b.WriteString(boolKV(tr(uid, "settings_tgl_debug"), s.EnableDebug))
	case "1c":
		b.WriteString(" — " + tr(uid, "settings_section_other") + "\n\n")
		b.WriteString(boolKV(tr(uid, "settings_tgl_cachedrop"), s.RemoveCacheOnDrop))
		b.WriteString(boolKV(tr(uid, "settings_tgl_responsive"), s.ResponsiveMode))
		b.WriteString(boolKV(tr(uid, "settings_tgl_usedisk"), s.UseDisk))
		b.WriteString(boolKV(tr(uid, "settings_tgl_fsactive"), s.ShowFSActiveTorr))
		b.WriteString(boolKV(tr(uid, "settings_tgl_timecode"), s.TrackTimecode))
		b.WriteString(boolKV(tr(uid, "settings_tgl_lpd"), s.EnableLPD))
		b.WriteString(boolKV(tr(uid, "settings_tgl_lpdv6"), s.LPDIPv6))
	case "2":
		b.WriteString(" — " + tr(uid, "settings_page2") + "\n\n")
		fmt.Fprintf(&b, "%s: <b>%d MB</b>\n", tr(uid, "settings_tgl_cache"), s.CacheSize/(1024*1024))
		fmt.Fprintf(&b, "%s: <b>%d%%</b>\n", tr(uid, "settings_tgl_preload"), s.PreloadCache)
		fmt.Fprintf(&b, "%s: <b>%d%%</b>\n", tr(uid, "settings_tgl_readahead"), s.ReaderReadAHead)
		fmt.Fprintf(&b, "%s: <b>%d</b>\n", tr(uid, "settings_tgl_conn"), s.ConnectionsLimit)
		fmt.Fprintf(&b, "%s: <b>%s</b>\n", tr(uid, "settings_tgl_port"), portStr(s.PeersListenPort))
		fmt.Fprintf(&b, "%s: <b>%ds</b>\n", tr(uid, "settings_tgl_timeout"), s.TorrentDisconnectTimeout)
		fmt.Fprintf(&b, "%s: <b>%s</b>\n", tr(uid, "settings_tgl_down"), rateStr(s.DownloadRateLimit))
		fmt.Fprintf(&b, "%s: <b>%s</b>\n", tr(uid, "settings_tgl_up"), rateStr(s.UploadRateLimit))
		fmt.Fprintf(&b, "%s: <b>%s</b>\n", tr(uid, "settings_tgl_retr"), retrackersStr(s.RetrackersMode))
	case "2a":
		b.WriteString(" — " + tr(uid, "settings_limits_cache") + "\n\n")
		fmt.Fprintf(&b, "%s: <b>%d MB</b>\n", tr(uid, "settings_tgl_cache"), s.CacheSize/(1024*1024))
		fmt.Fprintf(&b, "%s: <b>%d%%</b>\n", tr(uid, "settings_tgl_preload"), s.PreloadCache)
		fmt.Fprintf(&b, "%s: <b>%d%%</b>\n", tr(uid, "settings_tgl_readahead"), s.ReaderReadAHead)
	case "2b":
		b.WriteString(" — " + tr(uid, "settings_limits_connections") + "\n\n")
		fmt.Fprintf(&b, "%s: <b>%d</b>\n", tr(uid, "settings_tgl_conn"), s.ConnectionsLimit)
		fmt.Fprintf(&b, "%s: <b>%s</b>\n", tr(uid, "settings_tgl_port"), portStr(s.PeersListenPort))
		fmt.Fprintf(&b, "%s: <b>%ds</b>\n", tr(uid, "settings_tgl_timeout"), s.TorrentDisconnectTimeout)
	case "2c":
		b.WriteString(" — " + tr(uid, "settings_limits_speed") + "\n\n")
		fmt.Fprintf(&b, "%s: <b>%s</b>\n", tr(uid, "settings_tgl_down"), rateStr(s.DownloadRateLimit))
		fmt.Fprintf(&b, "%s: <b>%s</b>\n", tr(uid, "settings_tgl_up"), rateStr(s.UploadRateLimit))
		fmt.Fprintf(&b, "%s: <b>%s</b>\n", tr(uid, "settings_tgl_retr"), retrackersStr(s.RetrackersMode))
	case "3":
		b.WriteString(" — " + tr(uid, "settings_page3") + "\n\n")
		b.WriteString(kvLine(tr(uid, "settings_set_friendlyname"), maskStr(s.FriendlyName, 25)))
		b.WriteString(kvLine(tr(uid, "settings_set_path"), maskVal(s.TorrentsSavePath)))
		b.WriteString(kvLine(tr(uid, "settings_set_sslcert"), maskVal(s.SslCert)))
		b.WriteString(kvLine("TMDB", maskVal(s.TMDBSettings.APIKey)))
		fmt.Fprintf(&b, "Torznab: <b>%d</b>\n", len(s.TorznabUrls))
	case "4":
		b.WriteString(" — " + tr(uid, "settings_page4") + "\n\n")
		b.WriteString(kvLine(tr(uid, "settings_storage_settings"), storageType(s.StoreSettingsInJson)))
		b.WriteString(kvLine(tr(uid, "settings_storage_viewed"), storageType(s.StoreViewedInJson)))
		b.WriteString(kvLine("TMDB", maskVal(s.TMDBSettings.APIKey)))
		b.WriteString(kvLine(tr(uid, "settings_tmdb_image"), maskStr(s.TMDBSettings.ImageURL, 20)))
	}
	return b.String()
}

func sectionHead(title string) string {
	return "\n\n<b>" + title + "</b>\n"
}

func kvLine(label, val string) string {
	return label + ": <b>" + val + "</b>\n"
}

func boolKV(label string, v bool) string {
	return kvLine(label, boolIcon(v))
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

	if action == "more" {
		_ = c.Respond(&tele.CallbackResponse{})
		return showMoreHub(c, "admin", true)
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

	if answer, ok := strings.CutPrefix(action, "reset_def|"); ok {
		if answer != "1" {
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

	if val, ok := strings.CutPrefix(action, "storage_view|"); ok {
		if settings.ReadOnly {
			return c.Respond(&tele.CallbackResponse{Text: tr(uid, "settings_readonly")})
		}
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
	case "bonjour":
		sets.EnableBonjour = !sets.EnableBonjour
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
	case "usedisk":
		sets.UseDisk = !sets.UseDisk
	case "fsactive":
		sets.ShowFSActiveTorr = !sets.ShowFSActiveTorr
	case "timecode":
		sets.TrackTimecode = !sets.TrackTimecode
	case "lpd":
		sets.EnableLPD = !sets.EnableLPD
	case "lpdv6":
		sets.LPDIPv6 = !sets.LPDIPv6
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
	bonjour.Stop()
	if sets.EnableBonjour {
		bonjour.Start()
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
